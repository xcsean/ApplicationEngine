package packet

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/xcsean/ApplicationEngine/core/protocol"
)

// MakeReservedBody make a reserved body
func MakeReservedBody(kv map[string]string) []byte {
	rb := protocol.PacketReservedBody{Kv: kv}
	body, _ := json.Marshal(rb)
	return body
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

// ParseSessionVerCheckBody parse a VER-CHECK body
func ParseSessionVerCheckBody(body []byte) (string, error) {
	kv, err := ParseReservedBody(body)
	if err != nil {
		return "", err
	}
	ver, ok := kv["ver"]
	if !ok {
		return "", fmt.Errorf("key 'ver' not found")
	}
	return ver, nil
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
