package id

import (
	"testing"
	"time"
)

func TestSessionID(t *testing.T) {
	connID := uint16(1)
	settings := Settings{
		StartTime:      time.Now(),
		MachineID:      func() (uint16, error) { return connID, nil },
		CheckMachineID: func(uint16) bool { return true },
	}
	sf := NewSnowflake(settings)
	id, err := sf.NextID()
	if err != nil {
		t.Errorf("NextID failed: %s", err.Error())
	}
	t.Logf("ID: %d", id)
	time.Sleep(time.Second * 1)
	id, err = sf.NextID()
	if err != nil {
		t.Errorf("NextID failed: %s", err.Error())
	}
	t.Logf("ID: %d", id)
	m := Decompose(id)
	t.Logf("id=%v", m)
}
