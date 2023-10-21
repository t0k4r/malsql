package scrap

import (
	"MalSql/scrap/anime"
	"MalSql/scrap/anime/mal"
	"log/slog"
	"slices"
	"sync"
	"time"
)

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
			for _, r := range rsql {
				relations = append(relations, r.Sql(s.onConflict())+";\n")
			}
			for _, a := range asql {
				_, err := File.WriteString(a.Sql(s.onConflict()) + ";\n")
				if err != nil {
					slog.Error(err.Error())
				}
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
		var relations []string
		for _, anime := range animes {
			asql, rsql := anime.Sql()
			for _, r := range rsql {
				relations = append(relations, r.Sql(s.onConflict()))
			}
			for _, a := range asql {
				sql := a.Sql(s.onConflict())
				_, err := DB.Exec(sql)
				if err != nil {
					slog.Error(err.Error())
				}
			}
		}
		for _, rsql := range relations {
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
		var series goodSeries
		series.load(anime)
		s.done = append(s.done, series.done...)
		s.animes <- series.animes
	}
	close(s.animes)
	s.wg.Wait()
	slog.Info("Done", "animes", len(s.done), "took", time.Since(n))
}

type goodSeries struct {
	done   []int
	animes []*anime.Anime
}

func (s *goodSeries) load(root *anime.Anime) {
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
