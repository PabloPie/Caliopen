/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

package go_remoteIDs

import (
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/Sirupsen/logrus"
	"gopkg.in/robfig/cron.v2"
)

type cacheEntry struct {
	cronId         cron.EntryID
	iDkey          string
	pollInterval   string // in minutes
	remoteID       UUID
	remoteProtocol string
	userID         UUID
}

// updateCache fetches remote identities from db and adds them to poller cache
// it returns entries that have been added/removed/updated
func (p *Poller) updateCache() (added, removed, updated map[string]bool, err error) {
	added = make(map[string]bool)
	removed = make(map[string]bool)
	updated = make(map[string]bool)
	active := make(map[string]bool)
	const defaultInterval = "15"

	remotes, err := p.Store.RetrieveAllRemotes(false)
	if err != nil {
		logrus.WithError(err).Warn("[updateCache] failed to retrieve remote identities")
		return
	}

	p.cacheMux.Lock()
	for remote := range remotes {
		if p.statusTypeOK(remote) {
			idkey := remote.UserId.String() + remote.Id.String()
			active[idkey] = true
			if entry, ok := p.Cache[idkey]; ok {
				//check if pollinterval has changed
				var pollInterval string
				var ok bool
				if pollInterval, ok = remote.Infos["pollinterval"]; !ok || pollInterval == "" {
					// do not resign, take a default value instead
					pollInterval = defaultInterval
				}
				if entry.pollInterval != pollInterval {
					entry.pollInterval = pollInterval
					updated[idkey] = true
				}
				p.Cache[idkey] = entry
			} else {
				var pollInterval string
				var ok bool
				if pollInterval, ok = remote.Infos["pollinterval"]; !ok || pollInterval == "" {
					// do not resign, take a default value instead
					pollInterval = defaultInterval
				}
				p.Cache[idkey] = cacheEntry{
					iDkey:          idkey,
					pollInterval:   pollInterval,
					remoteID:       remote.Id,
					remoteProtocol: remote.Protocol,
					userID: remote.UserId,
				}
				added[idkey] = true
			}
		}
	}
	for key := range p.Cache {
		if _, ok := active[key]; !ok {
			removed[key] = true
		}
	}
	p.cacheMux.Unlock()
	return
}

// statusTypeOK checks if remote identity is 'active' and its type is within poller's scope
func (p *Poller) statusTypeOK(remote *UserIdentity) bool {
	if remote.Status != "active" {
		return false
	}
	for _, t := range p.Config.RemoteProtocols {
		if t == remote.Protocol {
			return true
		}
	}
	return false
}
