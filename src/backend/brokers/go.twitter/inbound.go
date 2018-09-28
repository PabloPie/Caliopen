/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

package twitter_broker

import (
	"errors"
	"fmt"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/CaliOpen/go-twitter/twitter"
	log "github.com/Sirupsen/logrus"
	"time"
)

const (
	natsMessageTmpl = "{\"order\":\"%s\",\"user_id\":\"%s\",\"remote_id\":\"%s\",\"message_id\": \"%s\"}"
	natsOrder       = "process_raw"
	NatsError       = "nats error"
)

func (broker *TwitterBroker) ProcessInDM(userID, remoteID UUID, dm *twitter.DirectMessageEvent, rawOnly bool) error {
	rawID, err := broker.SaveRawDM(dm, userID)
	if err != nil {
		return err
	}
	// send process order to nats
	natsMessage := fmt.Sprintf(natsMessageTmpl, natsOrder, userID.String(), remoteID.String(), rawID.String())
	resp, err := broker.NatsConn.Request(broker.Config.NatsTopicFetcher, []byte(natsMessage), 10*time.Second)
	if err != nil {
		if broker.NatsConn.LastError() != nil {
			log.WithError(broker.NatsConn.LastError()).Warnf("[TwitterBroker] failed to publish inbound request on NATS for user %s. Raw message has been saved with id %s", userID.String(), rawID.String())
			log.Infof("natsMessage: %s\nnatsResponse: %+v\n", natsMessage, resp)
		} else {
			log.WithError(err).Warnf("[TwitterBroker] failed to publish inbound request on NATS for user %s. Raw message has been saved with id %s", userID.String(), rawID.String())
			log.Infof("natsMessage: %s\nnatsResponse: %+v\n", natsMessage, resp)
		}
		return errors.New(NatsError)
	}
	return nil
}
