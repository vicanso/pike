package middleware

import (
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/vicanso/pike/pike"
)

type (
	// RecoverConfig recover config
	RecoverConfig struct {
		// Size of the stack to be printed.
		// Optional. Default value 4KB.
		StackSize int `yaml:"stack_size"`

		// DisableStackAll disables formatting stack traces of all other goroutines
		// into buffer after the trace for the current goroutine.
		// Optional. Default value false.
		DisableStackAll bool `yaml:"disable_stack_all"`

		// DisablePrintStack disables printing stack trace.
		// Optional. Default value as false.
		DisablePrintStack bool `yaml:"disable_print_stack"`
	}
)

var (
	// DefaultRecoverConfig is the default Recover middleware config.
	DefaultRecoverConfig = RecoverConfig{
		StackSize:         4 << 10, // 4 KB
		DisableStackAll:   false,
		DisablePrintStack: false,
	}
)

// Recover 异常捕获，异常程序shutdown
func Recover(config RecoverConfig) pike.Middleware {
	return func(c *pike.Context, next pike.Next) error {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				stack := make([]byte, config.StackSize)
				length := runtime.Stack(stack, !config.DisableStackAll)
				if !config.DisablePrintStack {
					log.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
				}
				c.Error(err)
			}
		}()
		return next()
	}
}
