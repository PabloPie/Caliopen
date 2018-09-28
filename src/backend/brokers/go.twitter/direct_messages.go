/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

package twitter_broker

import (
	"encoding/json"
	"fmt"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/CaliOpen/go-twitter/twitter"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"time"
)

// SaveRawDM marshal DM to json and save it as a raw message object in store
func (broker *TwitterBroker) SaveRawDM(dm *twitter.DirectMessageEvent, userId UUID) (rawMessageId UUID, err error) {

	jsonDM, e := json.Marshal(dm)
	if e != nil {
		err = fmt.Errorf("[Twitter Broker]SaveRawDM failed to marshal dm to json : %s", e)
		return
	}

	rawMsg := RawMessage{
		Raw_msg_id: UUID(uuid.NewV4()),
		Raw_Size:   uint64(len(jsonDM)),
		Raw_data:   string(jsonDM),
		Delivered:  false,
	}

	e = broker.Store.StoreRawMessage(rawMsg)
	if e != nil {
		err = fmt.Errorf("[Twitter Broker]SaveRawDM failed to store raw message in store : %s", e)
		return
	}

	return rawMsg.Raw_msg_id, nil
}

// UnmarshalDM creates a new Caliopen Message entity from a twitter1.1's DM event
func UnmarshalDM(dm *twitter.DirectMessageEvent, userId UUID) (message *Message, err error) {
	dmDate, e := time.Parse(time.RFC3339, dm.CreatedAt)
	if e != nil {
		return nil, fmt.Errorf("[Twitter Broker]UnmarshalDM failed to parse created_at string : %e", e)
	}
	message = &Message{
		Date:             dmDate,
		Date_insert:      time.Now(),
		Is_received:      true,
		Message_id:       UUID(uuid.NewV4()),
		Participants:     []Participant{},
		Privacy_features: &PrivacyFeatures{},
		Type:             TwitterProtocol,
		User_id:          userId,
	}
	return nil, errors.New("not implemented")
}
