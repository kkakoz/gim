package gim

import (
	"fmt"
	"github.com/kkakoz/gim/pkg/gox"
	"github.com/kkakoz/gim/pkg/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

type IChannel interface {
	Conn // <-- 这个就是前面说的kim.Conn
	Agent
	Close() error // <-- 重写net.Conn中的Close方法
	ReadLoop(lst MessageListener) error
	SetWriteWait(time.Duration)
	SetReadWait(time.Duration)
}

// Channel is a websocket implement of channel
type Channel struct {
	sync.Mutex
	id string
	Conn
	writechan chan []byte
	once      sync.Once
	writeWait time.Duration
	readWait  time.Duration
	state     int32 // 0 init 1 start 2 close
}

func (ch *Channel) SetWriteWait(duration time.Duration) {
	ch.writeWait = duration
}

func (ch *Channel) SetReadWait(duration time.Duration) {
	ch.readWait = duration
}

func (ch *Channel) ID() string {
	return ch.id
}

func (ch *Channel) Close() error {
	if !atomic.CompareAndSwapInt32(&ch.state, 1, 2) {
		return fmt.Errorf("channel has started")
	}
	close(ch.writechan)
	return nil
}

func NewChannel(id string, conn Conn, options ...channelOptionFunc) IChannel {
	channelOpt := newChannelOptions()
	for _, opt := range options {
		opt(channelOpt)
	}
	ch := &Channel{
		id:        id,
		Conn:      conn,
		writechan: make(chan []byte, 5),
		writeWait: time.Second * 10, //default value
	}
	gox.Go(func() {
		log := logger.WithFields(zap.String("struct", "Channel"), zap.String("func", "writeLoop"), zap.String("id", ch.id))
		err := ch.writeLoop()
		if err != nil {
			log.Error(err.Error())
		}
		logger.Debug("write loop closed")
	})
	return ch
}

func (ch *Channel) writeLoop() error {
	for payload := range ch.writechan {
		err := ch.WriteFrame(OpBinary, payload)
		if err != nil {
			return errors.New(fmt.Sprintf("wirte %s frame err:%s", string(payload), err.Error()))
		}
		chanlen := len(ch.writechan)
		for i := 0; i < chanlen; i++ {
			payload = <-ch.writechan
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return errors.New(fmt.Sprintf("wirte %s frame err:%s", string(payload), err.Error()))
			}
		}
		err = ch.Conn.Flush()
		if err != nil {
			return errors.New("flush frame err:" + err.Error())
		}
	}
	return nil
}

func (ch *Channel) Push(payload []byte) error {
	if atomic.LoadInt32(&ch.state) != 1 {
		return fmt.Errorf("channel %s has closed", ch.id)
	}
	// 异步写
	ch.writechan <- payload
	return nil
}

// overwrite Conn
func (ch *Channel) WriteFrame(code OpCode, payload []byte) error {
	_ = ch.Conn.SetWriteDeadline(time.Now().Add(ch.writeWait))
	return ch.Conn.WriteFrame(code, payload)
}

func (ch *Channel) ReadLoop(lst MessageListener) error {
	if !atomic.CompareAndSwapInt32(&ch.state, 0, 1) {
		return fmt.Errorf("channel has started")
	}
	log := logger.WithFields(zap.String("struct", "Channel"), zap.String("func", "Readloop"), zap.String("id", ch.id))
	for {
		_ = ch.SetReadDeadline(time.Now().Add(ch.readWait))

		frame, err := ch.ReadFrame()
		if err != nil {
			return err
		}
		if frame.GetOpCode() == OpClose {
			return errors.New("remote side close the channel")
		}
		if frame.GetOpCode() == OpPing {
			log.Info("recv a ping; resp with a pong")
			_ = ch.WriteFrame(OpPong, nil)
			continue
		}
		payload := frame.GetPayload()
		if len(payload) == 0 {
			continue
		}
		gox.Go(func() {
			lst.Receive(ch, payload)
		})
	}
}
