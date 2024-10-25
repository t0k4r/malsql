package main

import (
	"flag"
	"fmt"
	"malsql"
	"malsql/mal"
)

func main() {
	// 35073
	anime, _ := mal.FetchAnime(mal.IdToUrl(35073))
	fmt.Printf("%+v\n", anime)
	return
	f := malsql.Flags{}
	flag.IntVar(&f.Start, "begin", 1, "begin index")
	flag.IntVar(&f.End, "end", 100000, "end index")
	flag.BoolVar(&f.Fast, "fast", false, "gotta go fast")
	flag.StringVar(&f.File, "file", "anime.sql", "sql file to save ")
	flag.Parse()
	malsql.New(f)
}
