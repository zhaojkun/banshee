package bus

import "testing"

func TestClient(t *testing.T) {
	num := 0
	handler := func() {
		num++
	}
	Subscribe("event", handler)
	Publish("event")
	if num != 1 {
		t.Fail()
	}
	UnSubscribe("event", handler)
	Publish("event")
	if num != 1 {
		t.Fail()
	}
}
