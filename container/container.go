package container

import (
	"fmt"
	"github.com/kkakoz/gim"
	"github.com/kkakoz/gim/naming"
	"github.com/kkakoz/gim/pkg/gox"
	"github.com/kkakoz/gim/pkg/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
)

const (
	stateUninitialized = iota
	stateInitialized
	stateStarted
	stateClosed
)

type Container struct {
	sync.RWMutex
	Naming     naming.Naming
	Srv        gim.Server
	state      uint32
	srvclients map[string]IClientMap
	selector   Selector
	dialer     gim.Dialer
	deps       map[string]struct{}
}

var log = logger.WithFields(zap.String("module", "container"))
var c = &Container{
	state:      stateUninitialized,
	srvclients: make(map[string]IClientMap),
	selector:   &HashSelector{},
	deps:       make(map[string]struct{}, 0),
}

func Default() *Container {
	return c
}

func Init(srv gim.Server, deps ...string) error {
	if !atomic.CompareAndSwapUint32(&c.state, stateUninitialized, stateInitialized) {
		return errors.New("has Initialized")
	}
	c.Srv = srv
	for _, dep := range deps {
		if _, ok := c.deps[dep]; ok {
			continue
		}
		c.deps[dep] = struct{}{}
	}
	logger.WithFields(zap.String("func", "Init")).Info(fmt.Sprintf("srv %s:%s - deps %v", srv.ServiceID(), srv.ServiceName(), c.deps))
	c.srvclients = make(map[string]IClientMap, len(deps))
	return nil
}

// SetDialer set tcp dialer
func SetDialer(dialer gim.Dialer) {
	c.dialer = dialer
}

// SetSelector set a default selector
func SetSelector(selector Selector) {
	c.selector = selector
}

func Start() error {
	if c.Naming == nil {
		return fmt.Errorf("no naming")
	}
	if !atomic.CompareAndSwapUint32(&c.state, stateInitialized, stateStarted) {
		return fmt.Errorf("")
	}

	gox.Go(func() {
		func(srv gim.Server) {
			err := srv.Start()
			if err != nil {
				logger.Error("start container err:" + err.Error())
			}
		}(c.Srv)
	})

	for _, service := range c.deps {
		gox.Go(func() {
			connectToService(service)
		})
	}
	return nil
}

func connectToService(service struct{}) {

}
