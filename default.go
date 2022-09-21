package gim

import (
	"github.com/pkg/errors"
	"time"
)

const DefaultLoginWait = 60 * time.Second

const DefaultReadWait = 60 * time.Second

type DefaultAcceptor struct {
}

func (d DefaultAcceptor) Accept(conn Conn, duration time.Duration) (string, error) {
	// 1. 读取：客户端发送的鉴权数据包
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}
	// 2. 解析：数据包内容就是userId
	userID := string(frame.GetPayload())
	// 3. 鉴权：这里只是为了示例做一个fake验证，非空
	if userID == "" {
		return "", errors.New("user id is invalid")
	}
	return userID, nil
}
