package conn

import "testing"

func test1(t *testing.T, sessions []uint64, body []byte) {
	var pkt []byte
	var err error
	if len(sessions) == 1 {
		pkt, err = MakeOneSessionPkt(sessions[0], CmdSessionKick, 321, 654, body)
	} else {
		pkt, err = MakeSessionPkt(sessions, CmdSessionKick, 321, 654, body)
	}
	if err != nil {
		t.Errorf("make one session pkt failed: %s", err.Error())
	}
	pLen := LengthOfHeader + 2 + 8*len(sessions) + len(body)
	if len(pkt) != pLen {
		t.Errorf("make one session pkt length=%d, should be=%d", len(pkt), pLen)
	}

	h, num, sessions2, body2 := ParseSessionPkt(pkt)
	hdr := ParseHeader(h)
	if hdr.CmdID != CmdSessionKick {
		t.Error("header cmdID failed")
	}
	if hdr.UserData != 321 {
		t.Error("header userData failed")
	}
	if hdr.Timestamp != 654 {
		t.Error("header timestamp failed")
	}
	if num != uint16(len(sessions)) {
		t.Error("session num failed")
	}
	for i := 0; i < len(sessions2); i++ {
		if sessions2[i] != sessions[i] {
			t.Error("session id failed")
		}
	}
	if len(body2) != len(body) {
		t.Error("body failed")
	}
}

func TestSessionBody(t *testing.T) {
	body1 := make([]byte, 10)

	// multi-sessions
	sessions1 := make([]uint64, 2)
	sessions1[0] = 1234
	sessions1[1] = 5678
	test1(t, sessions1, body1)
	test1(t, sessions1, nil)

	// single-session
	sessions2 := make([]uint64, 1)
	sessions2[0] = 7689
	test1(t, sessions2, body1)
	test1(t, sessions2, nil)

	// null-session
	test1(t, nil, body1)
	test1(t, nil, nil)
}

func TestBodyLength(t *testing.T) {
	sessionID := uint64(12345)
	body1 := make([]byte, LengthOfMaxBody-100)
	_, err := MakeOneSessionPkt(sessionID, CmdKickAll, 0, 0, body1)
	if err != nil {
		t.Errorf("body length=%d shoudn't occur error", len(body1))
	}
	body2 := make([]byte, LengthOfMaxBody+1)
	_, err = MakeOneSessionPkt(sessionID, CmdKickAll, 0, 0, body2)
	if err == nil {
		t.Errorf("body length=%d should occur error", len(body2))
	}
}

func TestPrivateCmd(t *testing.T) {
	b := MakeMasterNot()
	hdr := ParseHeader(b)
	if hdr.CmdID != CmdMasterNot {
		t.Errorf("cmdID failed: %s", "CmdMasterNot")
	}
	if hdr.BodyLen != 0 {
		t.Error("bodyLen failed")
	}
	b = MakeMasterSet()
	hdr = ParseHeader(b)
	if hdr.CmdID != CmdMasterSet {
		t.Errorf("cmdID failed: %s", "CmdMasterSet")
	}
	if hdr.BodyLen != 0 {
		t.Error("bodyLen failed")
	}
	b = MakeMasterYou()
	hdr = ParseHeader(b)
	if hdr.CmdID != CmdMasterYou {
		t.Errorf("cmdID failed: %s", "CmdMasterYou")
	}
	if hdr.BodyLen != 0 {
		t.Error("bodyLen failed")
	}
	b = MakeKickAll()
	hdr = ParseHeader(b)
	if hdr.CmdID != CmdKickAll {
		t.Errorf("cmdID failed: %s", "CmdKickAll")
	}
	if hdr.BodyLen != 0 {
		t.Error("bodyLen failed")
	}
	b, err := MakeBroadcastAll(make([]byte, 1024))
	if err != nil {
		t.Errorf("make broadcast all failed: %s", err.Error())
	}
	hdr = ParseHeader(b)
	if hdr.CmdID != CmdBroadcastAll {
		t.Errorf("cmdID failed: %s", "CmdBroadcastAll")
	}
	if hdr.BodyLen != 1024 {
		t.Error("broadcast all bodyLen failed")
	}
}
