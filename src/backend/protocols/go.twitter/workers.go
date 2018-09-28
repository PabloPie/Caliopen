// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package twitterworker

import (
	"encoding/json"
	"fmt"
	broker "github.com/CaliOpen/Caliopen/src/backend/brokers/go.twitter"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/CaliOpen/go-twitter/twitter"
	"github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"github.com/dghubble/oauth1"
	"github.com/nats-io/go-nats"
	"sort"
	"sync"
	"time"
)

type Worker struct {
	PollDesk      chan uint
	broker        *broker.TwitterBroker
	id            int
	lastDMseen    string
	nextPoll      time.Time
	pollInterval  time.Duration
	twitterClient *twitter.Client
	userAccount   *TwitterAccount
}

type TwitterAccount struct {
	accessToken       string
	accessTokenSecret string
	userID            UUID
	remoteID          UUID
}

const (
	//pollDesk commands
	PollDM = uint(iota)
	PollTimeLine
	Stop
)

var (
	WorkersGuard   *sync.RWMutex
	TwitterWorkers []*Worker
	NatsConn       *nats.Conn
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
	case "cache_update":
		//TODO
	}
}

func NewWorker(id int, config *WorkerConfig, account TwitterAccount, pollInterval int, lastDMseen string) (worker *Worker, err error) {
	worker = new(Worker)
	worker.PollDesk = make(chan uint, 3)
	b, e := broker.Initialize(config.BrokerConfig)
	if e != nil {
		err = fmt.Errorf("[TwitterWorker]NewWorker failed to initialize a twitter broker : %s", e)
		return nil, err
	}
	worker.broker = b
	worker.id = id
	worker.lastDMseen = lastDMseen
	worker.pollInterval = time.Duration(pollInterval)

	authConf := oauth1.NewConfig("TSpJhxzkNo43Q0d64Vz10a29I", "kU4mWmFXZPXdzztePHA0uwU57m6CpfbGp2TNzpwgmDiGyjUyxL")
	token := oauth1.NewToken("2353377708-mijYkXocnStNt9ZttKqLWi6UicCt251UQRi1ivs", "dGi5OgwjDJjlSIDpR9VTp2vSPFlEaXNCyOBHFhbh4y4aA")
	httpClient := authConf.Client(oauth1.NoContext, token)
	worker.twitterClient = twitter.NewClient(httpClient)

	worker.userAccount = &account

	return
}

func (worker *Worker) Start() error {
	worker.nextPoll = time.Now().Add(worker.pollInterval * time.Second) // TODO
	go func(w *Worker) {
	deskLoop:
		for command := range worker.PollDesk {
			switch command {
			case PollDM:
				worker.PollDM()
			case Stop:
				worker.Stop()
				break deskLoop
			default:
				logrus.Warnf("worker %d received unknown command number %d", worker.id, command)
			}
		}
	}(worker)
	return nil
}

func (worker *Worker) Stop() error {
	//TODO
	return nil
}

func (worker *Worker) IsDue() bool {
	return time.Now().After(worker.nextPoll)
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
				logrus.WithError(err).Warnf("[TwitterWorker](worker %d) ProcessInDM failed for event : %+v", worker.id, event)
				if err.Error() != broker.NatsError {
					// raw message has not been saved, don't go further
					continue
				}
			}
			worker.lastDMseen = event.ID
		}
	}
	worker.nextPoll = worker.nextPoll.Add(worker.pollInterval * time.Second)
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
