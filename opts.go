package gim

import (
	"context"
	"time"
)

type ServerOptions struct {
	LoginWait time.Duration
	ReadWait  time.Duration
	WriteWait time.Duration
}

func NewServerOption() *ServerOptions {
	return &ServerOptions{
		LoginWait: DefaultLoginWait,
		ReadWait:  DefaultReadWait,
		WriteWait: DefaultReadWait,
	}
}

type ServerOptionsFunc func(options *ServerOptions)

func WithServerLoginWait(duration time.Duration) ServerOptionsFunc {
	return func(options *ServerOptions) {
		options.LoginWait = duration
	}
}

func WithServerRWWait(duration time.Duration) ServerOptionsFunc {
	return func(options *ServerOptions) {
		options.ReadWait = duration
		options.WriteWait = duration
	}
}

type ClientOptions struct {
	Heartbeat time.Duration
	ReadWait  time.Duration
	WriteWait time.Duration
}

func NewClientOptions() *ClientOptions {
	return &ClientOptions{Heartbeat: time.Second * 60, ReadWait: time.Second * 60, WriteWait: time.Second * 60}
}

type ClientOptionFunc func(*ClientOptions)

func WithClientHeartbeat(duration time.Duration) ClientOptionFunc {
	return func(options *ClientOptions) {
		options.Heartbeat = duration
	}
}

type ChannelOptions struct {
	ctx context.Context
}

type channelOptionFunc func(opt *ChannelOptions)

func WithChannelCtx(ctx context.Context) channelOptionFunc {
	return func(opt *ChannelOptions) {
		opt.ctx = ctx
	}
}

func newChannelOptions() *ChannelOptions {
	return &ChannelOptions{ctx: context.Background()}
}
