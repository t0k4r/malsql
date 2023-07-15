package anime

import (
	"MalSql/scrap/anime/gogo"
	"MalSql/scrap/anime/mal"
	"MalSql/scrap/anime/qb"
	"fmt"
	"strings"
	"time"
)

type Episode struct {
	mal.Episode
	index int
	url   string
	src   string
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
		return nil, err
	}
	anime := &Anime{Anime: ma}
	ep, err := mal.GetEpisodes(anime.MalUrl)
	if err != nil {
		return anime, err
	}
	es, err := gogo.GetEpisodes(anime.Title)
	if err != nil && err != gogo.ErrGoGo404 {
		return anime, err
	}
	anime.joinEpisodes(ep, es)
	anime.filterInfos()
	return anime, nil
}

func (a *Anime) joinEpisodes(ep []mal.Episode, es []string) {
	if len(ep) >= len(es) {
		for i, e := range ep {
			episode := Episode{Episode: e, src: "www3.gogoanimes.fi", index: i}
			episode.url = getOrEmpty(es, i)
			a.episodes = append(a.episodes, episode)
		}
	} else {
		for i, e := range es {
			episode := Episode{url: e, src: "www3.gogoanimes.fi", index: i}
			if i < len(ep) {
				episode.Episode = ep[i]
			}
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

func (a *Anime) Sql() (anime []string, relations []string) {
	var animeSql []string
	animeSql = append(animeSql, qb.
		Insert("anime_types").
		Str("type_of", a.typeOf).
		Sql())
	animeSql = append(animeSql, qb.
		Insert("seasons").
		Str("season", a.season).
		Str("value", a.seasonDate).
		Sql())
	animeSql = append(animeSql, qb.
		Insert("animes").
		Str("title", a.Title).
		Str("mal_url", a.MalUrl).
		Str("cover", a.ImgUrl).
		Str("aired_from", getOrEmpty(a.aired, 0)).
		Str("aired_to", getOrEmpty(a.aired, 1)).
		SubQ("type_id", `select id from anime_types where type_of = '%v'`, a.typeOf).
		SubQ("season_id", `select id from seasons where season = '%v'`, a.season).
		Sql())
	for _, title := range a.altTitles {
		animeSql = append(animeSql, qb.
			Insert("alt_title_types").
			Str("type_of", title.lang).
			Sql())
		animeSql = append(animeSql, qb.
			Insert("alt_titles").
			SubQ("alt_title_type_id", `select id from alt_title_types where type_of = '%v'`, title.lang).
			SubQ("anime_id", `select id from animes where mal_url = '%v'`, a.MalUrl).
			Str("alt_title", title.title).
			Sql())
	}
	for _, info := range a.Information {
		animeSql = append(animeSql, qb.
			Insert("info_types").
			Str("type_of", info.Key).
			Sql())
		animeSql = append(animeSql, qb.
			Insert("infos").
			Str("info", info.Value).
			Sql())
		animeSql = append(animeSql, qb.
			Insert("anime_infos").
			SubQ("anime_id", `select id from animes where mal_url = '%v'`, a.MalUrl).
			SubQ("info_id", `select id from infos where info = '%v'`, info.Value).
			Sql())
	}
	for _, episode := range a.episodes {
		animeSql = append(animeSql, qb.
			Insert("streamm_sources").
			Str("stream_source", episode.src).
			Sql())
		animeSql = append(animeSql, qb.
			Insert("episodes").
			Str("title", episode.Title).
			Str("alt_title", episode.AltTitle).
			Int("index_of", episode.index).
			SubQ("anime_id", `select id from animes where mal_url`, a.MalUrl).
			Sql())
		animeSql = append(animeSql, qb.
			Insert("episode_streams").
			Str("stream", episode.url).
			SubQRaw("episode_id", fmt.Sprintf(`
			select e.id from episodes e where e.anime_id = (select id from animes where mal_url = '%v') and e.index_of = %v`,
				a.MalUrl, episode.index)).
			SubQ("source_id", `select id from stream_sources where source = '%v'`, episode.src).
			Sql())
	}
	var relationsSql []string
	for _, r := range a.Related {
		relationsSql = append(relationsSql, qb.
			Insert("relation_types").
			Str("type_of", r.TypeOf).
			Sql())
		relationsSql = append(relationsSql, qb.
			Insert("relations").
			SubQ("root_anime_id", `select id from animes where mal_url = '%v'`, a.MalUrl).
			SubQ("related_anime_id", `select id from animes where mal_url = '%v'`, r.Url).
			SubQ("type_id", `select id from relation_types where type_of = '%v'`, r.TypeOf).
			Sql())
	}
	return animeSql, relationsSql
}

func getOrEmpty(arr []string, i int) string {
	if len(arr) > i {
		return arr[i]
	}
	return ""
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
