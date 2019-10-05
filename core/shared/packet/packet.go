package packet

import (
	"encoding/json"
	"fmt"
	"strconv"

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

// ParseReservedBody parse a reserved body
func ParseReservedBody(body []byte) (map[string]string, error) {
	var rb protocol.PacketReservedBody
	err := json.Unmarshal(body, &rb)
	if err != nil {
		return nil, err
	}
	return rb.Kv, nil
}

// MakeNotifyClient make a NOTIFY CLIENT packet
func MakeNotifyClient(kv map[string]string) []byte {
	return MakeReservedPacket(protocol.Packet_PRIVATE_NOTIFY_CLIENT, 0, 0, kv)
}

// MakeVerCheck make a VER-CHECK packet
func MakeVerCheck(ver string) []byte {
	return MakeReservedPacket(protocol.Packet_PUBLIC_SESSION_VERCHECK, 0, 0, map[string]string{"ver": ver})
}

// ParseVerReplyBody parse a VER-REPLY body
func ParseVerReplyBody(body []byte) (int32, error) {
	kv, err := ParseReservedBody(body)
	if err != nil {
		return -1, err
	}
	s, ok := kv["result"]
	if !ok {
		return -1, fmt.Errorf("key 'result' not found")
	}
	result, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1, err
	}
	return int32(result), nil
}
