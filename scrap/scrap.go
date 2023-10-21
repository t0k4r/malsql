package scrap

import (
	"MalSql/scrap/anime"
	"MalSql/scrap/anime/mal"
	"database/sql"
	_ "embed"
	"errors"
	"log"
	"log/slog"
	"os"
	"strings"
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

var DB *sql.DB

var File *os.File

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

type Scraper interface {
	Run()
	inserter()
	dumper()
}

func New(opt Options) Scraper {
	var done []int
	var err error
	if opt.Env {
		err := godotenv.Load()
		if err != nil {
			slog.Error(err.Error())
		}
		opt.Conn = os.Getenv("MALSQL_DB")
	}
	if !opt.File {
		DB, err = sql.Open(opt.Driver, opt.Conn)
		if err != nil {
			log.Panic(err)
		}
		schema := sqliteSchema
		if strings.Contains(opt.Driver, "postgres") {
			schema = pgSchema
		} else {
			if _, err := os.Stat("/path/to/whatever"); errors.Is(err, os.ErrNotExist) {
				f, err := os.Create(opt.Conn)
				if err != nil {
					log.Panic(err)
				}
				f.Close()
			}
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
		schema := sqliteSchema
		if strings.Contains(opt.Driver, "postgres") {
			schema = pgSchema
		}
		_, err = File.WriteString(schema)
		if err != nil {
			log.Panic(err)
		}
	}

	return &goodScrap{
		Options: opt,
		wg:      sync.WaitGroup{},
		animes:  make(chan []*anime.Anime),
		done:    done,
	}
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
