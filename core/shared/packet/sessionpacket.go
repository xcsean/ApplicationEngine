package packet

import (
	"encoding/json"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
)

// MakeReservedSessionPacket make a reserved session-packet
func MakeReservedSessionPacket(sessionIDs []uint64, cmdID protocol.PacketType, userdata, timestamp uint32, kv map[string]string) ([]byte, error) {
	rb := protocol.PacketReservedBody{Kv: kv}
	body, _ := json.Marshal(rb)
	return conn.MakeSessionPkt(sessionIDs, uint16(cmdID), userdata, timestamp, body)
}

// ParseReservedSessionBody parse a reserved session-body
func ParseReservedSessionBody(body []byte) ([]uint64, map[string]string, error) {
	_, sessionIDs, rbBody := conn.ParseSessionBody(body)

	var rb protocol.PacketReservedBody
	err := json.Unmarshal(rbBody, &rb)
	if err != nil {
		return sessionIDs, nil, err
	}
	return sessionIDs, rb.Kv, nil
}

// MakeSessionEnter make a SESSION ENTER packet
func MakeSessionEnter(sessionIDs []uint64, addr string) ([]byte, error) {
	return MakeReservedSessionPacket(sessionIDs, protocol.Packet_PRIVATE_SESSION_ENTER, 0, 0, map[string]string{"addr": addr})
}

// MakeSessionLeave make a SESSION LEAVE packet
func MakeSessionLeave(sessionIDs []uint64) ([]byte, error) {
	return MakeReservedSessionPacket(sessionIDs, protocol.Packet_PRIVATE_SESSION_LEAVE, 0, 0, nil)
}

// MakeSessionRoute make a SESSION ROUTE packet
func MakeSessionRoute(sessionIDs []uint64, addr string) ([]byte, error) {
	return MakeReservedSessionPacket(sessionIDs, protocol.Packet_PRIVATE_SESSION_ROUTE, 0, 0, map[string]string{"addr": addr})
}

// ParseSessionRouteBody parse a SESSION ROUTE body
func ParseSessionRouteBody(body []byte) ([]uint64, string) {
	sessionIDs, kv, _ := ParseReservedSessionBody(body)
	return sessionIDs, kv["addr"]
}

// MakeSessionKick make a SESSION KICK packet
func MakeSessionKick(sessionIDs []uint64) ([]byte, error) {
	return MakeReservedSessionPacket(sessionIDs, protocol.Packet_PRIVATE_SESSION_KICK, 0, 0, nil)
}

// MakeMasterSet make a MASTER SET packet
func MakeMasterSet() ([]byte, error) {
	return MakeReservedSessionPacket(nil, protocol.Packet_PRIVATE_MASTER_SET, 0, 0, nil)
}

// MakeMasterYou make a MASTER YOU packet
func MakeMasterYou() ([]byte, error) {
	return MakeReservedSessionPacket(nil, protocol.Packet_PRIVATE_MASTER_YOU, 0, 0, nil)
}

// MakeMasterNot make a MASTER NOT packet
func MakeMasterNot() ([]byte, error) {
	return MakeReservedSessionPacket(nil, protocol.Packet_PRIVATE_MASTER_NOT, 0, 0, nil)
}