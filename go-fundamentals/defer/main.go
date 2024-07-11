package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, playground")

	defer fmt.Println("world")

	fmt.Println("Hello, playground 2")
}
