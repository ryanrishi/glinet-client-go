package glinet_test

import (
	"context"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestQoSSpeedLimit(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	const mac = "02:00:00:00:00:16"

	set := sentParams(srv, "qos", "set_speed_limit_rule")
	if err := c.QoS.SetSpeedLimit(ctx, glinet.SpeedLimit{MAC: mac, Download: 1024, Upload: 0}); err != nil {
		t.Fatalf("SetSpeedLimit: %v", err)
	}
	checkClientWrite(t, *set, `{"mac":"02:00:00:00:00:16","download":1024,"upload":0}`)
	if got := clientByMAC(t, c, mac); got.LimitRx != 1024 || got.LimitTx != 0 {
		t.Errorf("client = %+v, want limit_rx 1024 and limit_tx 0", got)
	}

	remove := sentParams(srv, "qos", "remove_speed_limit_rule")
	if err := c.QoS.RemoveSpeedLimit(ctx, mac); err != nil {
		t.Fatalf("RemoveSpeedLimit: %v", err)
	}
	checkClientWrite(t, *remove, `{"mac":"02:00:00:00:00:16"}`)
	if got := clientByMAC(t, c, mac); got.LimitRx != 0 || got.LimitTx != 0 {
		t.Errorf("client = %+v, want no limits", got)
	}
}
