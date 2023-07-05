package anime

import (
	"MalSql/scrap/anime/gogo"
	"MalSql/scrap/anime/mal"
	"fmt"
	"strings"
	"time"
)

type Episode struct {
	Title string
	Url   string
}
type filInfo struct {
	typeOf     string
	season     string
	seasonDate string
	aired      []string
}

type Anime struct {
	mal.Anime
	Episodes []Episode
	filInf   filInfo
}

func LoadAnime[T string | int](malUrl T) (*Anime, error) {
	malAnime, err := mal.LoadAnime(malUrl)
	if err == mal.ErrMal429 {
		fmt.Println(err)
		mal.FixBlock()
		return LoadAnime(malUrl)
	} else if err != nil {
		return nil, err
	}
	episodes, err := loadEpisodes(malAnime.MalUrl, malAnime.Title)
	if err != nil {
		return nil, err
	}
	return &Anime{
		Anime:    *malAnime,
		Episodes: episodes,
	}, nil
}

func (a *Anime) filter() {
	var filterd filInfo
	var minInfo []mal.Info
	for _, i := range a.Information {
		switch i.Key {
		case "type":
			filterd.typeOf = i.Value
		case "premiered":
			filterd.season = i.Value
			sds := strings.Split(strings.Trim(i.Value, " \n"), " ")
			if len(sds) == 2 {
				t, err := time.Parse("2006", sds[1])
				if err == nil {
					switch sds[0] {
					case "Spring":
						t = t.AddDate(0, 2, 20)
					case "Summer":
						t = t.AddDate(0, 5, 21)
					case "Fall":
						t = t.AddDate(0, 8, 23)
					case "Winter":
						t = t.AddDate(0, 11, 22)
					}
					filterd.seasonDate = t.Format("2006-01-02")
				}
			}
		case "aired":
			for _, i := range strings.Split(i.Value, "to") {
				i = strings.Trim(i, " \n")
				switch len(i) {
				case 4:
					t, err := time.Parse("2006", i)
					if err == nil {
						filterd.aired = append(filterd.aired, t.Format("2006-01-02"))
					}
				case 8:
					t, err := time.Parse("Jan 2006", i)
					if err == nil {
						filterd.aired = append(filterd.aired, t.Format("2006-01-02"))
					}
				case 11, 12:
					t, err := time.Parse("Jan 2, 2006", i)
					if err == nil {
						filterd.aired = append(filterd.aired, t.Format("2006-01-02"))
					}
				}
			}
		default:
			minInfo = append(minInfo, i)
		}
	}
	a.filInf = filterd
	a.Information = minInfo
}

func loadEpisodes(malUrl string, title string) ([]Episode, error) {
	var episodes []Episode
	titles, err := mal.GetEpisodes(malUrl)
	if err == mal.ErrMal429 {
		fmt.Println(err)
		mal.FixBlock()
		return loadEpisodes(malUrl, title)
	} else if err != nil {
		return episodes, err
	}
	urls, err := gogo.GetEpisodes(title)
	if err != gogo.ErrGoGo404 && err != nil {
		return nil, err
	}
	if len(titles) < len(urls) {
		for i, url := range urls {
			episodes = append(episodes, Episode{Title: getOrEmpty(titles, i), Url: url})
		}
	} else {
		for i, title := range titles {
			episodes = append(episodes, Episode{Title: title, Url: getOrEmpty(urls, i)})
		}
	}
	return episodes, nil
}
