package scrap

import (
	"MalSql/scrap/anime"
	"MalSql/scrap/anime/mal"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/exp/slices"
)

type Scraper struct {
	done []int
}

func New() Scraper {
	return Scraper{}
}

func (s *Scraper) Run(start, end int) {
	for i := start; i < end; i++ {
		anime := loadAnime(i)
		s.done = append(s.done, i)
		if anime == nil {
			continue
		}
		animes := s.loadRelated(anime)
		fmt.Println(len(animes))
	}
}

func loadAnime[T int | string](id T) *anime.Anime {
	n := time.Now()
	anime, err := anime.LoadAnime(id)
	switch err {
	case mal.ErrMal404:
		return nil
	case mal.ErrMal429:
		fmt.Println("\x1b[0;33mMalBlocked\x1b[0m")
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
