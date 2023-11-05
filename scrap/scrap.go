package scrap

import (
	"MalSql/scrap/anime"
	"MalSql/scrap/anime/mal"
	"MalSql/scrap/plog"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/t0k4r/qb"
)

//go:embed schema_pg.sql
var pgSchema string

//go:embed schema_sqlite.sql
var sqliteSchema string

// var DB *sql.DB

// var File *os.File

type Options struct {
	Start  int
	End    int
	Skip   bool
	File   bool
	Update bool
	Driver string
	Conn   string
	Env    bool
}

func (o Options) onConflict() qb.Conflict {
	if o.Update {
		return qb.Replace
	} else {
		return qb.Ignore
	}
}

type saver interface {
	listen(chan []*anime.Anime)
	wait()
}

type scraper struct {
	Options
	saver
	done []int
}

func New(opts Options) scraper {
	slog.SetDefault(plog.NewPlog())

	if opts.Env {
		err := godotenv.Load()
		if err != nil {
			slog.Error(err.Error())
		}
		opts.Conn = os.Getenv("MALSQL_DB")
	}
	if opts.Driver != "sqlite3" && opts.Driver != "postgres" {
		log.Fatal("unknown driver ", opts.Driver)
	}

	var done []int
	var saver saver
	var err error
	if opts.File {
		if opts.Conn == "./MalSql.sqlite" {
			opts.Conn = fmt.Sprintf("MalSql.%v.sql", opts.Driver)
		}
		saver, err = newFSaver(opts)
		if err != nil {
			log.Fatal(err)
		}

	} else {
		dsv, err := newDSaver(opts)
		if err != nil {
			log.Fatal(err)
		}
		if opts.Skip {
			done, err = dsv.skip()
			if err != nil {
				log.Fatal(err)
			}
		}
		saver = dsv
	}

	return scraper{Options: opts, saver: saver, done: done}
}

func (s scraper) Run() {
	snd := make(chan []*anime.Anime)
	s.listen(snd)
	t := time.Now()
	for i := s.Start; i < s.End; i++ {
		if slices.Contains(s.done, i) {
			continue
		}
		anime := loadAnime(i)
		if anime == nil {
			continue
		}
		var ser series
		ser.load(anime)
		s.done = append(s.done, ser.done...)
		snd <- ser.animes
	}

	close(snd)
	s.wait()
	slog.Info("Done", "animes", len(s.done), "took", time.Since(t))
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
		return &anime
	default:
		time.Sleep(time.Second * 5)
		slog.Error(err.Error())
		return loadAnime(id)
	}
}

type series struct {
	done   []int
	animes []*anime.Anime
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
