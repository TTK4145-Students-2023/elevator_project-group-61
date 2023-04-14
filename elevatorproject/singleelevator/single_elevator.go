package singleelevator 

import (
    "elevatorproject/config"
    "elevatorproject/singleelevator/elevator_timers"
    "elevatorproject/singleelevator/elevio"
)

// Har bare lagt til alt som trengs for spam

func RunSingleElevator(ch_hra chan [config.NumFloors][2]bool, ch_cab_requests chan [config.NumFloors]bool, ch_completed_requests chan elevio.ButtonEvent, ch_new_requests chan elevio.ButtonEvent, ch_elevstate chan ElevState, ch_singleElevMode chan bool) {
    // Channels    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)   
	ch_door := make(chan int)
    ch_error := make(chan int)
    
    // Elevio
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)

    // Elevator timers
    go elevator_timers.CheckDoorTimer(ch_door)
    go elevator_timers.CheckErrorTimer(ch_error)

    // single elev mode ch lagt til
    Fsm_elevator(drv_buttons, drv_floors, ch_door, ch_error, ch_hra, ch_cab_requests, ch_completed_requests, ch_new_requests, ch_elevstate, ch_singleElevMode)
}

