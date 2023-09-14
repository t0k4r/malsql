package scrap

import (
	"MalSql/scrap/anime"
	"MalSql/scrap/anime/mal"
	"context"
	"database/sql"
	_ "embed"
	"log"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

//go:embed schema.sql
var schema string

var DB *sql.DB

var File *os.File

type Options struct {
	Start int
	End   int
	Skip  bool
	File  bool
	Quick bool
}

type Scraper interface {
	Run()
	inserter()
	dumper()
}

func New(opt Options) Scraper {
	var done []int
	var err error
	if !opt.File {
		DB, err = sql.Open("postgres", os.Getenv("PG_CONN"))
		if err != nil {
			log.Panic(err)
		}
		for _, sql := range strings.Split(schema, ";\n") {
			_, err = DB.Exec(sql)
			if err != nil {
				log.Panic(err)
			}
		}
		if opt.Skip {
			rows, err := DB.Query("select mal_url from animes")
			if err != nil {
				slog.Error(err.Error())
			}
			for rows.Next() {
				var url string
				err := rows.Scan(&url)
				if err != nil {
					slog.Error(err.Error())
				}
				done = append(done, mal.MagicNumber(url))
			}
		}
	} else {
		File, err = os.Create("MalSql_dump.sql")
		if err != nil {
			log.Panic(err)
		}
		_, err = File.WriteString(schema)
		if err != nil {
			log.Panic(err)
		}
	}

	if opt.Quick {
		return &fastScrap{
			Options:  opt,
			wg:       sync.WaitGroup{},
			animes:   make(chan []*anime.Anime),
			scrapped: []int{},
			inserted: []int{},
		}
	}
	return &goodScrap{
		Options: opt,
		wg:      sync.WaitGroup{},
		animes:  make(chan []*anime.Anime),
		done:    done,
	}
}

type goodScrap struct {
	Options
	wg     sync.WaitGroup
	animes chan []*anime.Anime
	done   []int
}

func (s *goodScrap) dumper() {
	for animes := range s.animes {
		var relations []string
		for _, anime := range animes {
			asql, rsql := anime.Sql()
			relations = append(relations, rsql...)
			_, err := File.WriteString(strings.Join(asql, "\n"))
			if err != nil {
				slog.Error(err.Error())
			}
		}
		for _, rsql := range relations {
			_, err := File.WriteString(rsql)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	}
	s.wg.Done()
}

func (s *goodScrap) inserter() {
	for animes := range s.animes {
		var relationsSql []string
		var animesSql []string
		for _, anime := range animes {
			asql, rsql := anime.Sql()
			relationsSql = append(relationsSql, rsql...)
			animesSql = append(animesSql, asql...)
		}
		for _, asql := range animesSql {
			_, err := DB.Exec(asql)
			if err != nil {
				slog.Error(err.Error())
			}
		}
		for _, rsql := range relationsSql {
			_, err := DB.Exec(rsql)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	}
	s.wg.Done()
}

func (s *goodScrap) Run() {
	if s.File {
		s.wg.Add(1)
		go s.dumper()
	} else {
		s.wg.Add(1)
		go s.inserter()
	}
	n := time.Now()
	for i := s.Start; i < s.End; i++ {
		if slices.Contains(s.done, i) {
			continue
		}
		anime := loadAnime(i)
		if anime == nil {
			continue
		}
		var series series
		series.load(anime)
		s.done = append(s.done, series.done...)
		s.animes <- series.animes
	}
	close(s.animes)
	s.wg.Wait()
	slog.Info("Done", "animes", len(s.done), "took", time.Since(n))
}

type fastScrap struct {
	Options
	wg       sync.WaitGroup
	animes   chan []*anime.Anime
	scrapped []int
	inserted []int
}

func (s *fastScrap) dumper() {
	for animes := range s.animes {
		var relations []string
		for _, anime := range animes {
			s.inserted = append(s.inserted, anime.MagicNumber())
			asql, rsql := anime.Sql()
			relations = append(relations, rsql...)
			_, err := File.WriteString(strings.Join(asql, "\n"))
			if err != nil {
				slog.Error(err.Error())
			}
		}
		for _, rsql := range relations {
			_, err := File.WriteString(rsql)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	}
	s.wg.Done()
}

func (s *fastScrap) inserter() {
	for animes := range s.animes {
		var relationsSql []string
		var animesSql []string
		for _, anime := range animes {
			s.inserted = append(s.inserted, anime.MagicNumber())
			asql, rsql := anime.Sql()
			relationsSql = append(relationsSql, rsql...)
			animesSql = append(animesSql, asql...)
		}
		for _, asql := range animesSql {
			_, err := DB.Exec(asql)
			if err != nil {
				slog.Error(err.Error())
			}
		}
		for _, rsql := range relationsSql {
			_, err := DB.Exec(rsql)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	}
	s.wg.Done()
}

type ctxkey string

func (s *fastScrap) Run() {

	//2023/09/14 01:03:42 INFO Done animes=47 unique=45 took=13.585431001s
	var localwg sync.WaitGroup
	urls := make(chan int)
	if s.File {
		s.wg.Add(1)
		go s.dumper()
	} else {
		s.wg.Add(1)
		go s.inserter()
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		localwg.Add(1)
		go func() {
			for url := range urls {
				anime := loadAnime(url)
				if anime != nil {
					ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ctxkey("fsc"), s))
					var series series
					series.loadCtx(anime, ctx, cancel)
					select {
					case <-ctx.Done():
						cancel()
					default:
						s.animes <- series.animes
					}

				}
			}
			localwg.Done()
		}()
	}
	n := time.Now()
	for i := s.Start; i < s.End; i++ {
		if slices.Contains(s.scrapped, i) {
			continue
		}
		urls <- i
	}
	close(urls)
	localwg.Wait()
	close(s.animes)
	s.wg.Wait()
	slog.Info("Done", "animes", len(s.scrapped), "unique", len(s.inserted), "took", time.Since(n))
}

// func (s *fastScrap) loadRelatedCtx(root *anime.Anime, ctx context.Context, cancel context.CancelFunc) []*anime.Anime {
// 	var wg sync.WaitGroup
// 	var animes []*anime.Anime
// 	select {
// 	case <-ctx.Done():
// 	default:
// 		animes = append(animes, root)
// 		s.scrapped = append(s.scrapped, root.MagicNumber())
// 		for _, r := range root.Related {
// 			_ = r
// 		}
// 	}
// 	wg.Wait()
// 	return animes
// }

// type Scraper struct {
// 	wg     sync.WaitGroup
// 	done   []int
// 	animes chan []*anime.Anime
// 	db     *sql.DB
// }

// func Newo() Scraper {
// 	return Scraper{
// 		wg:     sync.WaitGroup{},
// 		done:   []int{},
// 		animes: make(chan []*anime.Anime),
// 		db:     nil,
// 	}
// }

// func (s *Scraper) Run(opts Options) {
// 	n := time.Now()
// 	if opts.Skip || !opts.File {
// 		var err error
// 		s.db, err = sql.Open("postgres", os.Getenv("PG_CONN"))
// 		if err != nil {
// 			log.Panic(err)
// 		}
// 		for i, table := range strings.Split(schema, ";\n") {
// 			_, err := s.db.Exec(table)
// 			if err != nil {
// 				log.Println(i)
// 				log.Fatal(err)
// 			}
// 		}
// 	}
// 	if opts.Skip {
// 		rows, err := s.db.Query("select mal_url from animes")
// 		if err != nil {
// 			fmt.Printf("\x1b[0;31mError:\x1b[0m %v\n", err)
// 		}
// 		for rows.Next() {
// 			var url string
// 			err := rows.Scan(&url)
// 			if err != nil {
// 				fmt.Printf("\x1b[0;31mError:\x1b[0m %v\n", err)
// 			}
// 			s.done = append(s.done, mal.MagicNumber(url))
// 		}
// 	}
// 	s.wg.Add(1)

// 	if opts.File {
// 		go s.dumper(opts.Quick)
// 	} else {
// 		go s.inserter(opts.Quick)
// 	}
// 	iChan := make(chan int)
// 	var qwg sync.WaitGroup
// 	if opts.Quick {
// 		for i := 0; i < runtime.NumCPU(); i++ {
// 			qwg.Add(1)
// 			go func() {
// 				for i := range iChan {
// 					anime := loadAnime(i)
// 					if anime == nil {
// 						continue
// 					}
// 					s.animes <- s.loadRelated(anime)
// 				}
// 				qwg.Done()
// 			}()
// 		}
// 	}
// 	for i := opts.Start; i < opts.End; i++ {
// 		if slices.Contains(s.done, i) {
// 			continue
// 		}
// 		if opts.Quick {
// 			iChan <- i
// 			continue
// 		}
// 		anime := loadAnime(i)
// 		if anime == nil {
// 			continue
// 		}
// 		s.animes <- s.loadRelated(anime)
// 	}
// 	close(iChan)

// 	qwg.Wait()
// 	close(s.animes)
// 	s.wg.Wait()
// 	fmt.Printf("\x1b[0;34mStats:\x1b[0m %v animes(%v unique) \n\ttook %v\n", len(s.done), uniqueLen(s.done), time.Since(n))
// }
// func uniqueLen(s []int) int {
// 	var u []int
// 	for _, i := range s {
// 		if !slices.Contains(u, i) {
// 			u = append(u, i)
// 		}
// 	}
// 	return len(u)
// }

// func (s *Scraper) inserter(fast bool) {
// 	for animes := range s.animes {
// 		if fast {
// 			for _, anime := range animes {
// 				s.done = append(s.done, anime.MagicNumber())
// 			}
// 		}
// 		var animesSql []string
// 		var relationsSql []string
// 		for _, anime := range animes {
// 			aSql, rSql := anime.Sql()
// 			animesSql = append(animesSql, aSql...)
// 			relationsSql = append(relationsSql, rSql...)
// 		}
// 		for _, sql := range append(animesSql, relationsSql...) {
// 			_, err := s.db.Exec(sql)
// 			if err != nil {
// 				fmt.Printf("\x1b[0;31mError:\x1b[0m %v\n", err)
// 			}
// 		}
// 	}
// 	s.wg.Done()
// }
// func (s *Scraper) dumper(fast bool) {
// 	file, err := os.Create("MalSql_dump.sql")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	_, err = file.WriteString(schema)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	for animes := range s.animes {
// 		if fast {
// 			for _, anime := range animes {
// 				s.done = append(s.done, anime.MagicNumber())
// 			}
// 		}
// 		var animesSql []string
// 		var relationsSql []string
// 		for _, anime := range animes {
// 			aSql, rSql := anime.Sql()
// 			animesSql = append(animesSql, aSql...)
// 			relationsSql = append(relationsSql, rSql...)
// 		}
// 		for _, sql := range append(animesSql, relationsSql...) {
// 			_, err := file.WriteString(fmt.Sprintf("%v;\n", sql))
// 			if err != nil {
// 				log.Panic(err)
// 			}
// 		}
// 	}
// 	s.wg.Done()
// }

type series struct {
	done   []int
	animes []*anime.Anime
}

func (series *series) loadCtx(root *anime.Anime, ctx context.Context, cancel context.CancelFunc) {
	select {
	case <-ctx.Done():
		cancel()
	default:
		fscrap := ctx.Value(ctxkey("fsc")).(*fastScrap)
		var wg sync.WaitGroup
		series.animes = append(series.animes, root)
		series.done = append(series.done, root.MagicNumber())
		fscrap.scrapped = append(fscrap.scrapped, root.MagicNumber())
		for _, r := range root.Related {
			if slices.Contains(series.done, mal.MagicNumber(r.Url)) {
				continue
			}
			wg.Add(1)
			go func(url string) {
				anime := loadAnime(url)
				if anime != nil {
					fscrap.scrapped = append(fscrap.scrapped, anime.MagicNumber())
					if slices.Contains(fscrap.inserted, anime.MagicNumber()) {
						cancel()
					} else if !slices.Contains(series.done, anime.MagicNumber()) {
						series.loadCtx(anime, ctx, cancel)
					}
				}
				wg.Done()
			}(r.Url)
		}
		wg.Wait()
	}
}

func (s *series) load(root *anime.Anime) {
	var wg sync.WaitGroup
	s.animes = append(s.animes, root)
	s.done = append(s.done, root.MagicNumber())
	for _, r := range root.Related {
		if slices.Contains(s.done, mal.MagicNumber(r.Url)) {
			continue
		}
		wg.Add(1)
		go func(url string) {
			anime := loadAnime(url)
			if anime != nil && !slices.Contains(s.done, anime.MagicNumber()) {
				s.load(anime)
			}
			wg.Done()
		}(r.Url)

	}
	wg.Wait()

}

func loadAnime[T int | string](id T) *anime.Anime {
	n := time.Now()
	anime, err := anime.LoadAnime(id)
	switch err {
	case mal.ErrMal404:
		return nil
	case mal.ErrMal429:
		slog.Warn("MalBlocked")
		mal.FixBlock()
		return loadAnime(id)
	case nil:
		slog.Info("Scrapped", "anime", anime.Title, "episodes", len(anime.Episodes), "took", time.Since(n))
		return anime
	default:
		time.Sleep(time.Second * 5)
		slog.Error(err.Error())
		return loadAnime(id)
	}
}
