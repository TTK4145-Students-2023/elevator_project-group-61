package singleelevator

import (
	"elevatorproject/config"
	"elevatorproject/singleelevator/elevatortimers"
	"elevatorproject/singleelevator/elevio"
)

func RunSingleElevator(
	ch_hallRequests     <-chan [config.NumFloors][2]bool,
	ch_cabRequests      <-chan [config.NumFloors]bool,
	ch_singleElevMode   <-chan bool,
	ch_completedRequest chan<- elevio.ButtonEvent,
	ch_newRequest       chan<- elevio.ButtonEvent,
	ch_elevState        chan<- ElevState,
) {
	// Channels
	ch_buttons := make(chan elevio.ButtonEvent)
	ch_floors := make(chan int)
	ch_door := make(chan int)
	ch_error := make(chan int)

	// Elevio
	go elevio.PollButtons(ch_buttons)
	go elevio.PollFloorSensor(ch_floors)

	// Elevator timers
	go elevatortimers.CheckDoorTimer(ch_door)
	go elevatortimers.CheckErrorTimer(ch_error)

	// Finite state machine
	fsmElevator(
        ch_buttons, 
        ch_floors, 
        ch_door, 
        ch_error, 
        ch_hallRequests, 
        ch_cabRequests, 
        ch_singleElevMode, 
        ch_completedRequest, 
        ch_newRequest, 
        ch_elevState,
    )
}
