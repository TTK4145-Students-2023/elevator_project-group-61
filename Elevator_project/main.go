package main

import (
	"ElevatorProject/door_timer"
	"ElevatorProject/elevio"
	"time"
)

func main(){

    numFloors := 4

    elevio.Init("localhost:15657", numFloors) 
    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    // drv_stop    := make(chan bool)    

	ch_time := make(chan int)
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    // go elevio.PollStopButton(drv_stop)

    // New routine for checking delegated orders
    ch_delegated_order := make(chan elevio.ButtonEvent)
    

	go func() {
		for {
			door_timer.CheckTimer(ch_time)
			time.Sleep(10 * time.Millisecond)
		}
	}()
	
    Fsm_elevator(drv_buttons, drv_floors, ch_time)
}