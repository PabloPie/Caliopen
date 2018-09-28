/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

package twitter_broker

import (
	"github.com/nats-io/go-nats"
)

// retrieves a caliopen message from db, build a DM from it
// sends it through twitter API and stores the raw direct message sent in db
func (broker *TwitterBroker) natsMsgHandler(msg *nats.Msg) (resp []byte, err error) {

	return
}
