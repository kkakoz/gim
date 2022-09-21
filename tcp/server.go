package tcp

import (
	"context"
	"fmt"
	"github.com/kkakoz/gim"
	"github.com/kkakoz/gim/pkg/gox"
	"github.com/kkakoz/gim/pkg/logger"
	"github.com/kkakoz/gim/websocket"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

type Server struct {
	listen string
	gim.ServiceRegistration
	gim.IChannelMap
	gim.Acceptor
	gim.MessageListener
	gim.StateListener
	conn    gim.Conn
	once    sync.Once
	options *websocket.ServerOptions
}

func (s *Server) SetAcceptor(acceptor gim.Acceptor) {
	s.Acceptor = acceptor
}

func (s *Server) SetMessageListener(listener gim.MessageListener) {
	s.MessageListener = listener
}

func (s *Server) SetStateListener(listener gim.StateListener) {
	s.StateListener = listener
}

func (s *Server) SetReadWait(duration time.Duration) {
	s.options.ReadWait = duration
}

func (s *Server) SetChannelMap(channelMap gim.IChannelMap) {
	s.IChannelMap = channelMap
}

func (s *Server) Start() error {
	log := logger.WithFields(zap.String("module", "tcp.server"), zap.String("listen", s.listen), zap.String("id", s.ServiceID()))
	listen, err := net.Listen("tpc", s.listen)
	if err != nil {
		return err
	}

	for {
		log.Info("started\n")
		conn, err := listen.Accept()
		if err != nil {
			log.Error("accept conn err:" + err.Error())
		}
		tcpConn := NewConn(conn)
		id, err := s.Acceptor.Accept(s.conn, s.options.LoginWait)
		if err != nil {
			return errors.New("acceptor err:" + err.Error())
		}

		if err != nil {
			_ = tcpConn.WriteFrame(gim.OpClose, []byte(err.Error()))
			conn.Close()
			continue
		}
		if _, ok := s.Get(id); ok {
			log.Warn(fmt.Sprintf("channel %s existed", id))
			_ = tcpConn.WriteFrame(gim.OpClose, []byte("channelId is repeated"))
			conn.Close()
			continue
		}
		// step 4
		channel := gim.NewChannel(id, tcpConn)
		channel.SetWriteWait(s.options.WriteWait)
		channel.SetReadWait(s.options.ReadWait)
		s.Add(channel)

		gox.Go(func() {
			err := channel.ReadLoop(s.MessageListener)
			if err != nil {
				log.Info(err.Error())
			}
			// step 6
			s.Remove(channel.ID())
			err = s.Disconnect(channel.ID())
			if err != nil {
				log.Warn(err.Error())
			}
			channel.Close()
		})
	}

}

func (s *Server) Push(id string, data []byte) error {
	channel, ok := s.IChannelMap.Get(id)
	if ok {
		return channel.Push(data)
	}
	return errors.New("channel not found")
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}
