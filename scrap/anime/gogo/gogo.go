package gogo

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var ErrGoGo404 = errors.New("empty anime search gogo404")

func GetEpisodes(title string) ([]string, error) {
	var urls []string
	infoUrl, title, err := getInfoPage(title)
	if err != nil {
		return urls, err
	}
	id, err := getMovieId(infoUrl)
	if err != nil {
		return urls, err
	}
	pages, err := getVideoPages(id, title)
	if err != nil {
		return urls, err
	}
	var wg sync.WaitGroup
	urls = make([]string, len(pages))
	for i, page := range pages {
		wg.Add(1)
		go func(i int, url string) {
			res, err := http.Get(url)
			if err != nil {
				log.Panic(err)
			}
			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				log.Panic(err)
			}
			vid, ok := doc.Find("li.anime>a").Attr("data-video")
			if ok {
				urls[i] = fmt.Sprintf("%v", vid)
			}
			wg.Done()
		}(i, page)
	}
	wg.Wait()
	return urls, nil

}

func getInfoPage(title string) (string, string, error) {
	title = strings.ReplaceAll(title, " ", "_")
	url := fmt.Sprintf("https://www3.gogoanimes.fi/search.html?keyword=%v", title)
	res, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", err
	}
	nextUrl, ok := doc.Find("ul.items>li>div.img>a:first-of-type").Attr("href")
	if ok {
		return fmt.Sprintf("https://www3.gogoanimes.fi%v", nextUrl), strings.Split(nextUrl, "/")[2], nil
	}
	return "", "", ErrGoGo404
}

func getMovieId(infoUrl string) (string, error) {
	res, err := http.Get(infoUrl)
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	id, ok := doc.Find(".movie_id").Attr("value")
	if ok {
		return id, nil
	}
	return "", ErrGoGo404
}

func getVideoPages(id string, title string) ([]string, error) {
	var pages []string
	res, err := http.Get(fmt.Sprintf("https://ajax.gogo-load.com/ajax/load-list-episode?ep_start=0&ep_end=228922&id=%v&default_ep=0&alias=%v", id, title))
	if err != nil {
		return pages, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return pages, err
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		page, ok := s.Attr("href")
		if ok {
			page = strings.Trim(page, " ")
			page = fmt.Sprintf("https://www3.gogoanimes.fi/%v", page)
			pages = append([]string{page}, pages...)
		}
	})
	return pages, nil
}
