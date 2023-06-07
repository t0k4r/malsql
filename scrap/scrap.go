package scrap

import (
	"MalSql/scrap/anime"
	"MalSql/scrap/anime/mal"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/exp/slices"
)

type Scrapper struct {
	wg          *sync.WaitGroup
	done        []int
	urls        chan int
	series      chan *anime.Series
	workersDone int
	maxWorkers  int
	db          *sql.DB
}

func New() Scrapper {
	db, err := sql.Open("postgres", os.Getenv("PG_CONN"))
	if err != nil {
		log.Panic(err)
	}
	var done []int
	for _, sql := range strings.Split(anime.Schema, ";\n") {
		_, err := db.Exec(sql)
		if err != nil {
			log.Panic(err)
		}
	}
	rows, err := db.Query("SELECT anime_mal_URL FROM animes")
	if err != nil {
		log.Panic(err)
	}
	for rows.Next() {
		var url string
		rows.Scan(&url)
		done = append(done, mal.MagicNumber(url))
	}
	return Scrapper{
		wg:          &sync.WaitGroup{},
		urls:        make(chan int),
		series:      make(chan *anime.Series),
		workersDone: 0,
		maxWorkers:  16,
		done:        done,
		db:          db,
	}
}
func (s *Scrapper) Run(start, end int) {
	s.wg.Add(1)
	go s.onScrap()
	for i := 0; i < s.maxWorkers; i++ {
		s.wg.Add(1)
		go s.scrap()
	}
	strt := time.Now()
	for i := start; i < end; i++ {
		if !slices.Contains(s.done, i) {
			s.urls <- i
		}
	}
	close(s.urls)
	s.wg.Wait()
	fmt.Printf("Scrapped %v took %v\n", len(s.done), time.Since(strt))
}

func (s *Scrapper) scrap() {
	for url := range s.urls {
		if slices.Contains(s.done, url) {
			continue
		}
		st := time.Now()
		a, err := anime.LoadAnime(url)
		if err == mal.ErrMal404 {
			continue
		} else if err != nil {
			log.Panic(err)
		}
		ser := anime.NewSeries(a)
		s.done = append(s.done, ser.Done...)
		fmt.Printf("Scrapped %v +%v others took %v\n", a.Title, len(ser.Animes)-1, time.Since(st))
		s.series <- ser
	}
	s.workersDone++
	if s.workersDone == s.maxWorkers {
		close(s.series)
	}
	s.wg.Done()
}

func (s *Scrapper) onScrap() {
	for series := range s.series {
		for _, q := range series.Sql() {
			_, err := s.db.Exec(q)
			if err != nil {
				log.Println(err)
			}
		}
	}
	s.wg.Done()
}
