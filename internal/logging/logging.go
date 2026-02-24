package logging

import "fmt"

var verbose bool

func SetVerbose(enabled bool) {
	verbose = enabled
}

func Verbosef(format string, args ...any) {
	if !verbose {
		return
	}
	fmt.Printf(format+"\n", args...)
}

func Infof(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}
