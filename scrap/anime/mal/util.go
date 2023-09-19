package mal

import (
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
)

var fixlock sync.Mutex = sync.Mutex{}
var browser *rod.Browser

func FixBlock() {
	fixlock.Lock()
	if browser == nil {
		browser = rod.New().MustConnect()
	}
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
	fixlock.Unlock()
}
