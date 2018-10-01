// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package twitterworker

import (
	"github.com/nats-io/go-nats"
	"sync"
)

var (
	WorkersGuard   *sync.RWMutex
	TwitterWorkers map[string]*Worker
	NatsConn       *nats.Conn
)

func registerWorker(worker *Worker) {
	workerKey := worker.userAccount.userID.String() + worker.userAccount.remoteID.String()
	WorkersGuard.Lock()
	if _, ok := TwitterWorkers[workerKey]; !ok { // do not register same worker twice
		TwitterWorkers[workerKey] = worker
	}
	WorkersGuard.Unlock()
}
