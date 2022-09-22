package gox

import (
	"fmt"
	"github.com/kkakoz/gim/pkg/logger"
	"runtime"
)

type GoOption struct {
	panicBack func()
}

type GoOptionFunc func(option *GoOption)

func Go(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				err = fmt.Errorf("goroutine: panic recovered: %s\n%s", err, buf)
				logger.Error(fmt.Sprintf("goroutine: panic recovered: %s\n%s", err, buf))
			}
		}()
		f()
	}()
}
