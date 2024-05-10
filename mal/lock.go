package mal

import (
	"log/slog"
	"sync"
	"time"
)

var lock sync.Mutex

func waitLock() {
	switch lock.TryLock() {
	case true:
		t := time.Now().Add(time.Minute * 5)
		slog.Error("MAL is locked waiting for it to unlock next try on", "time", t)
		time.Sleep(time.Minute * 5)
	default:
		lock.Lock()
	}
	lock.Unlock()
}
