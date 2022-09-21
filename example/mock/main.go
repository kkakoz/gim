package main

import (
	"github.com/kkakoz/gim/pkg/conf"
	"github.com/kkakoz/gim/pkg/gox"
	"time"
)

func main() {
	conf.InitConfigFile("./configs/conf.yaml")
	gox.Go(func() {
		server := ServerDemo{}
		server.Start("1", "ws", ":9100")
	})

	for {
		time.Sleep(10 * time.Second)
	}
}
