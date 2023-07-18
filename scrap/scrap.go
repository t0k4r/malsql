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
	log.Printf("\x1b[0;34mStats:\x1b[0m %v animes \n\ttook %v\n", len(s.done), time.Since(n))
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
				log.Println(sql)
				log.Panic(err)
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

// type Scrapper struct {
// 	wg          *sync.WaitGroup
// 	done        []int
// 	urls        chan int
// 	series      chan *anime.Series
// 	workersDone int
// 	maxWorkers  int
// 	db          *sql.DB
// }

// func New() Scrapper {
// 	db, err := sql.Open("postgres", os.Getenv("PG_CONN"))
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	var done []int
// 	for _, sql := range strings.Split(anime.Schema, ";\n") {
// 		_, err := db.Exec(sql)
// 		if err != nil {
// 			log.Panic(err)
// 		}
// 	}
// 	rows, err := db.Query("SELECT mal_url FROM animes")
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	for rows.Next() {
// 		var url string
// 		rows.Scan(&url)
// 		done = append(done, mal.MagicNumber(url))
// 	}
// 	return Scrapper{
// 		wg:          &sync.WaitGroup{},
// 		urls:        make(chan int),
// 		series:      make(chan *anime.Series),
// 		workersDone: 0,
// 		maxWorkers:  8,
// 		done:        done,
// 		db:          db,
// 	}
// }
// func (s *Scrapper) Run(start, end int) {
// 	s.wg.Add(1)
// 	go s.onScrap()
// 	for i := 0; i < s.maxWorkers; i++ {
// 		s.wg.Add(1)
// 		go s.scrap()
// 	}
// 	strt := time.Now()
// 	for i := start; i < end; i++ {
// 		if !slices.Contains(s.done, i) {
// 			s.urls <- i
// 		}
// 	}
// 	close(s.urls)
// 	s.wg.Wait()
// 	fmt.Printf("Scrapped %v took %v\n", len(s.done), time.Since(strt))
// }

// func (s *Scrapper) scrap() {
// 	for url := range s.urls {
// 		if slices.Contains(s.done, url) {
// 			continue
// 		}
// 		st := time.Now()
// 		a, err := anime.LoadAnime(url)
// 		if err == mal.ErrMal404 {
// 			continue
// 		} else if err != nil {
// 			log.Panic(err)
// 		}
// 		ser := anime.NewSeries(a)
// 		s.done = append(s.done, ser.Done...)
// 		fmt.Printf("Scrapped %v +%v others took %v\n", a.Title, len(ser.Animes)-1, time.Since(st))
// 		s.series <- ser
// 	}
// 	s.workersDone++
// 	if s.workersDone == s.maxWorkers {
// 		close(s.series)
// 	}
// 	s.wg.Done()
// }

// func (s *Scrapper) onScrap() {
// 	for series := range s.series {
// 		for _, q := range series.Sql() {
// 			_, err := s.db.Exec(q)
// 			if err != nil {
// 				log.Println(err)
// 			}
// 		}
// 	}
// 	s.wg.Done()
// }
