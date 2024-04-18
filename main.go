package main

import (
	"MalSql/mal"
	"fmt"
	"log"
)

func main() {
	anime, err := mal.FetchAnime(mal.UrlFromId(32867))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(anime)
}
