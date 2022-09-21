package main

import (
	"github.com/kkakoz/gim/pkg/conf"
	"github.com/kkakoz/gim/pkg/gox"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	conf.InitConfigFile("../.././configs/conf.yaml")
	gox.Go(func() {
		demo := ClientDemo{}
		demo.Start("user_1", "ws", "ws://localhost:9100/")
	})

	for {
		time.Sleep(10 * time.Second)
	}
}
