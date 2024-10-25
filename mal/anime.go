package mal

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

func IdToUrl(id int) string {
	return fmt.Sprintf("https://myanimelist.net/anime/%v/Bungou_Stray_Dogs_2nd_Season", id)
}

func UrlToId(u string) (id int) {
	id, _ = strconv.Atoi(strings.Split(u, "/")[4])
	return
}

type Anime struct {
	Url               string
	Image             string
	Title             string
	Synopsis          string
	Background        string
	AlternativeTitles map[string]string
	Information       map[string][]string
	RelatedEntries    map[string][]string
}

// anime can be nil
func FetchAnime(url string) (*Anime, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	switch res.StatusCode {
	case 200:
	case 404:
		return nil, nil
	case 429, 405:
		waitLock()
		return FetchAnime(url)
	default:
		return nil, fmt.Errorf("unknown statuscode on anime fetch %v", res.StatusCode)
	}
	root, err := html.Parse(res.Body)
	if err != nil {
		return nil, err
	}
	doc := goquery.NewDocumentFromNode(root)

	var anime Anime
	anime.Url, _ = doc.Find("#horiznav_nav>ul>li>a").Attr("href")
	anime.Image, _ = doc.Find(".leftside>div>a>img").Attr("data-src")
	anime.Title = doc.Find("h1.title-name").Text()
	anime.Synopsis = doc.Find(`p[itemprop="description"]`).Text()
	anime.Background = readBackground(root)
	// anime.AlternativeTitles = readAlternativeTitles(root)
	// anime.Information = readInformation(root)
	// anime.RelatedEntries = readRelatedEntries(root)
	return &anime, nil
}

func readBackground(n *html.Node) string {
	sel, _ := cascadia.Parse("#background")
	node := cascadia.Query(n, sel).Parent.NextSibling
	var text string
	for node != nil {
		switch node.Type {
		case html.TextNode:
			text += node.Data
		default:
			if node.FirstChild != nil && node.FirstChild.Data != "div" {
				text += node.FirstChild.Data
			}
		}
		node = node.NextSibling
	}
	return text
}
func readAlternativeTitles(n *html.Node) (alt map[string]string) { return }
func readInformation(n *html.Node) (inf map[string][]string)     { return }
func readRelatedEntries(n *html.Node) (ent map[string][]string)  { return }
