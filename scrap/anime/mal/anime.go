package mal

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Related struct {
	TypeOf string
	Url    string
}
type Info struct {
	Key   string
	Value string
}

type Anime struct {
	MalUrl      string
	ImgUrl      string
	Title       string
	Description string
	Related     []Related
	Information []Info
}

var ErrMal404 = errors.New("no anime with given url exists mal404")
var ErrMal429 = errors.New("mal is blocked mal429")

func LoadAnime[T string | int](url T) (*Anime, error) {
	var anime Anime
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
		return nil, ErrMal404
	case 429:
		return nil, ErrMal429
	case 200:
		doc, err := goquery.NewDocumentFromReader(res.Body)

		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(anime.MalUrl, "co_ty_tutaj_robisz") {
			doc.Find("div.breadcrumb").ChildrenFiltered("div.di-ib").Each(func(i int, s *goquery.Selection) {
				if i == 2 {
					anime.MalUrl = s.ChildrenFiltered("a").AttrOr("href", "")
				}
			})
		}
		anime.ImgUrl = doc.Find(`img[itemprop="image"]`).AttrOr("data-src", "")
		anime.Title = doc.Find(`.title-name`).Text()
		anime.Description = doc.Find(`p[itemprop="description"]`).Text()
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
				anime.Information = append(anime.Information, Info{
					Key:   key,
					Value: v,
				})
			}
		})
		doc.Find(`.js-alternative-titles>.spaceit_pad`).Each(func(i int, s *goquery.Selection) {
			key := strings.Replace(strings.ToLower(s.ChildrenFiltered(`span.dark_text`).Text()), ":", "", 1)
			value := strings.Trim(strings.Join(strings.Split(s.Text(), ":")[1:], ":"), "\n ")
			anime.Information = append(anime.Information, Info{
				Key:   key,
				Value: value,
			})

		})
		doc.Find(".anime_detail_related_anime>tbody>tr").Each(func(i int, s *goquery.Selection) {
			typeof := ""
			s.ChildrenFiltered("td").Each(func(i int, s *goquery.Selection) {
				if i == 0 {
					typeof = strings.ToLower(s.Text())
					typeof = strings.Split(typeof, ":")[0]
				} else {
					s.ChildrenFiltered("a").Each(func(i int, s *goquery.Selection) {
						anime.Related = append(anime.Related, Related{
							TypeOf: typeof,
							Url:    fmt.Sprintf("https://myanimelist.net%v", s.AttrOr("href", "")),
						})
					})
				}
			})
		})
		var filteredRelated []Related
		for _, i := range anime.Related {
			if i.TypeOf != "adaptation" {
				filteredRelated = append(filteredRelated, i)
			}
		}
		anime.Related = filteredRelated
		return &anime, nil
	default:
		return nil, fmt.Errorf("unknown http status code %v", res.Status)
	}
}

func (a *Anime) MagicNumber() int {
	return MagicNumber(a.MalUrl)
}

func MagicNumber(url string) int {
	i, _ := strconv.Atoi(strings.Split(url, "/")[4])
	return i
}
