package mal

import (
	"fmt"
	"log"
	"net/http"
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

func (a *Anime) TypeOf() string {
	return ""
}
func (a *Anime) Infos() map[string][]string {
	return nil
}

func (a *Anime) TitleAlt() map[string]string {
	return nil
}
func (A *Anime) Aired() (time.Time, time.Time, bool) {
	return time.Now(), time.Now(), false
}

func (a *Anime) fetchEpisodes(url Url) func() error {
	return func() error {
		_ = url
		return nil
	}
}

func FetchAnime(url Url) (*Anime, error) {
	resp, err := http.Get(string(url))
	if err != nil {
		return nil, err
	}
	var anime Anime
	var eg errgroup.Group
	eg.Go(anime.fetchEpisodes(url))
	switch resp.StatusCode {
	case 404:
		return nil, nil
	case 429:
		fixLock()
		return FetchAnime(url)
	default:
		if resp.StatusCode != 200 {
			log.Panicf("unknown statuscode on anime fetch %v", resp.StatusCode)
		}
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
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
			v = strings.ReplaceAll(v, "\n", "")
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
