package scrap

import (
	"MalSql/scrap/anime"
	"MalSql/scrap/anime/mal"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
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

type Options struct {
	Start int
	End   int
	Skip  bool
	File  bool
	Quick bool
}

func New() Scraper {
	return Scraper{
		wg:     sync.WaitGroup{},
		done:   []int{},
		animes: make(chan []*anime.Anime),
		db:     nil,
	}
}

func (s *Scraper) Run(opts Options) {
	n := time.Now()
	if opts.Skip || !opts.File {
		var err error
		s.db, err = sql.Open("postgres", os.Getenv("PG_CONN"))
		if err != nil {
			log.Panic(err)
		}
		for i, table := range strings.Split(schema, ";\n") {
			_, err := s.db.Exec(table)
			if err != nil {
				log.Println(i)
				log.Fatal(err)
			}
		}
	}
	if opts.Skip {
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
	}
	s.wg.Add(1)

	if opts.File {
		go s.dumper(opts.Quick)
	} else {
		go s.inserter(opts.Quick)
	}
	iChan := make(chan int)
	var qwg sync.WaitGroup
	if opts.Quick {
		for i := 0; i < runtime.NumCPU(); i++ {
			qwg.Add(1)
			go func() {
				for i := range iChan {
					anime := loadAnime(i)
					if anime == nil {
						continue
					}
					s.animes <- s.loadRelated(anime)
				}
				qwg.Done()
			}()
		}
	}
	for i := opts.Start; i < opts.End; i++ {
		if slices.Contains(s.done, i) {
			continue
		}
		if opts.Quick {
			iChan <- i
			continue
		}
		anime := loadAnime(i)
		if anime == nil {
			continue
		}
		s.animes <- s.loadRelated(anime)
	}
	close(iChan)

	qwg.Wait()
	close(s.animes)
	s.wg.Wait()
	fmt.Printf("\x1b[0;34mStats:\x1b[0m %v animes(%v unique) \n\ttook %v\n", len(s.done), uniqueLen(s.done), time.Since(n))
}
func uniqueLen(s []int) int {
	var u []int
	for _, i := range s {
		if !slices.Contains(u, i) {
			u = append(u, i)
		}
	}
	return len(u)
}

func (s *Scraper) inserter(fast bool) {
	for animes := range s.animes {
		if fast {
			for _, anime := range animes {
				s.done = append(s.done, anime.MagicNumber())
			}
		}
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
func (s *Scraper) dumper(fast bool) {
	file, err := os.Create("MalSql_dump.sql")
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.WriteString(schema)
	if err != nil {
		log.Panic(err)
	}
	for animes := range s.animes {
		if fast {
			for _, anime := range animes {
				s.done = append(s.done, anime.MagicNumber())
			}
		}
		var animesSql []string
		var relationsSql []string
		for _, anime := range animes {
			aSql, rSql := anime.Sql()
			animesSql = append(animesSql, aSql...)
			relationsSql = append(relationsSql, rSql...)
		}
		for _, sql := range append(animesSql, relationsSql...) {
			_, err := file.WriteString(fmt.Sprintf("%v;\n", sql))
			if err != nil {
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
