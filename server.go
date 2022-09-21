package gim

import (
	"context"
	"net"
	"time"
)

type Server interface {
	SetAcceptor(Acceptor)
	SetMessageListener(MessageListener)
	SetStateListener(StateListener)
	SetReadWait(time.Duration)
	SetChannelMap(IChannelMap)

	Start() error
	Push(string, []byte) error
	Shutdown(context.Context) error
}

// Acceptor 握手相关操作
type Acceptor interface {
	Accept(Conn, time.Duration) (string, error)
}

// StateListener 上报断开连接
type StateListener interface {
	Disconnect(string) error
}

// MessageListener 消息监听器
type MessageListener interface {
	Receive(Agent, []byte)
}

type Agent interface {
	ID() string
	Push([]byte) error
}

// Frame websocket包
type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}

// Conn Connection
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}

type OpCode byte

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

// 定义了基础服务的抽象接口
type Service interface {
	ServiceID() string
	ServiceName() string
	GetMeta() map[string]string
}

// 定义服务注册的抽象接口
type ServiceRegistration interface {
	Service
	PublicAddress() string
	PublicPort() int
	DialURL() string
	GetTags() []string
	GetProtocol() string
	GetNamespace() string
	String() string
}
