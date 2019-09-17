package service

import "testing"

func TestExtFmt(t *testing.T) {
	re := &RedisConnection{
		Endpoint: Endpoint{
			IP: "192.168.0.123",
			Port: "6379",
		},
		AuthPass: "",
	}
	s := re.Addr()
	if s != "192.168.0.123:6379" {
		t.Errorf("RedisConnection Addr() failed: %s", s)
		return
	}

	my := &MysqlConnection{
		Endpoint: Endpoint{
			IP: "192.168.0.123",
			Port: "3306",
		},
		Username: "test",
		Password: "123456",
	}
	s = my.Addr()
	if my.Addr() != "192.168.0.123:3306" {
		t.Errorf("MysqlConnection Addr() faled: %s", s)
		return
	}
}