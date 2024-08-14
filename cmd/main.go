package main

import (
	"flag"
	"malsql"
)

func main() {
	start := flag.Int("begin", 1, "begin index")
	end := flag.Int("end", 100000, "end index")
	fast := flag.Bool("fast", false, "gotta go fast")
	flag.Parse()

	malsql.New(*start, *end, *fast)
}
