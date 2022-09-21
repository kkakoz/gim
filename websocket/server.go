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

type ServerOptions struct {
	LoginWait time.Duration
	ReadWait  time.Duration
	WriteWait time.Duration
}

func newServerOption() *ServerOptions {
	return &ServerOptions{
		LoginWait: gim.DefaultLoginWait,
		ReadWait:  gim.DefaultReadWait,
		WriteWait: gim.DefaultReadWait,
	}
}

type serverOptionsFunc func(options *ServerOptions)

func WithServerLoginWait(duration time.Duration) serverOptionsFunc {
	return func(options *ServerOptions) {
		options.LoginWait = duration
	}
}

func WithServerRWWait(duration time.Duration) serverOptionsFunc {
	return func(options *ServerOptions) {
		options.ReadWait = duration
		options.WriteWait = duration
	}
}

// DefaultServer is a websocket implement of the Server
type DefaultServer struct {
	listen string
	gim.ServiceRegistration
	gim.IChannelMap
	gim.Acceptor
	gim.MessageListener
	gim.StateListener
	once    sync.Once
	options *ServerOptions
}

func (s *DefaultServer) SetAcceptor(acceptor gim.Acceptor) {
	s.Acceptor = acceptor
}

func (s *DefaultServer) SetMessageListener(listener gim.MessageListener) {
	s.MessageListener = listener
}

func (s *DefaultServer) SetStateListener(listener gim.StateListener) {
	s.StateListener = listener
}

func (s *DefaultServer) SetReadWait(duration time.Duration) {
	s.options.ReadWait = duration
}

func (s *DefaultServer) SetChannelMap(channelMap gim.IChannelMap) {
	s.IChannelMap = channelMap
}

func (s *DefaultServer) Push(id string, data []byte) error {
	ch, ok := s.IChannelMap.Get(id)
	if !ok {
		return errors.New("channel no found")
	}
	return ch.Push(data)
}

func (s *DefaultServer) Shutdown(ctx context.Context) error {
	return nil
}

func NewDefaultServer(listen string, service gim.ServiceRegistration, optsfunc ...serverOptionsFunc) gim.Server {
	serverOption := newServerOption()
	for _, opt := range optsfunc {
		opt(serverOption)
	}
	return &DefaultServer{
		listen:              listen,
		ServiceRegistration: service,
		options:             serverOption,
	}
}

// Start server
func (s *DefaultServer) Start() error {
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

	})
	mux.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("test handler"))
	})
	log.Info("started\n")
	return http.ListenAndServe(s.listen, mux)
}
