package main

import (
	"fmt"
)

func lol(x *[]int) {
	*x = append(*x, 2137)
}

func main() {
	x := []int{2, 1, 3, 7}
	lol(&x)
	fmt.Println(x)
	fmt.Println("hello, world")
}
