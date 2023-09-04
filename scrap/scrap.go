package scrap

import (
	"MalSql/scrap/anime"
	"MalSql/scrap/anime/mal"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "embed"

	_ "github.com/lib/pq"
	"golang.org/x/exp/slices"
)

//go:embed schema.sql
var schema string

type Scraper struct {
	wg     sync.WaitGroup
	done   []int
	animes chan []*anime.Anime
	db     *sql.DB
}

func New(db *sql.DB) Scraper {
	for i, table := range strings.Split(schema, ";\n") {
		_, err := db.Exec(table)
		if err != nil {
			log.Println(i)
			log.Fatal(err)
		}
	}
	return Scraper{
		wg:     sync.WaitGroup{},
		done:   []int{},
		animes: make(chan []*anime.Anime),
		db:     db,
	}
}

func (s *Scraper) Run(start, end int) {
	n := time.Now()
	s.wg.Add(1)
	go s.inserter()
	for i := start; i < end; i++ {
		if slices.Contains(s.done, i) {
			continue
		}
		anime := loadAnime(i)
		if anime == nil {
			continue
		}
		s.animes <- s.loadRelated(anime)
	}
	close(s.animes)
	s.wg.Wait()
	fmt.Printf("\x1b[0;34mStats:\x1b[0m %v animes \n\ttook %v\n", len(s.done), time.Since(n))
}

func (s *Scraper) RunSkipDone(start, end int) {
	rows, err := s.db.Query("select mal_url from animes")
	if err != nil {
		fmt.Printf("\x1b[0;31mError:\x1b[0m %v\n", err)
	}
	for rows.Next() {
		var url string
		err := rows.Scan(&url)
		if err != nil {
			fmt.Printf("\x1b[0;31mError:\x1b[0m %v\n", err)
		}
		s.done = append(s.done, mal.MagicNumber(url))
	}
	s.Run(start, end)
}

func (s *Scraper) inserter() {
	for animes := range s.animes {
		var animesSql []string
		var relationsSql []string
		for _, anime := range animes {
			aSql, rSql := anime.Sql()
			animesSql = append(animesSql, aSql...)
			relationsSql = append(relationsSql, rSql...)
		}
		for _, sql := range append(animesSql, relationsSql...) {
			_, err := s.db.Exec(sql)
			if err != nil {
				fmt.Printf("\x1b[0;31mError:\x1b[0m %v\n", err)
			}
		}
	}
	s.wg.Done()
}

func (s *Scraper) loadRelated(root *anime.Anime) []*anime.Anime {
	var wg sync.WaitGroup
	var animes []*anime.Anime
	animes = append(animes, root)
	s.done = append(s.done, root.MagicNumber())
	for _, related := range root.Related {
		if slices.Contains(s.done, mal.MagicNumber(related.Url)) {
			continue
		}
		wg.Add(1)
		go func(related mal.Related) {
			if !slices.Contains(s.done, mal.MagicNumber(related.Url)) {
				anime := loadAnime(related.Url)
				if anime != nil && !slices.Contains(s.done, anime.MagicNumber()) {
					animes = append(animes, s.loadRelated(anime)...)
				}
			}
			wg.Done()
		}(related)
	}
	wg.Wait()
	return animes
}

func loadAnime[T int | string](id T) *anime.Anime {
	n := time.Now()
	anime, err := anime.LoadAnime(id)
	switch err {
	case mal.ErrMal404:
		return nil
	case mal.ErrMal429:
		fmt.Println("\x1b[0;33mMalBlocked!!!\x1b[0m")
		mal.FixBlock()
		return loadAnime(id)
	case nil:
		fmt.Printf("\x1b[0;32mScrapped:\x1b[0m %v\n\t%v episodes\n\ttook %v\n", anime.Title, len(anime.Episodes), time.Since(n))
		return anime
	default:
		time.Sleep(time.Second * 5)
		log.Printf("\x1b[0;31mError:\x1b[0m %v\n", err)
		return loadAnime(id)
	}
}
