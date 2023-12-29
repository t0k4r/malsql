package anime

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
)

var Fixlock sync.Mutex = sync.Mutex{}
var browser *rod.Browser = rod.New().MustConnect()

func FixBlock() {
	if Fixlock.TryLock() {
		slog.Warn("Mal Blocked")
		page := browser.MustPage("https://myanimelist.net/anime/32867/Bungou_Stray_Dogs_2nd_Season")
		for strings.Trim(page.MustElement("title").MustText(), " \n") != "Bungou Stray Dogs 2nd Season (Bungo Stray Dogs 2) - MyAnimeList.net" {
			btn, err := page.Element("button")
			if err != nil {
				break
			}
			btn.MustClick()
			time.Sleep(time.Second * 15)
		}
		page.Close()
	} else {
		Fixlock.Lock()
	}
	Fixlock.Unlock()
}
