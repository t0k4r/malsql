package main

import (
	"MalSql/scrap"
	"flag"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}
	start := flag.Int("start", 1, "start index")
	end := flag.Int("end", 75000, "end index")
	skip := flag.Bool("skip", false, "skip done animes")
	file := flag.Bool("file", false, "dump to file not db")
	flag.Parse()
	s := scrap.New()
	s.Run(scrap.Options{Start: *start, End: *end, Skip: *skip, File: *file})

}
