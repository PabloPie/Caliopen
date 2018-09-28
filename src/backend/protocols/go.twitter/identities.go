/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package twitterworker

import (
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
)

// logic to retrieve users' Twitter account credentials from db
// and to listen to NATS messages regarding users' account creation/update/deletion

// RetrieveAccountsCredentials is called at startup to fetch active twitter remote identities from store.
// It returns account token & secret for each identity found.
func RetrieveTwitterAccounts() (accounts []TwitterAccount, err error) {
	//TODO: get list of twitter accounts to start workers for from db

	accounts = []TwitterAccount{
		{
			accessToken:       "2353377708-mijYkXocnStNt9ZttKqLWi6UicCt251UQRi1ivs",
			accessTokenSecret: "dGi5OgwjDJjlSIDpR9VTp2vSPFlEaXNCyOBHFhbh4y4aA",
			userID:            EmptyUUID,
			remoteID:          EmptyUUID,
		},
	}
	return
}

// NatsMessageHandler is called by nats listener to parse messages regarding remote identities.
// It triggers actions accordingly.
func NatsMessageHandler() error {

	return nil
}
