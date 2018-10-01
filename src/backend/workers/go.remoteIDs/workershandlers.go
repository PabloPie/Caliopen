// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package go_remoteIDs

import (
	"encoding/json"
	"errors"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	log "github.com/Sirupsen/logrus"
)

// addWorkerFor publishes message on nats queue directed at relevant worker
func (p *Poller) addWorkerFor(idkey string) error {
	var entry cacheEntry
	var ok bool
	if entry, ok = p.Cache[idkey]; !ok {
		return errors.New("cache entry not found")
	}
	switch entry.remoteProtocol {
	case "twitter":
		natsMsg := TwitterOrder{
			Order:    "add_worker",
			UserId:   entry.userID.String(),
			RemoteId: entry.remoteID.String(),
		}
		j, e := json.Marshal(natsMsg)
		if e != nil {
			log.WithError(e).Warn("[addWorkerFor] failed to marshal nats message")
			return e
		}
		return p.NatsConn.Publish(p.Config.NatsTopics["twitter"], j)

	}
	return nil
}

// removeWorkerFor publishes message on nats queue directed at relevant worker
func (p *Poller) removeWorkerFor(idkey string) error {
	var entry cacheEntry
	var ok bool
	if entry, ok = p.Cache[idkey]; !ok {
		return errors.New("cache entry not found")
	}
	switch entry.remoteProtocol {
	case "twitter":
		natsMsg := TwitterOrder{
			Order:    "remove_worker",
			UserId:   entry.userID.String(),
			RemoteId: entry.remoteID.String(),
		}
		j, e := json.Marshal(natsMsg)
		if e != nil {
			log.WithError(e).Warn("[addWorkerFor] failed to marshal nats message")
			return e
		}
		return p.NatsConn.Publish(p.Config.NatsTopics["twitter"], j)

	}
	return nil
}

// updateWorkerFor publishes message on nats queue directed at relevant worker
func (p *Poller) updateWorkerFor(idkey string) error {
	var entry cacheEntry
	var ok bool
	if entry, ok = p.Cache[idkey]; !ok {
		return errors.New("cache entry not found")
	}
	switch entry.remoteProtocol {
	case "twitter":
		natsMsg := TwitterOrder{
			Order:    "reload_worker",
			UserId:   entry.userID.String(),
			RemoteId: entry.remoteID.String(),
		}
		j, e := json.Marshal(natsMsg)
		if e != nil {
			log.WithError(e).Warn("[addWorkerFor] failed to marshal nats message")
			return e
		}
		return p.NatsConn.Publish(p.Config.NatsTopics["twitter"], j)

	}
	return nil
}
