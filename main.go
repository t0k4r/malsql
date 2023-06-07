package main

import (
	"MalSql/scrap"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panic(err)
	}
	s := scrap.New()
	s.Run(1, 69000)
}
