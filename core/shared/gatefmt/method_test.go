package gatefmt

import "testing"

func TestCmd(t *testing.T) {
	cmdID := uint16(1)
	if !IsPublicCmd(cmdID) {
		t.Errorf("%d should be a public cmd", cmdID)
	}
	cmdID = uint16(65001)
	if !IsPrivateCmd(cmdID) {
		t.Errorf("%d should be a private cmd", cmdID)
	}
}

func TestHeader(t *testing.T) {
	b := MakeHeader(0, CmdMasterSet, 1, 2)
	hdr := ParseHeader(b)
	if hdr.BodyLen != 0 {
		t.Error("bodyLen failed")
	}
	if hdr.CmdID != CmdMasterSet {
		t.Error("cmdID failed")
	}
	if hdr.UserData != 1 {
		t.Error("userData failed")
	}
	if hdr.Timestamp != 2 {
		t.Error("timestamp failed")
	}
}
