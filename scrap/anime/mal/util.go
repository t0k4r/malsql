package mal

import (
	"strings"
	"time"

	"github.com/go-rod/rod"
)

func FixBlock() {
	page := rod.New().MustConnect().MustPage("https://myanimelist.net/anime/32867/Bungou_Stray_Dogs_2nd_Season")
	for strings.Trim(page.MustElement("title").MustText(), " \n") != "Bungou Stray Dogs 2nd Season (Bungo Stray Dogs 2) - MyAnimeList.net" {
		btn, err := page.Element("button")
		if err != nil {
			break
		}
		btn.MustClick()
		time.Sleep(time.Second * 15)
	}
}
