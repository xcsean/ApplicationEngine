package servicefmt

import "testing"

func TestBaseFmt(t *testing.T) {
	app := "app"
	server := "server"
	var id int64 = 123
	division := MakeDivision(app, server, id)
	if division != "app.server.123" {
		t.Errorf("MakeDivision failed: %s", division)
		return
	}

	key := MakeLookupKey(app, server, division)
	if key != "app!server!app.server.123" {
		t.Errorf("MakeLookupKey failed: %s", key)
		return
	}

	app2, server2, id2, err := ParseDivision(division)
	if err != nil {
		t.Errorf("ParseDivision faled: %s", err.Error())
		return
	}
	if app != app2 || server != server2 || id != id2 {
		t.Errorf("ParseDivison failed")
		return
	}
}