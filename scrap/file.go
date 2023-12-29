package scrap

import (
	"MalSql/scrap/anime"
	"log/slog"
	"os"
	"sync"

	"github.com/t0k4r/qb"
)

type fSaver struct {
	wg         sync.WaitGroup
	onConflict qb.OnConflict
	file       *os.File
}

func newFSaver(opts Options) (*fSaver, error) {
	fsv := fSaver{
		wg:         sync.WaitGroup{},
		onConflict: opts.onConflict(),
	}
	var err error
	fsv.file, err = os.Create(opts.Conn)
	if err != nil {
		return &fsv, err
	}
	switch opts.Driver {
	case "sqlite3":
		_, err = fsv.file.WriteString(sqliteSchema)
	case "postgres":
		_, err = fsv.file.WriteString(pgSchema)
	}
	return &fsv, err
}

func (f *fSaver) listen(schan chan []*anime.Anime) {
	f.wg.Add(1)
	go func() {
		for animes := range schan {
			var relations []string
			for _, anime := range animes {
				asql, rsql := anime.Sql()
				for _, anime := range asql {
					_, err := f.file.WriteString(anime.
						OnConflict(f.onConflict).
						Sql() + ";\n")
					if err != nil {
						slog.Error(err.Error())
					}
				}
				for _, relation := range rsql {
					relations = append(relations, relation.
						OnConflict(f.onConflict).
						Sql()+";\n")
				}
			}
			for _, sql := range relations {
				_, err := f.file.WriteString(sql)
				if err != nil {
					slog.Error(err.Error())
				}
			}

		}
		f.wg.Done()
	}()
}

func (f *fSaver) wait() {
	f.wg.Wait()
}
