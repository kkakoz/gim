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

func TestServer(t *testing.T) {
	conf.InitConfigFile("../../configs/conf.yaml")
	gox.Go(func() {
		server := ServerDemo{}
		server.Start("1", "ws", ":9100")
	})

	for {
		time.Sleep(10 * time.Second)
	}
}

func TestTCPClient(t *testing.T) {
	conf.InitConfigFile("../../configs/conf.yaml")
	gox.Go(func() {
		demo := ClientDemo{}
		demo.Start("user_1", "tcp", "localhost:9200")
	})

	for {
		time.Sleep(10 * time.Second)
	}
}

func TestTCPServer(t *testing.T) {
	conf.InitConfigFile("../../configs/conf.yaml")
	gox.Go(func() {
		server := ServerDemo{}
		server.Start("1", "tcp", ":9200")
	})

	for {
		time.Sleep(10 * time.Second)
	}
}
