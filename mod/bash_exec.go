package mod

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"syscall"
)

// BashExec runs the target program inside of a bash login shell to make sure
// all the env vars from the profiles are loaded. If no argument is passed,
// just run bash as a shell.
func BashExec(args []string) (int, error) {
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		return -1, fmt.Errorf("bash: %w", err)
	}

	targetUser, err := user.LookupId(fmt.Sprintf("%d", syscall.Getuid()))
	if err != nil {
		return -1, fmt.Errorf("user lookup: %w", err)
	}

	if len(args) == 0 {
		args = []string{bashPath, "--login"}
	} else {
		args = append([]string{bashPath, "--login", "-c", "exec \"$@\"", args[0]}, args...)
	}

	attr := &os.ProcAttr{
		// Override USER and HOME with the new values
		Env: append(os.Environ(),
			fmt.Sprintf("USER=%s", targetUser.Username),
			fmt.Sprintf("HOME=%s", targetUser.HomeDir),
		),

		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},

		// Drop the euid and egid
		Sys: &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(syscall.Getuid()),
				Gid: uint32(syscall.Getgid()),
			},
		},
	}

	// Trap signals and forward them to the child process
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)

	debug(args, attr)

	// Run the program
	proc, err := os.StartProcess(args[0], args, attr)
	if err != nil {
		return -1, err
	}

	// Forward signals to the process
	go func() {
		for sig := range sigs {
			proc.Signal(sig)
		}
	}()

	// Wait for the main program to exit
	state, err := proc.Wait()
	if state == nil {
		return -1, err
	}
	return state.ExitCode(), nil
}
