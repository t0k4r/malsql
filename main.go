package main

import (
	"MalSql/scrap"
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("postgres", os.Getenv("PG_CONN"))
	if err != nil {
		log.Panic(err)
	}
	s := scrap.New(db)
	s.Run(1, 10)
}
