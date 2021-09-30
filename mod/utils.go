package mod

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/google/renameio"
	"golang.org/x/sys/unix"
)

func debug(v ...interface{}) {
	if os.Getenv("KUTA_DEBUG") == "1" {
		v = append([]interface{}{"[kuta]"}, v...)
		log.Println(v...)
	}
}

func debugf(format string, v ...interface{}) {
	if os.Getenv("KUTA_DEBUG") == "1" {
		log.Printf("[kuta] "+format, v...)
	}
}

func chownR(path string, uid, gid int) error {
	return exec.Command("chown", "-R", fmt.Sprintf("%d:%d", uid, gid), path).Run()
}

func patchEtcPasswd(user string, newUID string, newGID string) error {
	path := "/etc/passwd"
	return patchFile(path, func(line string) (string, error) {
		fs := strings.Split(line, ":")
		if len(fs) != 7 {
			return "", errors.New("unexpected number of fields in /etc/passwd")
		}
		if fs[0] == user && (fs[2] != newUID || fs[3] != newGID) {
			debugf("replacing %s uid %s->%s and gid %s->%s\n", fs[0], fs[2], newUID, fs[3], newGID)
			fs[2] = newUID
			fs[3] = newGID
			return strings.Join(fs, ":"), nil
		}
		return line, nil
	})
}

func patchEtcGroup(user string, newGID string) error {
	path := "/etc/group"
	return patchFile(path, func(line string) (string, error) {
		fs := strings.Split(line, ":")
		if len(fs) != 4 {
			return "", errors.New("unexpected number of fields in /etc/group")
		}
		if fs[0] == user && fs[2] != newGID {
			debugf("replacing %s with gid %s->%s\n", fs[0], fs[2], newGID)
			fs[2] = newGID
		}
		return strings.Join(fs, ":"), nil
	})
}

func patchFile(path string, patchLine func(line string) (string, error)) error {
	// Open file
	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Lock the file to avoid concurrent rewrites
	err = unix.Flock(int(f.Fd()), unix.LOCK_EX)
	if err != nil {
		return err
	}

	// Replace all of the lines
	content := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		newLine, err := patchLine(line)
		if err != nil {
			return err
		}
		content = append(content, newLine)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	output := strings.Join(content, "\n") + "\n"

	// Atomically replace the original file
	t, err := renameio.TempFile("", path)
	if err != nil {
		return err
	}
	defer t.Cleanup()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if err := t.Chmod(fi.Mode().Perm()); err != nil {
		return err
	}

	_, err = t.WriteString(output)
	if err != nil {
		return err
	}
	return t.CloseAtomicallyReplace()
}

// Copied from https://github.com/openshift/images/blob/bcab0f7337420343611546aae2634eaf0d36c33e/pod/pod.go#L27
func reapChildProcesses() {
	for {
		// Pid -1 means to wait for any child process.
		// With syscall.WNOHANG option set, function will
		// not block and will exit immediately if no child
		// process has exited.
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		switch err {
		case nil:
			// If pid == 0 then one or more child processes still exist,
			// but have not yet changed state so we return and wait
			// for another SIGCHLD.
			if pid == 0 {
				return
			}
		case syscall.ECHILD:
			// No more child processes to reap. We can return and wait
			// for another SIGCHLD signal.
			return
		case syscall.EINTR:
			// Got interrupted. Shouldn't happen with WNOHANG option,
			// but it is better to handle it anyway and try again.
		default:
			// Some other unexpected error. Return and wait for
			// another SIGCHLD signal.
			debugf("Unexpected error waiting for child process: %v\n", err)
			return
		}
	}
}
