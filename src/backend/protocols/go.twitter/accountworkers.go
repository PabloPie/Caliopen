// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package twitterworker

import (
	"errors"
	"fmt"
	broker "github.com/CaliOpen/Caliopen/src/backend/brokers/go.twitter"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/CaliOpen/go-twitter/twitter"
	"github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"github.com/dghubble/oauth1"
	"sort"
)

type Worker struct {
	WorkerDesk    chan uint
	broker        *broker.TwitterBroker
	lastDMseen    string
	twitterClient *twitter.Client
	userAccount   *TwitterAccount
}

type TwitterAccount struct {
	accessToken       string
	accessTokenSecret string
	userID            UUID
	remoteID          UUID
	screenName        string
}

const (
	//WorkerDesk commands
	PollDM = uint(iota)
	PollTimeLine
	Stop

	lastSeenInfosKey = "lastseendm"
	lastSyncInfosKey = "lastsync"
)

// NewWorker create a worker dedicated to a specific twitter account.
// A worker holds remote identity credentials and data, as well as user context connection to twitter API.
func NewWorker(config *WorkerConfig, userID, remoteID string) (worker *Worker, err error) {
	worker = new(Worker)
	worker.WorkerDesk = make(chan uint, 3)
	b, e := broker.Initialize(config.BrokerConfig)
	if e != nil {
		err = fmt.Errorf("[TwitterWorker]NewWorker failed to initialize a twitter broker : %s", e)
		return nil, err
	}
	worker.broker = b
	var remote *UserIdentity
	// retrieve data from db
	remote, err = worker.broker.Store.RetrieveUserIdentity(userID, remoteID, true)
	if err != nil {
		log.WithError(err).Infof("[PollDM] failed to retrieve remote identity <%s> (user <%s>)", worker.userAccount.remoteID, worker.userAccount.userID)
		return
	}
	if remote.Credentials == nil {
		log.WithError(err).Infof("[PollDM] failed to retrieve credentials for remote identity <%s> (user <%s>)", worker.userAccount.remoteID, worker.userAccount.userID)
		return
	}
	worker.userAccount = &TwitterAccount{
		accessToken:       (*remote.Credentials)["token"],
		accessTokenSecret: (*remote.Credentials)["secret"],
		userID:            remote.UserId,
		remoteID:          remote.Id,
		screenName:        remote.Identifier,
	}
	if lastseen, ok := remote.Infos[lastSeenInfosKey]; ok {
		worker.lastDMseen = lastseen
	}

	authConf := oauth1.NewConfig(config.TwitterAppKey, config.TwitterAppSecret)
	token := oauth1.NewToken(worker.userAccount.accessToken, worker.userAccount.accessTokenSecret)
	httpClient := authConf.Client(oauth1.NoContext, token)
	if worker.twitterClient = twitter.NewClient(httpClient); worker.twitterClient == nil {
		return nil, errors.New("[NewWorker] twitter api failed to create http client")
	}

	return
}

func (worker *Worker) Start() error {
	go func(w *Worker) {
	deskLoop:
		for command := range worker.WorkerDesk {
			switch command {
			case PollDM:
				worker.PollDM()
			case Stop:
				break deskLoop
			default:
				logrus.Warnf("worker received unknown command number %d", command)
			}
		}
	}(worker)
	return nil
}

// PollDM calls Twitter API endpoint to fetch DM for worker
func (worker *Worker) PollDM() {
	DMs, _, err := worker.twitterClient.DirectMessages.EventsList(&twitter.DirectMessageEventsListParams{Count: 10})
	if err != nil {
		logrus.Warn(err)
	}

	sort.Sort(ByAscID(DMs.Events)) // reverse events order to get older DMs first
	if worker.dmNotSeen(DMs.Events[0]) {
		//TODO: handle pagination with `cursor` param
	}

	for _, event := range DMs.Events {
		if worker.dmNotSeen(event) {
			err := worker.broker.ProcessInDM(worker.userAccount.userID, worker.userAccount.remoteID, &event, true)
			if err != nil {
				logrus.WithError(err).Warnf("[TwitterWorker] ProcessInDM failed for event : %+v", event)
				if err.Error() != broker.NatsError {
					// raw message has not been saved, don't go further
					continue
				}
			}
			worker.lastDMseen = event.ID
		}
	}
}

func (worker *Worker) dmNotSeen(event twitter.DirectMessageEvent) bool {
	return worker.lastDMseen < event.ID
}

// sort interface
type ByAscID []twitter.DirectMessageEvent

func (bri ByAscID) Len() int {
	return len(bri)
}

func (bri ByAscID) Less(i, j int) bool {
	return bri[i].ID < bri[j].ID
}

func (bri ByAscID) Swap(i, j int) {
	bri[i], bri[j] = bri[j], bri[i]
}
