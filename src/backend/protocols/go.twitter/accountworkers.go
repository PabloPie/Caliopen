// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package twitterworker

import (
	"encoding/json"
	"errors"
	"fmt"
	broker "github.com/CaliOpen/Caliopen/src/backend/brokers/go.twitter"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/CaliOpen/go-twitter/twitter"
	"github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"github.com/dghubble/oauth1"
	"sort"
	"strconv"
)

type Worker struct {
	WorkerDesk       chan uint
	broker           *broker.TwitterBroker
	lastDMseen       string
	twitterClient    *twitter.Client
	userAccount      *TwitterAccount
	usersScreenNames map[int64]string // a cache facility to avoid calling too often twitter API for screen_name lookup
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

	worker.usersScreenNames = map[int64]string{}

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
		if e, ok := err.(twitter.APIError); ok {
			for _, err := range e.Errors {
				if err.Code == 88 {
					log.Infof("twitter returned rate limit error, slowing down worker for account @%s", worker.userAccount.screenName)
					accountInfos, err := worker.broker.Store.RetrieveRemoteInfosMap(worker.userAccount.userID.String(), worker.userAccount.remoteID.String())
					if err == nil {
						if pollInterval, ok := accountInfos["pollinterval"]; ok {
							interval, e := strconv.Atoi(pollInterval)
							if e == nil {
								newInterval := strconv.Itoa(interval * 2)
								accountInfos["pollinterval"] = newInterval
								worker.broker.Store.UpdateRemoteInfosMap(worker.userAccount.userID.String(), worker.userAccount.remoteID.String(), accountInfos)
								order := RemoteIDNatsMessage{
									IdentityId:   worker.userAccount.remoteID.String(),
									Order:        "update_interval",
									PollInterval: newInterval,
									Protocol:     "twitter",
									UserId:       worker.userAccount.userID.String(),
								}
								jorder, jerr := json.Marshal(order)
								if jerr == nil {
									worker.broker.NatsConn.Publish("idCache", jorder)
								}
							}
						}
					}
				}
			}
		} else {
			logrus.WithError(err).Warnf("[TwitterWorker] PollDM failed for account @%s (remote %s, user %s)", worker.userAccount.screenName, worker.userAccount.remoteID.String(), worker.userAccount.userID.String())
		}
	}

	sort.Sort(ByAscID(DMs.Events)) // reverse events order to get older DMs first
	if len(DMs.Events) > 0 && worker.dmNotSeen(DMs.Events[0]) {
		//TODO: handle pagination with `cursor` param
	}

	for _, event := range DMs.Events {
		if worker.dmNotSeen(event) {
			//lookup sender's screen_name because it's not embedded in event object
			var senderName string
			var inCache bool
			senderID, err := strconv.ParseInt(event.Message.SenderID, 10, 64)
			if err == nil {
				if senderName, inCache = worker.usersScreenNames[senderID]; !inCache {
					users, _, err := worker.twitterClient.Users.Lookup(&twitter.UserLookupParams{UserID: []int64{senderID}})
					if err == nil && len(users) > 0 {
						senderName = users[0].ScreenName
						worker.usersScreenNames[senderID] = senderName
					}
				}
			}
			err = worker.broker.ProcessInDM(worker.userAccount.userID, worker.userAccount.remoteID, &event, senderName, true)
			if err != nil {
				// something went wrong, forget this DM id
				logrus.WithError(err).Warnf("[TwitterWorker] ProcessInDM failed for event : %+v", event)
				continue
			}
			worker.lastDMseen = event.ID
		}
	}
	//TODO: write lastsync & last_check to remote identity in db at last
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
