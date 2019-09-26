package main

import "github.com/xcsean/ApplicationEngine/core/shared/log"

// all dispatchXXX functions run in the main routine context!!!

func dispatchRPC(cmd *innerCmd) bool {
	cmdID := cmd.getID()
	switch cmdID {
	case innerCmdRegisterVM:
		log.Debug("register vm")
	case innerCmdUnregisterVM:
		log.Debug("unregister vm")
	}
	return false
}
