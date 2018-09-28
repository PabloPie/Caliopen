// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package go_remoteIDs

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
)

type twitterJob struct {
	remoteId  string
	natsTopic string
	poller    *Poller
	userId    string
}

func (j twitterJob) Run() {
	msg, err := json.Marshal(TwitterOrder{
		Order:    "sync",
		UserId:   j.userId,
		RemoteId: j.remoteId,
	})
	if err != nil {
		logrus.WithError(err).Fatal("unable to marshal natsOrder")
	}

	j.poller.NatsConn.Publish(j.natsTopic, msg)
	j.poller.NatsConn.Flush()

	if err := j.poller.NatsConn.LastError(); err != nil {
		logrus.WithError(err).Fatal("nats publish failed")
	}

	logrus.Infof("ordering to sync Twitter DMs for remote %s (user %s)", j.remoteId, j.userId)
}
