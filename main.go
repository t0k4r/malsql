package main

import (
	"MalSql/scrap"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panic(err)
	}
	s := scrap.New()
	n := time.Now()
	s.Run(21, 22)
	fmt.Println(time.Since(n))
	// n := time.Now()
	// a, err := anime.LoadAnime(21)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// fmt.Println(time.Since(n), a.Title)
}
