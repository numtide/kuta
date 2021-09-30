package mod

import (
	"fmt"
	"os"
	"os/user"
	"syscall"

	"github.com/numtide/kuta/passwd"
)

// PatchUser
func PatchUser() error {
	var (
		err       error
		patchUser *user.User
	)

	// Patch the container /etc/passwd, /etc/group and home permissions.
	targetUID := syscall.Getuid()
	targetGID := syscall.Getgid()
	targetUIDStr := fmt.Sprintf("%d", targetUID)
	targetGIDStr := fmt.Sprintf("%d", targetGID)

	// If a user with the process UID already exists, use that as the target user.
	targetUser, err := passwd.LookupID(targetUIDStr)
	if err == nil {
		debug("found existing user with id", targetUID)
		patchUser = targetUser
	} else {
		// Otherwise lookup using the USER environment variable
		username := os.Getenv("USER")
		if username == "" {
			return fmt.Errorf("please set the USER environment variable to the user you want to mutate")
		}
		patchUser, err = passwd.Lookup(username)
		if err != nil {
			return fmt.Errorf("user lookup error: %w", err)
		}
	}

	if patchUser.Uid != targetUIDStr || patchUser.Gid != targetGIDStr {
		debug("patching", patchUser)

		if syscall.Geteuid() != 0 {
			return fmt.Errorf("entrypoint has no setuid or is not owned by root")
		}

		// Update /etc/passwd
		err = patchEtcPasswd(patchUser.Username, targetUIDStr, targetGIDStr)
		if err != nil {
			return err
		}

		if patchUser.Gid != targetGIDStr {
			// Update /etc/group
			err = patchEtcGroup(patchUser.Username, targetGIDStr)
			if err != nil {
				return err
			}
		}

		if patchUser.HomeDir != "" {
			// Chown the home folder to the new UID/GID
			err = chownR(patchUser.HomeDir, targetUID, targetGID)
			if err != nil {
				return fmt.Errorf("home chown error: %w", err)
			}
		}
	}

	return nil
}
