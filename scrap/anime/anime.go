package anime

import (
	"MalSql/scrap/anime/gogo"
	"MalSql/scrap/anime/mal"
	oqb "MalSql/scrap/anime/qb"
	"fmt"
	"strings"
	"time"

	"github.com/t0k4r/qb"
	"golang.org/x/sync/errgroup"
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
	Episodes   []Episode
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
	g := new(errgroup.Group)
	var EpTitles []mal.Episode
	g.Go(func() error {
		ep, err := mal.GetEpisodes(anime.MalUrl)
		EpTitles = ep
		return err
	})
	var EpStreams []string
	g.Go(func() error {
		es, err := gogo.GetEpisodes(anime.Title)
		EpStreams = es
		if err == gogo.ErrGoGo404 {
			return nil
		}
		return err
	})
	anime.filterInfos()
	err = g.Wait()
	anime.joinEpisodes(EpTitles, EpStreams)
	return anime, err
}

func (a *Anime) joinEpisodes(ep []mal.Episode, es []string) {
	if len(ep) >= len(es) {
		a.Episodes = make([]Episode, len(ep))
		for i, e := range ep {
			episode := Episode{Episode: e, src: "www3.gogoanimes.fi", index: i}
			episode.url = getOrEmpty(es, i)
			a.Episodes[i] = episode
		}
	} else {
		a.Episodes = make([]Episode, len(es))
		for i, e := range es {
			episode := Episode{url: e, src: "www3.gogoanimes.fi", index: i}
			if i < len(ep) {
				episode.Episode = ep[i]
			}
			a.Episodes[i] = episode
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
			txt := info.Value
			txt = strings.Join(strings.Fields(txt), " ")
			info.Value = txt
			filtered = append(filtered, info)
		}
	}
	a.Information = filtered
}

func (a *Anime) Sql() (anime []string, relations []string) {
	var animeSql []string

	animeSql = append(animeSql, oqb.
		Insert("anime_types").
		Str("type_of", a.typeOf).
		Sql())
	animeSql = append(animeSql, oqb.
		Insert("seasons").
		Str("season", a.season).
		Str("value", a.seasonDate).
		Sql())
	animeSql = append(animeSql, oqb.
		Insert("animes").
		Str("title", a.Title).
		Str("description", a.Description).
		Str("mal_url", a.MalUrl).
		Str("cover", a.ImgUrl).
		Str("aired_from", getOrEmpty(a.aired, 0)).
		Str("aired_to", getOrEmpty(a.aired, 1)).
		SubQ("type_id", `select id from anime_types where type_of = '%v'`, a.typeOf).
		SubQ("season_id", `select id from seasons where season = '%v'`, a.season).
		Sql())
	for _, title := range a.altTitles {
		animeSql = append(animeSql, oqb.
			Insert("alt_title_types").
			Str("type_of", title.lang).
			Sql())
		animeSql = append(animeSql, oqb.
			Insert("alt_titles").
			SubQ("alt_title_type_id", `select id from alt_title_types where type_of = '%v'`, title.lang).
			SubQ("anime_id", `select id from animes where mal_url = '%v'`, a.MalUrl).
			Str("alt_title", title.title).
			Sql())
	}
	for _, info := range a.Information {
		animeSql = append(animeSql, oqb.
			Insert("info_types").
			Str("type_of", info.Key).
			Sql())
		animeSql = append(animeSql, oqb.
			Insert("infos").
			Str("info", info.Value).
			SubQ("type_id", `select id from info_types where type_of = '%v'`, info.Key).
			Sql())
		animeSql = append(animeSql, oqb.
			Insert("anime_infos").
			SubQ("anime_id", `select id from animes where mal_url = '%v'`, a.MalUrl).
			SubQ("info_id", `select id from infos where info = '%v'`, info.Value).
			Sql())
	}
	for _, episode := range a.Episodes {
		animeSql = append(animeSql, oqb.
			Insert("stream_sources").
			Str("stream_source", episode.src).
			Sql())
		animeSql = append(animeSql, oqb.
			Insert("episodes").
			Str("title", episode.Title).
			Str("alt_title", episode.AltTitle).
			Int("index_of", episode.index).
			SubQ("anime_id", `select id from animes where mal_url = '%v'`, a.MalUrl).
			Sql())
		if episode.url != "" {
			animeSql = append(animeSql, oqb.
				Insert("episode_streams").
				Str("stream", episode.url).
				SubQRaw("episode_id", fmt.Sprintf(`
			select e.id from episodes e where e.anime_id = (select id from animes where mal_url = '%v') and e.index_of = %v
			`, a.MalUrl, episode.index)).
				SubQ("source_id", `select id from stream_sources where stream_source = '%v'`, episode.src).
				Sql())
		}
	}
	var relationsSql []string

	for _, r := range a.Related {
		relationsSql = append(relationsSql, oqb.
			Insert("relation_types").
			Str("type_of", r.TypeOf).
			Sql())
		relationsSql = append(relationsSql, oqb.
			Insert("relations").
			SubQ("root_anime_id", `select id from animes where mal_url = '%v'`, a.MalUrl).
			SubQ("related_anime_id", `select id from animes where mal_url = '%v'`, r.Url).
			SubQ("type_id", `select id from relation_types where type_of = '%v'`, r.TypeOf).
			Sql())
	}
	return animeSql, relationsSql
}

