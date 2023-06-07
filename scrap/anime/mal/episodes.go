package mal

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func GetEpisodes(url string) ([]string, error) {
	var episodes []string
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
		episodes = append(episodes, s.Text())
	})
	return episodes, nil
}
