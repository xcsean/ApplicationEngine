package main

var (
	vmmgr *vmMgr
)

func init() {
	vmmgr = newVMMgr()
}

// all dispatchXXX functions run in the main routine context!!!

func dispatchRPC(cmd *innerCmd) bool {
	cmdID := cmd.getID()
	switch cmdID {
	case innerCmdRegisterVM:
		division, version, rspChannel := cmd.getRPCReq()
		result := vmmgr.addVM(&vmEntity{division: division, version: version})
		rspChannel <- newRPCRsp(innerCmdRegisterVM, result)
	case innerCmdUnregisterVM:
		division, _, rspChannel := cmd.getRPCReq()
		result := vmmgr.delVM(division)
		rspChannel <- newRPCRsp(innerCmdUnregisterVM, result)
	}
	return false
}
