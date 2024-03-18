package mal

import (
	"fmt"
	"log/slog"
	"sync"
)

var lock sync.Mutex

func fixLock() {
	switch lock.TryLock() {
	case true:
		slog.Error("MAL is locked please visit url to unlockit and press enter", "url", "")
		fmt.Scanln()
	default:
		lock.Lock()
	}
	lock.Unlock()
}