// anime, relations
func (a *Anime) Nsql() ([]qb.QInsert, []qb.QInsert) {
	var anime []qb.QInsert
	anime = append(anime, qb.
		Insert("anime_types").
		Col("type_of", a.typeOf))
	anime = append(anime, qb.
		Insert("seasons").
		Col("season", a.season).
		Col("value", a.seasonDate))
	anime = append(anime, qb.
		Insert("animes").
		Col("title", a.Title).
		Col("description", a.Description).
		Col("mal_url", a.MalUrl).
		Col("cover", a.ImgUrl).
		Col("aired_from", getOrEmpty(a.aired, 0)).
		Col("aired_to", getOrEmpty(a.aired, 1)).
		Col("type_id", qb.
			Select("anime_types").
			Cols("id").
			Wheref("type_of = '%v'", a.typeOf)).
		Col("season_id", qb.
			Select("seasons").
			Cols("id").
			Wheref("season = '%v'", a.season)))
	for _, title := range a.altTitles {
		anime = append(anime, qb.
			Insert("alt_title_types").
			Col("type_of", title.lang))
		anime = append(anime, qb.
			Insert("alt_titles").
			Col("alt_title_type_id", qb.
				Select("alt_title_types").
				Cols("id").
				Wheref("type_of = '%v'", title.lang)).
			Col("anime_id", qb.
				Select("animes").
				Cols("id").
				Wheref("mal_url = '%v'", a.MalUrl)).
			Col("alt_title", title.title))
	}
	for _, info := range a.Information {
		anime = append(anime, qb.
			Insert("info_types").
			Col("type_of", info.Key))
		anime = append(anime, qb.
			Insert("infos").
			Col("info", info.Value).
			Col("type_id", qb.
				Select("info_types").
				Cols("id").
				Wheref("type_of = '%v'", info.Key)))
		anime = append(anime, qb.
			Insert("anime_infos").
			Col("anime_id", qb.
				Select("animes").
				Cols("id").
				Wheref("mal_url = '%v'", a.MalUrl)).
			Col("info_id", qb.
				Select("infos").
				Cols("id").
				Wheref("info = '%v'", info.Value)))
	}
	for _, episode := range a.Episodes {
		anime = append(anime, qb.
			Insert("stream_sources").
			Col("stream_source", episode.src))
		anime = append(anime, qb.
			Insert("episodes").
			Col("title", episode.Title).
			Col("alt_title", episode.AltTitle).
			Col("index_of", episode.index).
			Col("anime_id", qb.
				Select("animes").
				Cols("id").
				Wheref("mal_url = '%v'", a.MalUrl)))
		if episode.url != "" {
			anime = append(anime, qb.
				Insert("episode_streams").
				Col("stream", episode.url).
				Col("episode_id", qb.
					Select("episodes e").
					Wheref("e.anime_id = (select id from animes where mal_url = '%v') and e.index_of = %v",
						a.MalUrl, episode.index)).
				Col("source_id", qb.
					Select("stream_sources").
					Cols("id").
					Wheref("stream_source = '%v'", episode.src)))
		}
	}
	var relations []qb.QInsert
	for _, r := range a.Related {
		relations = append(relations, qb.
			Insert("relation_types").
			Col("type_of", r.TypeOf))
		relations = append(relations, qb.
			Insert("relations").
			Col("root_anime_id", qb.
				Select("animes").
				Cols("id").
				Wheref("mal_url = '%v'", a.MalUrl)).
			Col("related_anime_id", qb.
				Select("animes").
				Cols("id").
				Wheref("mal_url = '%v'", r.Url)).
			Col("type_id", qb.
				Select("relation_types").
				Cols("id").
				Wheref("type_of = '%v'", r.TypeOf)))
	}
	return anime, relations
}

func getOrEmpty(arr []string, i int) string {
	if len(arr) > i {
		return arr[i]
	}
	return ""
}
