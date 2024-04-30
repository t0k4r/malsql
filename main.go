package main

import (
	"MalSql/get"
	"MalSql/mal"
	"fmt"
	"net/http"
)

type AddHeaderTransport struct {
	T http.RoundTripper
}

func (adt *AddHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0")
	return adt.T.RoundTrip(req)
}

func init() {
	http.DefaultClient.Transport = &AddHeaderTransport{
		T: http.DefaultTransport,
	}
}

func main() {

	err := get.
		NewUniqChan().
		Generate(1, 16).
		ParalellEach(func(uc *get.UniqChan, i int) error {
			anime, err := mal.FetchAnime(mal.UrlFromId(i))
			if err != nil {
				return err
			}
			if anime == nil {
				return nil
			}
			var ids []int
			for _, v := range anime.Related {
				for _, id := range v {
					ids = append(ids, id.Id())
				}
			}
			uc.Append(ids...)
			fmt.Println(anime.Title)
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}
}
