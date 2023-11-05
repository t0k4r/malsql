package scrap

import (
	"MalSql/scrap/anime"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/t0k4r/qb"
)

type dSaver struct {
	db         *sql.DB
	wg         sync.WaitGroup
	onConflict qb.Conflict
}

func newDSaver(opts Options) (*dSaver, error) {
	dsv := dSaver{
		wg:         sync.WaitGroup{},
		onConflict: opts.onConflict(),
	}

	if opts.Driver == "sqlite3" {
		_, err := os.Stat(opts.Conn)
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(opts.Conn)
			if err != nil {
				return &dsv, err
			}
			f.Close()
		}
	}

	var err error
	dsv.db, err = sql.Open(opts.Driver, opts.Conn)
	if err != nil {
		return &dsv, err
	}
	switch opts.Driver {
	case "sqlite3":
		for _, table := range strings.Split(sqliteSchema, ";\n") {
			_, err = dsv.db.Exec(table)
			if err != nil {
				return &dsv, err
			}
		}
	case "postgres":
		for _, table := range strings.Split(pgSchema, ";\n") {
			_, err = dsv.db.Exec(table)
			if err != nil {
				return &dsv, err
			}
		}
	}
	return &dsv, nil
}

func (d *dSaver) listen(schan chan []*anime.Anime) {
	d.wg.Add(1)
	go func() {
		for animes := range schan {
			var relations []string
			for _, anime := range animes {
				asql, rsql := anime.Sql()
				for _, anime := range asql {
					_, err := d.db.Exec(anime.Sql(d.onConflict))
					if err != nil {
						slog.Error(err.Error())
					}
				}
				for _, relation := range rsql {
					relations = append(relations, relation.Sql(d.onConflict))
				}
			}
			for _, sql := range relations {
				_, err := d.db.Exec(sql)
				if err != nil {
					slog.Error(err.Error())
				}
			}

		}
		d.wg.Done()
	}()
}

func (d *dSaver) wait() {
	d.wg.Wait()
}

func (d *dSaver) skip() ([]int, error) {
	var done []int
	q, err := d.db.Query("select id from animes")
	if err != nil {
		return done, err
	}

	for q.Next() {
		var i int
		err = q.Scan(&i)
		if err != nil {
			return done, err
		}
		done = append(done, i)
	}

	return done, nil

}
