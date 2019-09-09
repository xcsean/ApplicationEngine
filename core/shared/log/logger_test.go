package log

import "testing"

func TestLogger(t *testing.T) {
	err := SetupMainLogger("./", "", "debug")
	if err != nil {
		t.Errorf("setup logger failed: %s", err.Error())
		return
	}

	Debug("abc=%d", 2)
	Info("efg=%d", 3)
	ChangeLevel("info")
	Debug("you can't see this message")
}
