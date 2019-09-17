package conn

import (
	"encoding/binary"
	"net"

	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/streambuffer"
)

// HandleStream stream handler
func HandleStream(conn net.Conn, handler func(net.Conn, []byte, []byte)) error {
	buffer := streambuffer.New(conn, 2*LengthOfMaxPacket)
	for {
		_, err := buffer.Read()
		if err != nil {
			log.Debug("read conn=%s failed: %s", conn.RemoteAddr().String(), err.Error())
			return err
		}
		shift := false
		for {
			hdr, err := buffer.Peek(LengthOfHeader)
			if err != nil {
				break
			}
			// get body length
			currLen := buffer.Length()
			bodyLen := binary.BigEndian.Uint16(hdr[BodyLenStart:BodyLenEnd])
			meetLen := int(LengthOfHeader + bodyLen)
			if currLen >= meetLen {
				// get the body
				body := buffer.Fetch(LengthOfHeader, int(bodyLen))
				shift = true
				if bodyLen <= LengthOfMaxBody {
					// trigger the custom handler
					handler(conn, hdr, body)
				} else {
					cmdID := binary.BigEndian.Uint16(hdr[CmdIDStart:CmdIDEnd])
					log.Error("body=%d above max body, cmd=%d, discarded!", bodyLen, cmdID)
				}
				// continue to peek
				continue
			}
		}
		if shift {
			buffer.Shift()
		}
	}
}
