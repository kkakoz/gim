package main

import (
	"context"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/kkakoz/gim"
	"github.com/kkakoz/gim/pkg/logger"
	"github.com/kkakoz/gim/websocket"
	"net"
	"time"
)

type ClientDemo struct {
}

type WebsocketDialer struct {
}

func (w *WebsocketDialer) DialAndHandshake(ctx gim.DialerContext) (net.Conn, error) {
	logger.Info("start ws dial: " + ctx.Address)
	// 1 调用ws.Dial拨号
	ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), ctx.Timeout)
	defer cancel()

	conn, _, _, err := ws.Dial(ctxWithTimeout, ctx.Address)
	if err != nil {
		return nil, err
	}
	// 2. 发送用户认证信息，示例就是userid
	err = wsutil.WriteClientBinary(conn, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}
	// 3. return conn
	return conn, nil
}

//入口方法
func (c *ClientDemo) Start(userID, protocol, addr string) {
	var cli gim.Client

	// step1: 初始化客户端
	if protocol == "ws" {
		cli = websocket.NewClient(userID, "client", websocket.WithClientHeartbeat(time.Second*30))
		// set dialer
		cli.SetDialer(&WebsocketDialer{})
	} /*else if protocol == "tcp" {
		cli = tcp.NewClient("test1", "client", tcp.ClientOptions{})
		cli.SetDialer(&TCPDialer{})
	}*/

	// step2: 建立连接

	err := cli.Connect(addr)
	if err != nil {
		logger.Fatal("client conn" + err.Error())
	}
	count := 10
	go func() {
		// step3: 发送消息然后退出
		for i := 0; i < count; i++ {
			err := cli.Send([]byte("hello"))
			if err != nil {
				logger.Error(err.Error())
				return
			}
			time.Sleep(time.Second)
		}
	}()

	// step4: 接收消息
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Error(err.Error())
			break
		}
		if frame.GetOpCode() != gim.OpBinary {
			continue
		}
		recv++
		logger.Warn(fmt.Sprintf("%s receive message [%s]", cli.ID(), frame.GetPayload()))
		if recv == count { // 接收完消息
			break
		}
	}
	//退出
	cli.Close()
}

type ClientHandler struct {
}

// Receive default listener
func (h *ClientHandler) Receive(ag gim.Agent, payload []byte) {
	logger.Warn(fmt.Sprintf("%s receive message [%s]", ag.ID(), string(payload)))
}

// Disconnect default listener
func (h *ClientHandler) Disconnect(id string) error {
	logger.Warn(fmt.Sprintf("disconnect %s", id))
	return nil
}
