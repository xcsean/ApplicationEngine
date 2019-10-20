package asset

import (
	"sync"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/mysql"
)

type assetData struct {
	asset *protocol.GhostUserasset
}

var (
	dbpool  *mysql.DB
	assets  map[uint64]*assetData
	reqC    chan *Req
	exitC   chan struct{}
	wg      sync.WaitGroup
	flag    bool
	ghostID int64
)

func init() {
	assets = make(map[uint64]*assetData)
	reqC = make(chan *Req, 3000)
	exitC = make(chan struct{})
}

func start(id int64, pool *mysql.DB) {
	if !flag {
		flag = true
		ghostID = id
		dbpool = pool
		wg.Add(1)
		go assetLoop()
	}
}

func stop() {
	close(exitC)
	wg.Wait()
}

func assetLoop() {
	defer wg.Add(-1)
	log.Debug("asset loop start...")

	for {
		select {
		case <-exitC:
			goto exit
		}
	}

exit:
	log.Debug("asset loop exit...")
}
