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
	offset := 0
	for {
		var newEpisodes []Episode

		res, err := http.Get(fmt.Sprintf("%v/episode?offset=%v", url, offset))
		if err != nil {
			return nil, err
		}
		if res.StatusCode == 429 {
			return episodes, ErrMal429
		}
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return episodes, err
		}
		doc.Find(`td.episode-title .fl-l`).Each(func(i int, s *goquery.Selection) {
			newEpisodes = append(newEpisodes, Episode{Title: s.Text()})
		})
		doc.Find(`td.episode-title di-ib`).Each(func(i int, s *goquery.Selection) {
			newEpisodes[i].AltTitle = s.Text()
		})
		if len(newEpisodes) == 0 {
			break
		}
		episodes = append(episodes, newEpisodes...)
		offset += 100
	}
	return episodes, nil
}
