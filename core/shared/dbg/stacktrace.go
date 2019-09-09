package dbg

import (
	"runtime/debug"
	"strings"

	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

// Stacktrace trace the call stack
func Stacktrace() {
	if err := recover(); err != nil {
		log.Error("%v", err)
		stack := string(debug.Stack())
		ss := strings.Split(stack, "\n")
		for i := 0; i < len(ss); i++ {
			str := strings.Replace(ss[i], "\t", "    ", -1)
			log.Error("%s", str)
		}
	}
}
