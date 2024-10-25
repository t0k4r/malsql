package malsql

import (
	"log"
	"log/slog"
	"malsql/mal"
	"os"
	"strings"
	"sync"

	"github.com/t0k4r/x/chanx"
	"github.com/t0k4r/x/iterx"
)

type Flags struct {
	Start int
	End   int
	Fast  bool
	File  string
}

func New(f Flags) {
	file, err := os.Create(f.File)
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan int)
	wg := sync.WaitGroup{}
	it := iterx.Uniq(chanx.All(c))
	go func() {
		wg.Add(1)
		for i := f.Start; i <= f.End; i++ {
			c <- i
		}
		wg.Done()
		wg.Wait()
		close(c)
	}()
	for url := range it {
		wg.Add(1)
		malAnime, err := mal.FetchAnime(mal.IdToUrl(url))
		if err != nil {
			slog.Error(err.Error())
		}
		if malAnime != nil {
			anime := Anime{Anime: *malAnime}
			aniemSql, _ := anime.Sql()
			if _, err = file.WriteString(strings.Join(aniemSql, ";\n")); err != nil {
				slog.Error(err.Error())
			}

			// fmt.Println(anime.Related)
			// wg.Add(1)
			// go func() {
			// 	for _, urls := range anime.Related {
			// 		for _, url := range urls {
			// 			c <- mal.UrlToId(url)
			// 		}
			// 	}
			// 	wg.Done()
			// }()
			slog.Info(anime.Title)
		}
		wg.Done()
	}
}
