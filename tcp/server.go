package tcp

import (
	"context"
	"fmt"
	"github.com/kkakoz/gim"
	"github.com/kkakoz/gim/pkg/gox"
	"github.com/kkakoz/gim/pkg/logger"
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
	options *gim.ServerOptions
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
	listen, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}

	for {
		log.Info("started\n")
		conn, err := listen.Accept()
		if err != nil {
			log.Error("accept conn err:" + err.Error())
		}
		gox.Go(func() {
			s.conn = NewConn(conn)
			id, err := s.Accept(s.conn, s.options.LoginWait)
			if err != nil {
				logger.Error("acceptor err:" + err.Error())
			}

			if err != nil {
				_ = s.conn.WriteFrame(gim.OpClose, []byte(err.Error()))
				conn.Close()
			}
			if _, ok := s.Get(id); ok {
				log.Warn(fmt.Sprintf("channel %s existed", id))
				_ = s.conn.WriteFrame(gim.OpClose, []byte("channelId is repeated"))
				conn.Close()
			}
			// step 4
			channel := gim.NewChannel(id, s.conn)
			channel.SetWriteWait(s.options.WriteWait)
			channel.SetReadWait(s.options.ReadWait)
			s.Add(channel)

			err = channel.ReadLoop(s.MessageListener)
			if err != nil {
				log.Info("read loop err:" + err.Error())
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

func NewServer(listen string, service gim.ServiceRegistration, optsfunc ...gim.ServerOptionsFunc) gim.Server {
	serverOption := gim.NewServerOption()
	for _, opt := range optsfunc {
		opt(serverOption)
	}
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		options:             serverOption,
	}
}
