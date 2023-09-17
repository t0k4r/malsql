package main

import (
	"MalSql/scrap"
	"flag"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/exp/slog"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Info(err.Error())
	}
	var opts scrap.Options
	flag.IntVar(&opts.Start, "start", 1, "start index")
	flag.IntVar(&opts.End, "end", 75000, "end index")
	flag.BoolVar(&opts.Skip, "skip", false, "skip done animes")
	flag.BoolVar(&opts.File, "file", false, "dump to file not db")
	flag.BoolVar(&opts.Quick, "quick", false, "faster but very inefficient")
	flag.Parse()
	scrap.New(opts).Run()
}
