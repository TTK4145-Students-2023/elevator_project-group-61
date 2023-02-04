package main

import (

)

var somelength int = 10

var someslice []int

func main() {
	someslice = make([]int, somelength)
	for i := 0; i < somelength; i++ {
		someslice[i] = i
	}
}