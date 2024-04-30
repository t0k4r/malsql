package main

import (
	"MalSql/get"
	"MalSql/mal"
	"fmt"
	"net/http"
)

func main() {
	http.DefaultClient.Transport = &AddHeaderTransport{
		T: http.DefaultTransport,
	}
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
