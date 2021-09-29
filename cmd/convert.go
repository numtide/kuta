package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

type convertFlag struct {
	userName string
	uid      string
	gid      string
}

var kf convertFlag
var curr_user *user.User

func init() {
	kf = convertFlag{}
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	curr_user = usr
	convert.Flags().StringVar(&kf.userName, "user", curr_user.Username, "username for uid and guid changes")
	convert.Flags().StringVar(&kf.uid, "uid", "1000", "uid that want to be changed")
	convert.Flags().StringVar(&kf.gid, "gid", "1000", "gid that want to be changed")
	rootCmd.AddCommand(convert)
}

var convert = &cobra.Command{
	Use:   "convert",
	Short: "Convert uid and guid of directory",
	Long:  `Convert all files' uid and guid in the selected directory`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("Current user detected: %s\n", kf.userName)
		if kf.uid != curr_user.Uid {
			fmt.Printf("[UPDATING]: uid=%s to uid=%s\n", curr_user.Uid, kf.uid)
		}
		if kf.gid != curr_user.Gid {
			fmt.Printf("[UPDATING]: gid=%s to gid=%s\n", curr_user.Gid, kf.gid)
		}

		fmt.Println("Replacing uid and/or guid in /etc/passwd..")
		if got := ReplacePASSWD(&kf.userName, &kf.uid, &kf.gid); got != nil {
			panic(got)
		}
		fmt.Println("Replacing guid in /etc/group")
		if got := ReplaceGroup(&kf.userName, &kf.uid, &kf.gid); got != nil {
			panic(got)
		}

		// filepath.Walk
		files, err := WalkAllFilesInDir(args[0])
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			fmt.Println(file)
		}

		fmt.Printf(
			"uid=%d euid=%d gid=%d egid=%d\n",
			syscall.Getuid(), syscall.Geteuid(),
			syscall.Getgid(), syscall.Getegid(),
		)

		// uid64, _ := strconv.ParseUint(kf.uid, 2, 32)
		// attr := &os.ProcAttr{
		// 	Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		// 	Sys: &syscall.SysProcAttr{
		// 		Credential: &syscall.Credential{
		// 			Uid: uint32(uid64),
		// 		},
		// 	},
		// }

		// usermod := []string{
		// 	"/usr/sbin/usermod",
		// 	"-u",
		// 	kf.uid,
		// 	kf.userName,
		// }
		// syscall.Setuid(1003)
		// process, err := os.StartProcess("/usr/bin/usermod", usermod, attr)
		// if err != nil {
		// 	panic(err)
		// }
		// _, err = process.Wait()
		// if err != nil {
		// 	panic(err)
		// }
	},
}

func WalkAllFilesInDir(arg string) ([]string, error) {
	var files []string
	err := filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func ReplacePASSWD(user *string, newUid *string, newGid *string) error {
	path := "/etc/passwd"
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	return replacePASSWD(file, path, user, newUid, newGid)
}

func replacePASSWD(r io.Reader, path string, user *string, newUid *string, newGid *string) error {
	lines := bufio.NewReader(r)
	content := []string{}

	for {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}
		name, err := parsePasswd(string(copyBytes(line)), user, newUid, newGid)
		if err != nil {
			return err
		}
		content = append(content, name)
	}
	output := strings.Join(content, "\n")
	err := ioutil.WriteFile(path, []byte(output), 0644)
	if err != nil {
		panic(err)
	}
	return nil
}

func parsePasswd(line string, user *string, newUid *string, newGid *string) (string, error) {
	fs := strings.Split(line, ":")
	if len(fs) != 7 {
		return "", errors.New("unexpected number of fields in /etc/passwd")
	}
	if fs[0] == *user {
		fs[2] = *newUid
		fs[3] = *newGid
		fmt.Printf("replacing %s with uid=%s and guid=%s\n", fs[0], fs[2], fs[3])
	}
	return strings.Join(fs, ":"), nil
}

func ReplaceGroup(user *string, newUid *string, newGid *string) error {
	path := "/etc/group"
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	return replaceGroup(file, path, user, newUid, newGid)
}

func replaceGroup(r io.Reader, path string, user *string, newUid *string, newGid *string) error {
	lines := bufio.NewReader(r)
	content := []string{}

	for {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}
		name, err := parseGroup(string(copyBytes(line)), user, newGid)
		if err != nil {
			return err
		}
		content = append(content, name)
	}
	output := strings.Join(content, "\n")
	err := ioutil.WriteFile(path, []byte(output), 0644)
	if err != nil {
		panic(err)
	}
	return nil
}

func parseGroup(line string, user *string, newGid *string) (string, error) {
	fs := strings.Split(line, ":")
	if len(fs) != 4 {
		return "", errors.New("unexpected number of fields in /etc/group")
	}
	if fs[0] == *user {
		fs[2] = *newGid
		fmt.Printf("replacing %s with guid=%s\n", fs[0], fs[2])
	}
	return strings.Join(fs, ":"), nil
}

func copyBytes(x []byte) []byte {
	y := make([]byte, len(x))
	copy(y, x)
	return y
}
