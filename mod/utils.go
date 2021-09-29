package mod

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

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
			log.Printf("[kuta] replacing %s uid %s->%s and gid %s->%s\n", fs[0], fs[2], newUID, fs[3], newGID)
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
			log.Printf("[kuta] replacing %s with gid %s->%s\n", fs[0], fs[2], newGID)
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
	output := strings.Join(content, "\n")

	// Rewind and write to file
	_, err = f.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(output))
	return err
}
