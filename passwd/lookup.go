// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package passwd

import (
	"os/user"
)

// Current returns the current user.
//
// The first call will cache the current user information.
// Subsequent calls will return the cached value and will not reflect
// changes to the current user.
func Current() (*user.User, error) {
	return current()
}

// Lookup looks up a user by username. If the user cannot be found, the
// returned error is of type user.UnknownUserError.
func Lookup(username string) (*user.User, error) {
	return lookupUser(username)
}

// LookupID looks up a user by userid. If the user cannot be found, the
// returned error is of type user.UnknownUserIdError.
func LookupID(uid string) (*user.User, error) {
	return lookupUserID(uid)
}

// LookupGroup looks up a group by name. If the group cannot be found, the
// returned error is of type user.UnknownGroupError.
func LookupGroup(name string) (*user.Group, error) {
	return lookupGroup(name)
}

// LookupGroupID looks up a group by groupid. If the group cannot be found, the
// returned error is of type user.UnknownGroupIdError.
func LookupGroupID(gid string) (*user.Group, error) {
	return lookupGroupID(gid)
}
