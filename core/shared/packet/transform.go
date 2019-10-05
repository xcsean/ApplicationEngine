package packet

import (
	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
)

// TransformRPCToSocket transform the RPC defined struct to an array used by socket
func TransformRPCToSocket(pkt *protocol.SessionPacket) ([]byte, error) {
	return conn.MakeSessionPkt(pkt.Sessions, uint16(pkt.Common.CmdId), pkt.Common.UserData, pkt.Common.Timestamp, []byte(pkt.Common.Body))
}

// TransformSocketToRPC transform the array used by socket to RPC defined struct
func TransformSocketToRPC(hdr, body []byte) *protocol.SessionPacket {
	header := conn.ParseHeader(hdr)
	_, sessionIDs, innerBody := conn.ParseSessionBody(body)
	pkt := &protocol.SessionPacket{
		Sessions: sessionIDs,
		Common: &protocol.Packet{
			CmdId:     int32(header.CmdID),
			UserData:  header.UserData,
			Timestamp: header.Timestamp,
			Body:      string(innerBody[:]),
		},
	}
	return pkt
}
