package main

import (
	"fmt"
	"time"
)


func test_function(some_ch chan<-int) {
	some_ch <- 5
}

func main() {
	some_channel := make(chan int)

	go test_function(some_channel)
	time.Sleep(time.Second * 2)
	a := some_channel
	fmt.Println(a)
}