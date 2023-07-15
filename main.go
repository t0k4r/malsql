package main

import (
	"MalSql/scrap/anime"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panic(err)
	}
	a, err := anime.LoadAnime(1)
	if err != nil {
		log.Panic(err)
	}
	println(a.Title)
	// s := scrap.New()
	// s.Run(1, 69000)
}
