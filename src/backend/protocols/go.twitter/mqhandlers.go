// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package twitterworker

import (
	"encoding/json"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/go-nats"
	"time"
)

func NatsMsgHandler(msg *nats.Msg) {
	message := TwitterOrder{}
	err := json.Unmarshal(msg.Data, &message)
	if err != nil {
		log.WithError(err).Errorf("Unable to unmarshal message from NATS. Payload was <%s>", string(msg.Data))
		return
	}
	switch message.Order {
	case "sync":
		//TODO
		log.Infof("received sync order for remote twitter ID %s", message.RemoteId)
		if worker, ok := TwitterWorkers[message.UserId+message.RemoteId]; ok {
			select {
			case worker.WorkerDesk <- PollDM:
				log.Infof("[NatsMsgHandler] ordering to pollDM for remote %s (user %s)", message.RemoteId, message.UserId)
			case <-time.After(30 * time.Second):
				log.Warnf("[NatsMsgHandler] worker's desk is full for remote %s (user %s)", message.RemoteId, message.UserId)
			}
		} else {
			log.Warnf("[NatsMsgHandler] failed to retrieve registered worker for remote %s (user %s). Trying to add one.", message.RemoteId, message.UserId)
			worker, err := NewWorker(AppConfig, message.UserId, message.RemoteId)
			if err != nil {
				log.WithError(err).Warnf("[NatsMsgHandler] failed to create new worker for remote %s (user %s)", message.RemoteId, message.UserId)
			}
			registerWorker(worker)
			worker.Start()
			select {
			case worker.WorkerDesk <- PollDM:
				log.Infof("[NatsMsgHandler] ordering to pollDM for remote %s (user %s)", message.RemoteId, message.UserId)
			case <-time.After(30 * time.Second):
				log.Warnf("[NatsMsgHandler] worker's desk is full for remote %s (user %s)", message.RemoteId, message.UserId)
			}
		}
	case "reload_worker":
		//TODO: order to force refreshing cache data for an account
		log.Infof("received reload_worker order for remote twitter ID %s", message.RemoteId)
	case "add_worker":
		log.Infof("received add_worker order for remote twitter ID %s", message.RemoteId)
		worker, err := NewWorker(AppConfig, message.UserId, message.RemoteId)
		if err != nil {
			log.WithError(err).Warnf("[NatsMsgHandler] failed to create new worker for remote %s (user %s)", message.RemoteId, message.UserId)
		}
		registerWorker(worker)
		worker.Start()
	case "remove_worker":
		//TODO
		log.Infof("received remove_worker order for remote twitter ID %s", message.RemoteId)
	}
}
