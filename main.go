package main

import (
	"MalSql/mal"
	"MalSql/uniqchan"
	"fmt"
	"log"
	"sync"
)

func main() {
	// pull, push := get.NewUniqChan().Generate(0, 10).Run()
	// for i := range pull {
	// 	anime, err := mal.FetchAnime(mal.UrlFromId(i))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if anime == nil {
	// 		continue
	// 	}
	// 	var ids []int
	// 	for _, v := range anime.Related {
	// 		for _, id := range v {
	// 			ids = append(ids, id.Id())
	// 		}
	// 	}
	// 	fmt.Println(anime.Title)
	// 	push <- ids
	// 	// return nil
	// }

	// err := get.
	// 	NewUniqChan().
	// 	Generate(1, 10).
	// 	Each(func(uc *get.UniqChan, i int) error {
	// 		anime, err := mal.FetchAnime(mal.UrlFromId(i))
	// 		if err != nil {
	// 			return err
	// 		}
	// 		if anime == nil {
	// 			return nil
	// 		}
	// 		var ids []int
	// 		for _, v := range anime.Related {
	// 			for _, id := range v {
	// 				ids = append(ids, id.Id())
	// 			}
	// 		}
	// 		uc.Append(ids...)
	// 		fmt.Println(anime.Title)
	// 		return nil
	// 	})
	// if err != nil {
	// 	fmt.Println(err)
	// }
	uq := uniqchan.New[int]()
	pushWg := &sync.WaitGroup{}
	uniqchan.Generate(uq, pushWg, 1, 256)
	for i := range uq.Chan {
		anime, err := mal.FetchAnime(mal.UrlFromId(i))
		if err != nil {
			log.Fatal(err)
		}
		if anime == nil {
			continue
		}
		var ids []int
		for _, v := range anime.Related {
			for _, id := range v {
				ids = append(ids, id.Id())
			}
		}
		uq.GoPush(pushWg, ids...)
		fmt.Println(anime.Title)

	}
}
