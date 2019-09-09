package log

import "testing"

func TestHourRotate(t *testing.T) {
	rot := NewRotatorByHour("./", "")
	if rot == nil {
		t.Errorf("new rotator failed")
		return
	}
	defer rot.Close()
	n, err := rot.Write([]byte("abc\n"))
	if err != nil || n != 4 {
		t.Errorf("write failed: %s", err.Error())
		return
	}
	prefix := rot.GetFilePrefix()
	ext := rot.GetFileExt()
	typ := rot.GetRotateType()
	t.Logf("prefix=%s, ext=%s, type=%v", prefix, ext, typ)
}
