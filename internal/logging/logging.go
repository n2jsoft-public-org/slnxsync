package logging

import (
	"fmt"
	"sync/atomic"
)

var verbose atomic.Bool

func SetVerbose(enabled bool) {
	verbose.Store(enabled)
}

func Verbosef(format string, args ...any) {
	if !verbose.Load() {
		return
	}
	fmt.Printf(format+"\n", args...)
}

func Infof(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}
