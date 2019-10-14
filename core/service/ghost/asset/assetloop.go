package asset

import (
	"sync"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/mysql"
)

const (
	modForUUID = 7
)

type assetOwner struct {
	ownerType   uint8
	ownerID     uint64
	expiredTime int64
}

type assetData struct {
	owner assetOwner
	asset *protocol.GhostUserasset
}

var (
	dbpool *mysql.DB
	assets map[uint64]*assetData
	reqC   chan *Req
	exitC  chan struct{}
	wg     sync.WaitGroup
	flag   bool
)

func init() {
	assets = make(map[uint64]*assetData)
	reqC = make(chan *Req, 3000)
	exitC = make(chan struct{})
}

func start(pool *mysql.DB) {
	dbpool = pool
	wg.Add(1)
	go assetLoop()
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
		case req := <-reqC:
			if req.Type == assetCmdLock {
				now := uint64(time.Now().Unix())
				req.RspChannel <- &Rsp{
					Result:    errno.OK,
					Userdata1: 1,
					Userdata2: now + 60,
					Userasset: nil,
				}
			} else if req.Type == assetCmdUnlock {
				req.RspChannel <- &Rsp{Result: errno.OK}
			}
		case <-exitC:
			goto exit
		}
	}

exit:
	// TODO save all assets into database
	log.Debug("asset loop exit...")
}
