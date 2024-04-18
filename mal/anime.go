package mal

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

type Url string

func UrlFromId(id int) Url {
	return Url(fmt.Sprintf("https://myanimelist.net/anime/%v/Bungou_Stray_Dogs_2nd_Season", id))
}

func (u Url) Id() int {
	i, _ := strconv.Atoi(strings.Split(string(u), "/")[4])
	return i
}

type Episode struct {
	Url      string
	Title    string
	TitleAlt string
	Descr    string
}

type Anime struct {
	MalUrl      Url
	ImgUrl      string
	Title       string
	Description string
	infos       map[string][]string
	Related     map[string][]Url
	Episodes    []Episode
}

func (a *Anime) Type() (string, bool) {
	t, ok := a.infos["type"]
	return strings.Join(t, ""), ok
}
func (a *Anime) Season() (string, time.Time, bool) {
	s, ok := a.infos["premiered"]
	if !ok {
		return "", time.Now(), false
	}
	str := strings.Join(s, "")
	var t time.Time
	var err error
	sds := strings.Split(strings.Trim(str, " \n"), " ")
	if len(sds) == 2 {
		t, err = time.Parse("2006", sds[1])
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
		}
	}
	return str, t, ok

}
func (a *Anime) Infos() map[string][]string {
	infos := make(map[string][]string)
	for k, v := range a.infos {
		switch k {
		case "synonyms", "japanese", "english", "german", "french", "spanish", "type", "premiered":
			continue
		case "theme", "genre", "demographic", "producer", "licensor", "studio":
			infos[k+"s"] = v
		default:
			infos[k] = v
		}
	}
	return infos
}

func (a *Anime) TitleAlt() map[string]string {
	titles := make(map[string]string)
	for k, v := range a.infos {
		if slices.Contains([]string{"synonyms", "japanese", "english", "german", "french", "spanish"}, k) {
			titles[k] = strings.Join(v, ", ")
		}
	}
	return titles
}

func (a *Anime) fetchEpisodes(url Url) func() error {
	return func() error {
		res, err := http.Get(string(url))
		if err != nil {
			return err
		}
		switch res.StatusCode {
		case 404:
			return nil
		case 429:
			fixLock(string(url))
			return a.fetchEpisodes(url)()
		default:
			if res.StatusCode != 200 {
				return fmt.Errorf("unknown statuscode on episode fetch %v", res.StatusCode)
			}
		}
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return err
		}
		_ = doc
		return nil
	}
}

func FetchAnime(url Url) (*Anime, error) {
	var anime Anime
	var eg errgroup.Group
	eg.Go(anime.fetchEpisodes(url))
	res, err := http.Get(string(url))
	if err != nil {
		return nil, err
	}
	switch res.StatusCode {
	case 404:
		return nil, nil
	case 429:
		fixLock(string(url))
		return FetchAnime(url)
	default:
		if res.StatusCode != 200 {
			return &anime, fmt.Errorf("unknown statuscode on anime fetch %v", res.StatusCode)
		}
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	anime.MalUrl = Url(doc.Find("div.breadcrumb").ChildrenFiltered("div.di-ib:nth-of-type(3)").ChildrenFiltered("a").AttrOr("href", ""))
	anime.ImgUrl = doc.Find(`img[itemprop="image"]`).AttrOr("data-src", "")
	anime.Title = doc.Find(`.title-name`).Text()
	anime.Description = doc.Find(`p[itemprop="description"]`).Text()
	anime.infos = readInfos(doc)
	anime.Related = readRelated(doc)
	return &anime, eg.Wait()
}

func readInfos(doc *goquery.Document) map[string][]string {
	infos := map[string][]string{}
	doc.Find(`.js-alternative-titles>.spaceit_pad`).Each(func(i int, s *goquery.Selection) {
		key := strings.Replace(strings.ToLower(s.ChildrenFiltered(`span.dark_text`).Text()), ":", "", 1)
		v := strings.Trim(strings.Join(strings.Split(s.Text(), ":")[1:], ":"), "\n ")
		vals, ok := infos[key]
		if ok {
			infos[key] = append(vals, v)
		} else {
			infos[key] = []string{v}
		}
	})

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
			v = strings.Join(strings.Fields(v), " ")
			vals, ok := infos[key]
			if ok {
				infos[key] = append(vals, v)
			} else {
				infos[key] = []string{v}
			}
		}
	})
	return infos
}
func readRelated(doc *goquery.Document) map[string][]Url {
	related := map[string][]Url{}
	doc.Find(".anime_detail_related_anime>tbody>tr").Each(func(i int, s *goquery.Selection) {
		key := ""
		s.ChildrenFiltered("td").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				key = strings.ToLower(s.Text())
				key = strings.Split(key, ":")[0]
			} else {
				s.ChildrenFiltered("a").Each(func(i int, s *goquery.Selection) {
					v := Url(fmt.Sprintf("https://myanimelist.net%v", s.AttrOr("href", "")))
					vals, ok := related[key]
					if ok {
						related[key] = append(vals, v)
					} else {
						related[key] = []Url{v}
					}
				})
			}
		})
	})
	delete(related, "adaptation")
	return related
}
