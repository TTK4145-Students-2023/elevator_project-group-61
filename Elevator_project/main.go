package main

import (
	// "ElevatorProject/elevio_driver"
	"ElevatorProject/door_timer"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hey im just printing something first.")
	door_timer.StartTimer()
	for i := 0; i < 10; i++ {
		if door_timer.CheckTimer() {
			fmt.Println("Timer FINISHED!!")
			return
		}
		fmt.Println("Not finished :(")
		time.Sleep(400 * time.Millisecond)
	}
}
