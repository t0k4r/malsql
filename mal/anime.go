package mal

import (
	"log"
	"net/http"
	"slices"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

type Url string

func (u Url) Id() int {
	return 0
}

type Episode struct {
	Url      string
	Title    string
	TitleAlt string
	Descr    string
}

type Anime struct {
	doc      *goquery.Document
	Url      Url
	Title    string
	Descr    string
	Titles   map[string]string
	Infos    map[string][]string
	Related  map[string][]Url
	Episodes []Episode
}

func (a *Anime) readUrl()     {}
func (a *Anime) readTitles()  {}
func (a *Anime) readDescr()   {}
func (a *Anime) readInfos()   {}
func (a *Anime) readRelated() {}

func (a *Anime) fetchEpisodes(url string) func() error {
	return func() error {
		_ = url
		return nil
	}
}

func (a *Anime) FetchRelated(done *[]int) ([]*Anime, error) {
	var animes []*Anime
	for _, urls := range a.Related {
		for _, url := range urls {
			if slices.Contains(*done, url.Id()) {
				continue
			}
			anime, err := FetchAnime(string(url))
			if err != nil {
				return animes, err
			}
			animes = append(animes, anime)
			*done = append(*done, url.Id())
		}
	}
	for _, anime := range animes {
		related, err := anime.FetchRelated(done)
		if err != nil {
			return animes, err
		}
		animes = append(animes, related...)
	}
	return animes, nil
}

func FetchAnime(url string) (*Anime, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	var anime Anime
	var eg errgroup.Group
	eg.Go(anime.fetchEpisodes(url))
	switch resp.StatusCode {
	case 404:
		return nil, nil
	case 429:
		fixLock()
		return FetchAnime(url)
	default:
		if resp.StatusCode != 200 {
			log.Panicf("unknown statuscode on anime fetch %v", resp.StatusCode)
		}
	}
	anime.doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	anime.readUrl()
	anime.readTitles()
	anime.readDescr()
	anime.readInfos()
	anime.readRelated()
	return &anime, eg.Wait()
}
