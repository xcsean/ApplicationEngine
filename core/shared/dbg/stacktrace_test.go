package dbg

import (
	"testing"

	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

func test() bool {
	defer Stacktrace()
	var a map[int]int
	a[0] = 1
	return true
}
func TestStacktrace(t *testing.T) {
	log.SetupMainLogger("./", "trace.log", "debug")
	test()
}
