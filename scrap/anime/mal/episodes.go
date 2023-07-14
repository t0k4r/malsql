package mal

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Episode struct {
	Title    string
	AltTitle string
}

func GetEpisodes(url string) ([]Episode, error) {
	var episodes []Episode
	res, err := http.Get(fmt.Sprintf("%v/episode", url))
	if err != nil {
		return episodes, err
	}
	if res.StatusCode == 429 {
		return episodes, ErrMal429
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return episodes, err
	}
	doc.Find(`td.episode-title .fl-l`).Each(func(i int, s *goquery.Selection) {
		episodes = append(episodes, Episode{Title: s.Text()})
	})
	doc.Find(`td.episode-title di-ib`).Each(func(i int, s *goquery.Selection) {
		episodes[i].AltTitle = s.Text()
	})
	return episodes, nil
}
