package main

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

type Anime struct {
	doc *goquery.Document
}

func (a *Anime) fetchEpisodes() func() error {
	return func() error { return nil }
}

func FetchAnime(url string) (*Anime, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 404:
		return nil, nil
	case 429:
		log.Panic("mal locked")
	default:
		if resp.StatusCode != 200 {
			log.Panicf("unknown statuscode on anime fetch %v", resp.StatusCode)
		}
	}
	var anime Anime
	anime.doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	var eg errgroup.Group
	eg.Go(anime.fetchEpisodes())
	return &anime, eg.Wait()
}
