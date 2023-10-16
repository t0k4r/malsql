package main

import (
	"MalSql/scrap"
	"flag"
	"os"

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
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.IntVar(&opts.Start, "start", 1, "start index")
	flag.IntVar(&opts.End, "end", 75000, "end index")
	flag.BoolVar(&opts.Quick, "quick", false, "faster but very inefficient")
	flag.BoolVar(&opts.File, "file", false, "dump to sql file not database")
	flag.BoolVar(&opts.Skip, "skip", false, "skip done animes (not availible if -file)")
	flag.BoolVar(&opts.Update, "update", false, "on conflict update/replace (not availible if -file)")
	flag.StringVar(&opts.Dialect, "dialect", "sqlite", "postgress or sqlite")
	flag.StringVar(&opts.Conn, "conn", "./MalSql.sqlite", "database connection string")
	flag.BoolVar(&opts.Env, "env", false, "read database connection from env (MALSQL_DB)")
	flag.Parse()

	scrap.New(opts).Run()
}
