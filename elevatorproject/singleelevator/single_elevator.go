package singleelevator 

import (
    "elevatorproject/singleelevator/elevator_timers"
    "elevatorproject/singleelevator/elevio"
)

// Har bare lagt til alt som trengs for spam

func RunSingleElevator(ch_single_mode chan bool, ch_cab_lamps chan[]bool, ch_hra chan [][2]bool, ch_init_cab_requests chan []bool, ch_completed_hall_requests chan elevio.ButtonEvent, ch_new_hall_requests chan elevio.ButtonEvent, ch_elevstate chan ElevState) {
    // Channels    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)   
	ch_door := make(chan int)
    ch_obstruction := make(chan int)
    ch_mech := make(chan int)
    ch_spam := make(chan int)
    
    // Elevio
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)

    // Elevator timers
    go elevator_timers.CheckDoorTimer(ch_door)
    go elevator_timers.CheckMechanicalTimer(ch_mech)
    go elevator_timers.CheckObstructionTimer(ch_obstruction)
    go elevator_timers.CheckSpamTimer(ch_spam)

    // single elev mode ch lagt til
	
    Fsm_elevator(drv_buttons, drv_floors, ch_door, ch_mech, ch_obstruction, ch_hra, ch_single_mode, ch_init_cab_requests, ch_completed_hall_requests, ch_new_hall_requests, ch_elevstate, ch_cab_lamps)
}

