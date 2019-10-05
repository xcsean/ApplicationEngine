package packet

import (
	"encoding/json"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
)

// IsPrivateCmdID tell whether the cmdID is a private cmd id or not
func IsPrivateCmdID(cmdID protocol.PacketType) bool {
	return (cmdID >= protocol.Packet_PRIVATE_BEGIN) && (cmdID <= protocol.Packet_PRIVATE_END)
}

// IsPublicCmdID tell whether the cmdID is a public cmd id or not
func IsPublicCmdID(cmdID protocol.PacketType) bool {
	return !IsPrivateCmdID(cmdID)
}

// MakeReservedPacket make a reserved packet
func MakeReservedPacket(cmdID protocol.PacketType, userdata, timestamp uint32, kv map[string]string) []byte {
	rb := protocol.PacketReservedBody{Kv: kv}
	body, _ := json.Marshal(rb)
	return conn.MakeCommonPkt(uint16(cmdID), userdata, timestamp, body)
}

// MakeNotifyClient make a NOTIFY CLIENT packet
func MakeNotifyClient(kv map[string]string) []byte {
	return MakeReservedPacket(protocol.Packet_PRIVATE_NOTIFY_CLIENT, 0, 0, kv)
}

// MakeVerCheck make a VER-CHECK packet
func MakeVerCheck(ver string) []byte {
	return MakeReservedPacket(protocol.Packet_PUBLIC_SESSION_VERCHECK, 0, 0, map[string]string{"ver": ver})
}

// ParseSessionEnterBody parse a SESSION ENTER body
func ParseSessionEnterBody(body []byte) string {
	kv, _ := ParseReservedBody(body)
	return kv["addr"]
}
