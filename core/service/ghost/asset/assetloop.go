package asset

import (
	"time"

	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

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
