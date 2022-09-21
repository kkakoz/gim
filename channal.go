package gim

import (
	"github.com/kkakoz/gim/pkg/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Channel interface {
	Conn // <-- 这个就是前面说的kim.Conn
	Agent
	Close() error // <-- 重写net.Conn中的Close方法
	ReadLoop(lst MessageListener) error
	SetWriteWait(time.Duration)
	SetReadWait(time.Duration)
}

// ChannelImpl is a websocket implement of channel
type ChannelImpl struct {
	sync.Mutex
	id string
	Conn
	writechan chan []byte
	once      sync.Once
	writeWait time.Duration
	closed    *Event
	readWait  time.Duration
}

func (ch *ChannelImpl) SetWriteWait(duration time.Duration) {
	ch.writeWait = duration
}

func (ch *ChannelImpl) SetReadWait(duration time.Duration) {
	ch.readWait = duration
}

func (ch *ChannelImpl) ID() string {
	return ch.id
}

func NewChannel(id string, conn Conn) Channel {
	log := logger.WithFields(zap.String("module", "tpc_channel"), zap.String("id", id))
	ch := &ChannelImpl{
		id:        id,
		Conn:      conn,
		writechan: make(chan []byte, 5),
		closed:    NewEvent(),
		writeWait: time.Second * 10, //default value
	}
	go func() {
		err := ch.writeLoop()
		if err != nil {
			log.Info(err.Error())
		}
	}()
	return ch
}

func (ch *ChannelImpl) writeLoop() error {
	for {
		select {
		case payload := <-ch.writechan:
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return err
			}
			// 批量写
			chanlen := len(ch.writechan)
			for i := 0; i < chanlen; i++ {
				payload = <-ch.writechan
				err := ch.WriteFrame(OpBinary, payload)
				if err != nil {
					return err
				}
			}
			err = ch.Conn.Flush()
			if err != nil {
				return err
			}
		case <-ch.closed.Done():
			return nil
		}
	}
}

func (ch *ChannelImpl) Push(payload []byte) error {
	if ch.closed.HasFired() {
		return errors.New("channel has closed")
	}
	// 异步写
	ch.writechan <- payload
	return nil
}

// overwrite Conn
func (ch *ChannelImpl) WriteFrame(code OpCode, payload []byte) error {
	_ = ch.Conn.SetWriteDeadline(time.Now().Add(ch.writeWait))
	return ch.Conn.WriteFrame(code, payload)
}

func (ch *ChannelImpl) ReadLoop(lst MessageListener) error {
	ch.Lock()
	defer ch.Unlock()
	log := logger.WithFields(zap.String("struct", "ChannelImpl"), zap.String("func", "Readloop"), zap.String("id", ch.id))
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
		go lst.Receive(ch, payload)
	}
}
