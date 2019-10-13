package asset

import "github.com/xcsean/ApplicationEngine/core/shared/log"

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
