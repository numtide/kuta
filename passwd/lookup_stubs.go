// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package passwd

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
)

func current() (*user.User, error) {
	uid := currentUID()
	// $USER and /etc/passwd may disagree; prefer the latter if we can get it.
	// See issue 27524 for more information.
	u, err := lookupUserID(uid)
	if err == nil {
		return u, nil
	}

	homeDir, _ := os.UserHomeDir()
	u = &user.User{
		Uid:      uid,
		Gid:      currentGID(),
		Username: os.Getenv("USER"),
		Name:     "", // ignored
		HomeDir:  homeDir,
	}
	// cgo isn't available, but if we found the minimum information
	// without it, use it:
	if u.Uid != "" && u.Username != "" && u.HomeDir != "" {
		return u, nil
	}
	var missing string
	if u.Username == "" {
		missing = "$USER"
	}
	if u.HomeDir == "" {
		if missing != "" {
			missing += ", "
		}
		missing += "$HOME"
	}
	return u, fmt.Errorf("user: Current requires cgo or %s set in environment", missing)
}

func currentUID() string {
	if id := os.Getuid(); id >= 0 {
		return strconv.Itoa(id)
	}
	// Note: Windows returns -1, but this file isn't used on
	// Windows anyway, so this empty return path shouldn't be
	// used.
	return ""
}

func currentGID() string {
	if id := os.Getgid(); id >= 0 {
		return strconv.Itoa(id)
	}
	return ""
}
