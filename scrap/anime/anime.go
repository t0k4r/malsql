package anime

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/t0k4r/qb"
	"golang.org/x/sync/errgroup"
)

type Anime struct {
	errgroup.Group
	Title       string
	Description string
	MalUrl      string
	ImgUrl      string
	Information map[string][]string
	Related     map[string][]string
	Episodes    []Episode
}

func LoadAnime[T int | string](url T) (*Anime, error) {
	anime := Anime{
		Group:       errgroup.Group{},
		Title:       "",
		Description: "",
		MalUrl:      "",
		ImgUrl:      "",
		Information: map[string][]string{},
		Related:     map[string][]string{},
		Episodes:    []Episode{},
	}
	switch any(url).(type) {
	case string:
		anime.MalUrl = fmt.Sprintf("%v", url)
	case int:
		anime.MalUrl = fmt.Sprintf("https://myanimelist.net/anime/%v/co_ty_tutaj_robisz", url)
	}
	res, err := http.Get(anime.MalUrl)
	if err != nil {
		return nil, err
	}
	switch res.StatusCode {
	case 404:
		return nil, nil
	case 429:
		FixBlock()
		return LoadAnime(anime.Title)
	case 200:
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, err
		}
		anime.Title = doc.Find(`.title-name`).Text()
		anime.Go(anime.LoadEpisodes)
		anime.Description = doc.Find(`p[itemprop="description"]`).Text()
		doc.Find("div.breadcrumb").ChildrenFiltered("div.di-ib").Each(func(i int, s *goquery.Selection) {
			if i == 2 {
				anime.MalUrl = s.ChildrenFiltered("a").AttrOr("href", "")
			}
		})
		anime.ImgUrl = doc.Find(`img[itemprop="image"]`).AttrOr("data-src", "")
		doc.Find(`div.leftside>div.spaceit_pad`).Each(func(i int, s *goquery.Selection) {
			key := strings.Replace(strings.ToLower(s.ChildrenFiltered(`span.dark_text`).Text()), ":", "", 1)
			var values []string
			s.ChildrenFiltered(`a`).Each(func(i int, s *goquery.Selection) {
				values = append(values, s.Text())
			})
			if len(values) == 0 {
				values = append(values, strings.Trim(strings.Join(strings.Split(s.Text(), ":")[1:], ":"), "\n "))
			}
			for _, v := range values {
				vals, ok := anime.Information[key]
				if ok {
					anime.Information[key] = append(vals, v)
				} else {
					anime.Information[key] = []string{v}
				}
			}
		})
		doc.Find(`.js-alternative-titles>.spaceit_pad`).Each(func(i int, s *goquery.Selection) {
			key := strings.Replace(strings.ToLower(s.ChildrenFiltered(`span.dark_text`).Text()), ":", "", 1)
			v := strings.Trim(strings.Join(strings.Split(s.Text(), ":")[1:], ":"), "\n ")
			vals, ok := anime.Information[key]
			if ok {
				anime.Information[key] = append(vals, v)
			} else {
				anime.Information[key] = []string{v}
			}
		})
		doc.Find(".anime_detail_related_anime>tbody>tr").Each(func(i int, s *goquery.Selection) {
			key := ""
			s.ChildrenFiltered("td").Each(func(i int, s *goquery.Selection) {
				if i == 0 {
					key = strings.ToLower(s.Text())
					key = strings.Split(key, ":")[0]
				} else {
					s.ChildrenFiltered("a").Each(func(i int, s *goquery.Selection) {
						v := fmt.Sprintf("https://myanimelist.net%v", s.AttrOr("href", ""))
						vals, ok := anime.Related[key]
						if ok {
							anime.Related[key] = append(vals, v)
						} else {
							anime.Related[key] = []string{v}
						}
					})
				}
			})
		})
		delete(anime.Related, "adaptation")
		return &anime, anime.Wait()
	default:
		panic(fmt.Sprintf("unknown status code %v", res.StatusCode))
	}
}
func (a *Anime) LoadEpisodes() error {
	var src string
	var streams []string
	var episodes []Episode
	wg := new(errgroup.Group)
	wg.Go(func() error {
		var err error
		src, streams, err = LoadEpisodeStreams(a.Title)
		return err
	})
	wg.Go(func() error {
		var err error
		episodes, err = LoadEpisodes(a.MalUrl)
		return err
	})
	err := wg.Wait()
	if err != nil {
		return err
	}
	a.Episodes = episodes
	if len(a.Episodes) >= len(episodes) {
		for i := range a.Episodes {
			a.Episodes[i].Index = i
			a.Episodes[i].StreamSrc = src
			if i < len(streams) {
				a.Episodes[i].StreamUrl = streams[i]
			}
		}
	} else {
		for i, stream := range streams {
			if i < len(a.Episodes) {
				a.Episodes[i].Index = i
				a.Episodes[i].StreamSrc = src
				a.Episodes[i].StreamUrl = stream
			} else {
				a.Episodes = append(a.Episodes, Episode{
					Index:     i,
					Title:     "",
					AltTitle:  "",
					StreamSrc: src,
					StreamUrl: stream,
				})
			}
		}
	}
	return nil

}

