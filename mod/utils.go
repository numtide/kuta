package mod

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

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
