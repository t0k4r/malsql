package anime

import (
	"MalSql/scrap/anime/mal"
	"log"
	"sync"

	"golang.org/x/exp/slices"
)

type Relation struct {
	Root    string
	Related string
	TypeOf  string
}

type Series struct {
	Done      []int
	Animes    []Anime
	Relations []Relation
}

func NewSeries(anime *Anime) *Series {
	var series Series
	series.load(anime)
	return &series
}

func (s *Series) load(anime *Anime) {
	if slices.Contains(s.Done, anime.MagicNumber()) {
		return
	}
	s.Done = append(s.Done, anime.MagicNumber())
	s.Animes = append(s.Animes, *anime)
	s.pushRelations(anime)
	var wg sync.WaitGroup
	for _, r := range anime.Related {
		if !slices.Contains(s.Done, mal.MagicNumber(r.Url)) {
			wg.Add(1)
			go func(r mal.Related) {
				if !slices.Contains(s.Done, mal.MagicNumber(r.Url)) {
					rel, err := LoadAnime(r.Url)
					if err != mal.ErrMal404 {
						if err != nil {
							log.Panic(err)
						}
						s.load(rel)
					}
				}
				wg.Done()
			}(r)
		}
	}
	wg.Wait()
}

func (s *Series) pushRelations(anime *Anime) {
	for _, r := range anime.Related {
		s.Relations = append(s.Relations, Relation{
			Root:    anime.MalUrl,
			TypeOf:  r.TypeOf,
			Related: r.Url,
		})
	}
}

func (s *Series) Sql() []string {
	var sql []string
	for _, anime := range s.Animes {
		anime.filter()
		sql = append(sql, animeSql(anime)...)
	}
	for _, r := range s.Relations {
		sql = append(sql, relationsSql(r)...)
	}
	return sql
}

func getOrEmpty(arr []string, i int) string {
	if len(arr) > i {
		return arr[i]
	}
	return ""
}
