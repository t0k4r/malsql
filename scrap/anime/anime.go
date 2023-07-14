package anime

import (
	"MalSql/scrap/anime/gogo"
	"MalSql/scrap/anime/mal"
	"strings"
	"time"
)

type Episode struct {
	mal.Episode
	Url string
}

type Title struct {
	lang  string
	title string
}

type Anime struct {
	*mal.Anime
	episodes   []Episode
	typeOf     string
	season     string
	seasonDate string
	aired      []string
	altTitles  []Title
}

func LoadAnime[T string | int](malUrl T) (*Anime, error) {
	ma, err := mal.LoadAnime(malUrl)
	if err != nil {
		panic(err)
	}
	anime := &Anime{Anime: ma}
	ep, err := mal.GetEpisodes(anime.MalUrl)
	if err != nil {
		panic(err)
	}
	es, err := gogo.GetEpisodes(anime.Title)
	if err != nil && err != gogo.ErrGoGo404 {
		panic(err)
	}
	anime.joinEpisodes(ep, es)
	anime.filterInfos()
	return anime, nil
}

func (a *Anime) joinEpisodes(ep []mal.Episode, es []string) {
	if len(ep) >= len(es) {
		for i, e := range ep {
			episode := Episode{Episode: e}
			episode.Url = getOrEmpty(es, i)
			a.episodes = append(a.episodes, episode)
		}
	} else {
		for _, e := range es {
			episode := Episode{Url: e}
			a.episodes = append(a.episodes, episode)
		}

	}
}

func (a *Anime) filterInfos() {
	var filtered []mal.Info
	for _, info := range a.Information {
		switch info.Key {
		case "type":
			a.typeOf = info.Value
		case "premiered":
			a.season = info.Value
			sds := strings.Split(strings.Trim(info.Value, " \n"), " ")
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
					a.seasonDate = t.Format("2006-01-02")
				}
			}
		case "aired":
			for _, i := range strings.Split(info.Value, "to") {
				i = strings.Trim(i, " \n")
				switch len(i) {
				case 4:
					t, err := time.Parse("2006", i)
					if err == nil {
						a.aired = append(a.aired, t.Format("2006-01-02"))
					}
				case 8:
					t, err := time.Parse("Jan 2006", i)
					if err == nil {
						a.aired = append(a.aired, t.Format("2006-01-02"))
					}
				case 11, 12:
					t, err := time.Parse("Jan 2, 2006", i)
					if err == nil {
						a.aired = append(a.aired, t.Format("2006-01-02"))
					}
				}
			}
		case "synonyms", "japanese", "english", "german", "french", "spanish":
			title := Title{lang: info.Key, title: info.Value}
			a.altTitles = append(a.altTitles, title)
		case "theme", "genre", "demographic", "producer", "licensor", "studio":
			info.Key += "s"
			filtered = append(filtered, info)
		default:
			filtered = append(filtered, info)
		}
	}
	a.Information = filtered
}

// type Episode struct {
// 	mal.Episode
// 	Url string
// }
// type filInfo struct {
// 	typeOf     string
// 	season     string
// 	seasonDate string
// 	aired      []string
// }

// type Anime struct {
// 	mal.Anime
// 	Episodes []Episode
// 	filInf   filInfo
// }

// func LoadAnime[T string | int](malUrl T) (*Anime, error) {
// 	malAnime, err := mal.LoadAnime(malUrl)
// 	if err == mal.ErrMal429 {
// 		fmt.Println(err)
// 		mal.FixBlock()
// 		return LoadAnime(malUrl)
// 	} else if err != nil {
// 		return nil, err
// 	}
// 	episodes, err := loadEpisodes(malAnime.MalUrl, malAnime.Title)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &Anime{
// 		Anime:    *malAnime,
// 		Episodes: episodes,
// 	}, nil
// }

// func (a *Anime) filter() {
// 	var filterd filInfo
// 	var minInfo []mal.Info
// 	for _, i := range a.Information {
// 		switch i.Key {
// 		case "type":
// 			filterd.typeOf = i.Value
// 		case "premiered":
// 			filterd.season = i.Value
// 			sds := strings.Split(strings.Trim(i.Value, " \n"), " ")
// 			if len(sds) == 2 {
// 				t, err := time.Parse("2006", sds[1])
// 				if err == nil {
// 					switch sds[0] {
// 					case "Spring":
// 						t = t.AddDate(0, 2, 20)
// 					case "Summer":
// 						t = t.AddDate(0, 5, 21)
// 					case "Fall":
// 						t = t.AddDate(0, 8, 23)
// 					case "Winter":
// 						t = t.AddDate(0, 11, 22)
// 					}
// 					filterd.seasonDate = t.Format("2006-01-02")
// 				}
// 			}
// 		case "aired":
// 			for _, i := range strings.Split(i.Value, "to") {
// 				i = strings.Trim(i, " \n")
// 				switch len(i) {
// 				case 4:
// 					t, err := time.Parse("2006", i)
// 					if err == nil {
// 						filterd.aired = append(filterd.aired, t.Format("2006-01-02"))
// 					}
// 				case 8:
// 					t, err := time.Parse("Jan 2006", i)
// 					if err == nil {
// 						filterd.aired = append(filterd.aired, t.Format("2006-01-02"))
// 					}
// 				case 11, 12:
// 					t, err := time.Parse("Jan 2, 2006", i)
// 					if err == nil {
// 						filterd.aired = append(filterd.aired, t.Format("2006-01-02"))
// 					}
// 				}
// 			}
// 		default:
// 			minInfo = append(minInfo, i)
// 		}
// 	}
// 	a.filInf = filterd
// 	a.Information = minInfo
// }

// func loadEpisodes(malUrl string, title string) ([]Episode, error) {
// 	var episodes []Episode
// 	titles, err := mal.GetEpisodes(malUrl)
// 	if err == mal.ErrMal429 {
// 		fmt.Println(err)
// 		mal.FixBlock()
// 		return loadEpisodes(malUrl, title)
// 	} else if err != nil {
// 		return episodes, err
// 	}
// 	urls, err := gogo.GetEpisodes(title)
// 	if err != gogo.ErrGoGo404 && err != nil {
// 		return nil, err
// 	}
// 	if len(titles) < len(urls) {
// 		for i, url := range urls {
// 			episodes = append(episodes, Episode{Title: getOrEmpty(titles, i), Url: url})
// 		}
// 	} else {
// 		for i, title := range titles {
// 			episodes = append(episodes, Episode{Title: title, Url: getOrEmpty(urls, i)})
// 		}
// 	}
// 	return episodes, nil
// }
