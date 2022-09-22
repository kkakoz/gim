package websocket

import (
	"context"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/kkakoz/gim"
	"github.com/kkakoz/gim/pkg/gox"
	"github.com/kkakoz/gim/pkg/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

// Server is a websocket implement of the Server
type Server struct {
	listen string
	gim.ServiceRegistration
	gim.IChannelMap
	gim.Acceptor
	gim.MessageListener
	gim.StateListener
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

func (s *Server) Push(id string, data []byte) error {
	ch, ok := s.IChannelMap.Get(id)
	if !ok {
		return errors.New("channel no found")
	}
	return ch.Push(data)
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

// Start server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	log := logger.WithFields(zap.String("module", "ws.server"), zap.String("listen", s.listen), zap.String("id", s.ServiceID()))
	if s.Acceptor == nil {
		s.Acceptor = new(gim.DefaultAcceptor)
	}
	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}
	// 连接管理器
	if s.IChannelMap == nil {
		s.IChannelMap = gim.NewChannels()
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// step 1
		rawconn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			logger.Error("upgrade http err:" + err.Error())
			//resp(w, http.StatusBadRequest, err.Error())
			return
		}

		// step 2 包装conn
		conn := NewConn(rawconn)

		// step 3
		id, err := s.Accept(conn, s.options.LoginWait)
		if err != nil {
			_ = conn.WriteFrame(gim.OpClose, []byte(err.Error()))
			conn.Close()
			return
		}
		if _, ok := s.Get(id); ok {
			log.Warn(fmt.Sprintf("channel %s existed", id))
			_ = conn.WriteFrame(gim.OpClose, []byte("channelId is repeated"))
			conn.Close()
			return
		}
		// step 4
		channel := gim.NewChannel(id, conn)
		channel.SetWriteWait(s.options.WriteWait)
		channel.SetReadWait(s.options.ReadWait)
		s.Add(channel)

		gox.Go(func() {
			err := channel.ReadLoop(s.MessageListener)
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

	})
	mux.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("test handler"))
	})
	log.Info("started\n")
	return http.ListenAndServe(s.listen, mux)
}
