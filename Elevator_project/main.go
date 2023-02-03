package main

import (
	// "ElevatorProject/elevio_driver"
	"ElevatorProject/door_timer"
	"fmt"
)

func main() {
	some_door_ch := make(chan int)

	fmt.Println("Starting timer!")

	go door_timer.DoorTimerChannel(some_door_ch)
	go func() {
		for {
			select {
			case <-some_door_ch:
				fmt.Println("Channel received! Timer ended!")
				return
			default:
				fmt.Println("Wtf")
			}
		}
	}()
}
