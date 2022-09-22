package tcp

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/kkakoz/gim"
	"github.com/kkakoz/gim/pkg/logger"
	"github.com/pkg/errors"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type client struct {
	sync.Mutex
	once    sync.Once
	id      string
	name    string
	Timeout time.Duration
	state   int32 // 0未连接 1连接

	conn gim.Conn
	gim.Dialer
	options *gim.ClientOptions
}

func (c *client) GetMeta() map[string]string {
	return map[string]string{}
}

func (c *client) ServiceID() string {
	return c.id
}

func (c *client) ServiceName() string {
	return c.name
}

func (c *client) Connect(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return fmt.Errorf("client has connected")
	}
	conn, err := c.Dialer.DialAndHandshake(gim.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: time.Second,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}
	c.conn = NewConn(conn)

	if c.options.Heartbeat > 0 {
		go func() {
			err := c.heartbeatLoop(c.conn)
			if err != nil {
				logger.Error("heartbealoop stopped: " + err.Error())
			}
		}()
	}
	return nil
}

func (c *client) SetDialer(dialer gim.Dialer) {
	c.Dialer = dialer
}

func (c *client) Send(bytes []byte) error {
	return c.conn.WriteFrame(gim.OpBinary, bytes)
}

func (c *client) Read() (gim.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}
	if c.options.ReadWait > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait))
	}
	frame, err := c.conn.ReadFrame()
	if err != nil {
		return nil, err
	}
	if frame.GetOpCode() == gim.OpClose {
		return nil, errors.New("remote side close the channel")
	}
	return frame, nil
}

func (c *client) Close() {
	_ = c.conn.Close()
}

func (c *client) heartbeatLoop(conn gim.Conn) error {
	ticker := time.NewTicker(time.Second * 30)
	for range ticker.C {
		if err := c.ping(conn); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) ping(conn gim.Conn) error {
	c.Lock()
	defer c.Unlock()
	err := conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("%s send ping to server", c.id))
	return wsutil.WriteClientMessage(conn, ws.OpPing, nil)
}

func NewClient(id string, name string, options ...gim.ClientOptionFunc) *client {
	clientOpts := gim.NewClientOptions()
	for _, opt := range options {
		opt(clientOpts)
	}
	return &client{id: id, name: name, options: clientOpts}
}
