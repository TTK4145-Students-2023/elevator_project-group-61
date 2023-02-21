package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, playground")
	var somevar int = 2

	switch (somevar) {
	case 1:
	case 2:
		fmt.Println("2")
	case 3:
		fmt.Println("3")
	default:
		fmt.Println("default")
	}

}