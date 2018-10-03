/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

package twitter_broker

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/CaliOpen/go-twitter/twitter"
	log "github.com/Sirupsen/logrus"
	"time"
)

const (
	natsMessageTmpl  = "{\"order\":\"%s\",\"user_id\":\"%s\",\"remote_id\":\"%s\",\"message_id\": \"%s\"}"
	natsOrder        = "process_raw"
	NatsError        = "nats error"
	lastSeenInfosKey = "lastseendm"
	lastSyncInfosKey = "lastsync"
)

func (broker *TwitterBroker) ProcessInDM(userID, remoteID UUID, dm *twitter.DirectMessageEvent, senderName string, rawOnly bool) error {

	// add sender's screen name before saving it as raw_message
	dm.Message.SenderScreenName = senderName

	rawID, err := broker.SaveRawDM(dm, userID)
	if err != nil {
		return err
	}
	// send process order to nats
	natsMessage := fmt.Sprintf(natsMessageTmpl, natsOrder, userID.String(), remoteID.String(), rawID.String())
	resp, err := broker.NatsConn.Request(broker.Config.LDAConfig.InTopic, []byte(natsMessage), 10*time.Second)
	if err != nil {
		if broker.NatsConn.LastError() != nil {
			log.WithError(broker.NatsConn.LastError()).Warnf("[TwitterBroker] failed to publish inbound request on NATS for user %s. Raw message has been saved with id %s", userID.String(), rawID.String())
			log.Infof("natsMessage: %s\nnatsResponse: %+v\n", natsMessage, resp)
		} else {
			log.WithError(err).Warnf("[TwitterBroker] failed to publish inbound request on NATS for user %s. Raw message has been saved with id %s", userID.String(), rawID.String())
			log.Infof("natsMessage: %s\nnatsResponse: %+v\n", natsMessage, resp)
		}
		return errors.New(NatsError)
	} else {
		nats_ack := new(map[string]interface{})
		err := json.Unmarshal(resp.Data, &nats_ack)
		if err != nil {
			log.WithError(err).Infof("natsMessage: %s\nnatsResponse: %+v\n", natsMessage, resp)
			return errors.New("[ProcessInDM] failed to parse inbound ack on NATS")
		}
		if err, ok := (*nats_ack)["error"]; ok {
			log.WithError(errors.New(err.(string))).Infof("natsMessage: %s\nnatsResponse: %+v\n", natsMessage, resp)
			return errors.New("[ProcessInDM inbound delivery failed")
		}

		//nats delivery OK, update sync status in db
		// TODO: algorithm to shorten pollinterval after new DM has been received
		infos, err := broker.Store.RetrieveRemoteInfosMap(userID.String(), remoteID.String())
		if err != nil {
			return err
		}
		infos[lastSeenInfosKey] = dm.ID
		infos[lastSyncInfosKey] = time.Now().Format(time.RFC3339)
		err = broker.Store.UpdateRemoteInfosMap(userID.String(), remoteID.String(), infos)
		return nil
	}
}
