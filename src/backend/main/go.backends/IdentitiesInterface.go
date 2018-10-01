/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

package backends

import (
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
)

type (
	IdentityStorage interface {
		RetrieveLocalsIdentities(user_id string) ([]UserIdentity, error)
		CreateUserIdentity(userIdentity *UserIdentity) CaliopenError
		RetrieveUserIdentity(userId, RemoteId string, withCredentials bool) (*UserIdentity, error)
		LookupIdentityByIdentifier(string, ...string) ([][2]string, error)
		LookupIdentityByType(string, ...string) ([][2]string, error)
		UpdateUserIdentity(userIdentity *UserIdentity, fields map[string]interface{}) error
		DeleteUserIdentity(userIdentity *UserIdentity) error
		RetrieveRemoteIdentities(userId string, withCredentials bool) ([]*UserIdentity, error)
		RetrieveAllRemotes(withCredentials bool) (<-chan *UserIdentity, error)
		UpdateRemoteInfosMap(userId, remoteId string, infos map[string]string) error
		RetrieveRemoteInfosMap(userId, remoteId string) (infos map[string]string, err error)
		IsLocalIdentity(userId, identityId string) bool
		IsRemoteIdentity(userId, identityId string) bool
		Close()
	}
)
