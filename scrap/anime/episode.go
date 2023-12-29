package anime

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

type Episode struct {
	Index     int
	Title     string
	AltTitle  string
	StreamSrc string
	StreamUrl string
}

func LoadEpisodeStreams(title string) (string, []string, error) {
	var urls []string
	src := "www3.gogoanimes.fi"

	title = strings.ReplaceAll(title, " ", "_")
	url := fmt.Sprintf("https://www3.gogoanimes.fi/search.html?keyword=%v", title)
	res, err := http.Get(url)
	if err != nil {
		return src, urls, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return src, urls, err
	}
	nextUrl, ok := doc.Find("ul.items>li>div.img>a:first-of-type").Attr("href")
	if !ok {
		return src, urls, nil
	}
	infoUrl, title := fmt.Sprintf("https://www3.gogoanimes.fi%v", nextUrl), strings.Split(nextUrl, "/")[2]
	res, err = http.Get(infoUrl)
	if err != nil {
		return src, urls, err
	}
	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return src, urls, err
	}
	id, ok := doc.Find(".movie_id").Attr("value")
	if !ok {
		return src, urls, nil
	}
	var pages []string
	res, err = http.Get(fmt.Sprintf("https://ajax.gogo-load.com/ajax/load-list-episode?ep_start=0&ep_end=228922&id=%v&default_ep=0&alias=%v", id, title))
	if err != nil {
		return src, urls, err
	}
	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return src, urls, err

	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		page, ok := s.Attr("href")
		if ok {
			page = strings.Trim(page, " ")
			page = fmt.Sprintf("https://www3.gogoanimes.fi/%v", page)
			pages = append([]string{page}, pages...)
		}
	})

	wg := new(errgroup.Group)
	urls = make([]string, len(pages))
	for i, page := range pages {
		wg.Go(func(i int, page string) func() error {
			return func() error {
				res, err := http.Get(url)
				if err != nil {
					return err
				}
				doc, err := goquery.NewDocumentFromReader(res.Body)
				if err != nil {
					return err
				}
				vid, ok := doc.Find("li.anime>a").Attr("data-video")
				if ok {
					urls[i] = fmt.Sprintf("%v", vid)
				}
				return nil
			}
		}(i, page))
	}
	return src, urls, wg.Wait()
}

func LoadEpisodes(url string) ([]Episode, error) {
	var episodes []Episode
	offset := 0
	for {
		var newEpisodes []Episode
		res, err := http.Get(fmt.Sprintf("%v/episode?offset=%v", url, offset))
		if err != nil {
			return nil, err
		}
		if res.StatusCode == 429 {
			FixBlock()
			return LoadEpisodes(url)
		}
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return episodes, err
		}
		doc.Find(`td.episode-title .fl-l`).Each(func(i int, s *goquery.Selection) {
			newEpisodes = append(newEpisodes, Episode{Title: s.Text()})
		})
		doc.Find(`td.episode-title .di-ib`).Each(func(i int, s *goquery.Selection) {
			if len(newEpisodes) > i {
				newEpisodes[i].AltTitle = s.Text()
			} else {
				newEpisodes = append(newEpisodes, Episode{AltTitle: s.Text()})
			}
		})
		if len(newEpisodes) == 0 {
			break
		}
		episodes = append(episodes, newEpisodes...)
		offset += 100
		if len(newEpisodes) < 100 {
			break
		}
	}
	return episodes, nil
}
