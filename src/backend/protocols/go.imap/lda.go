/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

package imap_worker

import (
	"errors"
	"fmt"
	broker "github.com/CaliOpen/Caliopen/src/backend/brokers/go.emails"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/satori/go.uuid"
	"time"
)

//Local Delivery Agent, in charge of IO between fetcher and our email broker
type Lda struct {
	Config           WorkerConfig
	broker           *broker.EmailBroker
	brokerConnectors broker.EmailBrokerConnectors
}

func NewLda(config WorkerConfig) (*Lda, error) {
	var err error
	lda := Lda{}
	lda.Config = config
	lda.broker, lda.brokerConnectors, err = broker.Initialize(config.LDAConfig)
	return &lda, err
}

func (lda *Lda) shutdown() error {
	lda.broker.ShutDown()
	return nil
}

func (lda *Lda) deliverMail(mail *Email, userId string) (err error) {
	emailMsg := &EmailMessage{
		Email: mail,
		Message: &Message{
			User_id: UUID(uuid.FromStringOrNil(userId)),
		},
	}
	incoming := &broker.SmtpEmail{
		EmailMessage: emailMsg,
		Response:     make(chan *DeliveryAck),
	}
	defer close(incoming.Response)

	lda.brokerConnectors.Ingress <- incoming

	select {
	case response := <-incoming.Response:
		if response.Err {
			return errors.New(fmt.Sprintf("[deliverMail] Error : " + response.Response))
		}
		return nil
	case <-time.After(30 * time.Second):
		return errors.New("[deliverMail] LDA timeout")
	}
}
