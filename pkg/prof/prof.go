package prof

import (
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

var count int32 = 0

// Add .
func Add(n int32) {
	atomic.AddInt32(&count, n)
}

// Run .
func Run() {
	for {
		log.Debug().Int32("count", atomic.LoadInt32(&count)).Msg("prof")
		time.Sleep(time.Second)
	}
}
