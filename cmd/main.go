package main

import (
	"flag"
	"fmt"
	"log"
	"malsql/mal"

	"github.com/t0k4r/x/chanx"
)

func main() {
	start := flag.Int("begin", 1, "begin index")
	end := flag.Int("end", 100000, "end index")
	multi := flag.Bool("fast", false, "gotta go fast")
	flag.Parse()
	_ = multi
	uc := chanx.NewUniq[int]()
	go func() {
		for i := *start; i < *end; i++ {
			fmt.Println(i, *start, *end)
			uc.Send(i)
		}
		uc.Wait()
		uc.Close()
	}()
	for i, ok := uc.Recv(); ok; {
		fmt.Println(i)
		anime, err := mal.FetchAnime(mal.UrlFromId(i))
		if err != nil {
			panic(err)
		}
		if anime != nil {
			log.Printf("%v: %v", i, anime.Title)
		}
	}

}