func (a *Anime) Sql() ([]*qb.QInsert, []*qb.QInsert) {
	var animeSql []*qb.QInsert
	var relationsSql []*qb.QInsert
	for k, v := range a.Information {
		for i := range v {
			v[i] = strings.Join(strings.Fields(v[i]), " ")
		}
		a.Information[k] = v
	}
	TypeOf, ok := a.Information["type"]
	if ok {
		animeSql = append(animeSql, qb.
			Insert("anime_types").
			Set("type_of", TypeOf[0]))
	} else {
		TypeOf = append(TypeOf, "")
	}
	Season, ok := a.Information["premiered"]
	if ok {
		sds := strings.Split(Season[0], " ")
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
				animeSql = append(animeSql, qb.Insert("seasons").
					Set("season", Season[0]).
					Set("value", t.Format("2006-01-02")))
			}
		}
	} else {
		Season = append(Season, "")
	}
	Aired, ok := a.Information["aired"]
	if ok {
		NewAired := make([]string, 2)
		for _, i := range strings.Split(Aired[0], "to") {
			i = strings.Trim(i, " \n")
			switch len(i) {
			case 4:
				t, err := time.Parse("2006", i)
				if err == nil {
					NewAired = append(NewAired, t.Format("2006-01-02"))
				}
			case 8:
				t, err := time.Parse("Jan 2006", i)
				if err == nil {
					NewAired = append(NewAired, t.Format("2006-01-02"))
				}
			case 11, 12:
				t, err := time.Parse("Jan 2, 2006", i)
				if err == nil {
					NewAired = append(NewAired, t.Format("2006-01-02"))
				}
			}
		}
		for len(NewAired) < 2 {
			NewAired = append(NewAired, "")
		}
		Aired = NewAired
	} else {
		Aired = append(Aired, "")
		Aired = append(Aired, "")
	}
	animeSql = append(animeSql, qb.
		Insert("animes").
		Set("title", a.Title).
		Set("description", a.Description).
		Set("mal_url", a.MalUrl).
		Set("cover", a.ImgUrl).
		Set("aired_from", Aired[0]).
		Set("aired_to", Aired[1]).
		Setf("type_id",
			"(select id from anime_types where type_of = '%v')", TypeOf[0]).
		Setf("season_id",
			"(select id from seasons where season = '%v')", Season[0]))

	for key, values := range a.Information {
		switch key {
		case "type", "premiered", "aired":
			continue
		case "synonyms", "japanese", "english", "german", "french", "spanish":
			for _, value := range values {
				animeSql = append(animeSql, qb.
					Insert("alt_title_types").
					Set("type_of", key))
				animeSql = append(animeSql, qb.
					Insert("alt_titles").
					Setf("alt_title_type_id",
						"(select id from alt_title_types where type_of = '%v')", key).
					Setf("anime_id",
						"(select id from animes where mal_url = '%v')", a.MalUrl).
					Set("alt_title", value))
			}
			continue
		case "theme", "genre", "demographic", "producer", "licensor", "studio":
			key += "s"
		}
		for _, value := range values {
			animeSql = append(animeSql, qb.
				Insert("info_types").
				Set("type_of", key))
			animeSql = append(animeSql, qb.
				Insert("infos").
				Set("info", value).
				Setf("type_id",
					"(select id from info_types where type_of = '%v')", key))
			animeSql = append(animeSql, qb.
				Insert("anime_infos").
				Setf("anime_id",
					"(select id from animes where mal_url = '%v')", a.MalUrl).
				Setf("info_id",
					"(select id from infos where info = '%v')", strings.ReplaceAll(value, "'", "''")))
		}
	}
	for _, episode := range a.Episodes {
		animeSql = append(animeSql, qb.
			Insert("stream_sources").
			Set("stream_source", episode.StreamSrc))
		animeSql = append(animeSql, qb.
			Insert("episodes").
			Set("title", episode.Title).
			Set("alt_title", episode.AltTitle).
			Set("index_of", episode.Index).
			Setf("anime_id",
				"(select id from animes where mal_url = '%v')", a.MalUrl))
		if episode.StreamUrl != "" {
			animeSql = append(animeSql, qb.
				Insert("episode_streams").
				Set("stream", episode.StreamUrl).
				Setf("episode_id",
					"(select e.id from episodes e where e.anime_id = (select id from animes where mal_url = '%v' and e.index_of = %v))",
					a.MalUrl, episode.Index).
				Setf("source_id",
					"(select id from stream_sources where stream_source = '%v')", episode.StreamSrc))
		}
	}

	return animeSql, relationsSql
}

func (a *Anime) MagicNumber() int {
	return MagicNumber(a.MalUrl)
}

func MagicNumber(url string) int {
	i, _ := strconv.Atoi(strings.Split(url, "/")[4])
	return i
}
