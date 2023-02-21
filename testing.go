package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, playground")
	var somevar int = 1

	switch (somevar) {
	case 1:
		somevar, bool := 3, true
		fmt.Println(somevar, bool)
	case 2:
		fmt.Println("2")
	case 3:
		fmt.Println("3")
	default:
		fmt.Println("default")
	}

}