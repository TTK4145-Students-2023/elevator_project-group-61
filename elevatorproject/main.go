package main

import (
	"elevatorproject/config"
	"elevatorproject/hallrequestassigner"
	"elevatorproject/network"
	"elevatorproject/network/peers"
	"elevatorproject/singleelevator"
	"elevatorproject/systemview"
	"elevatorproject/singleelevator/elevio"
	"elevatorproject/singleelevator/doortimer"
)


func main() {
	numFloors := 4

    elevio.Init("10.100.23.27:15657", numFloors) 
	
	// singleelevator
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)   
	ch_door := make(chan int)

    ch_completedHallRequests := make(chan elevio.ButtonEvent)
    ch_newHallRequests := make(chan elevio.ButtonEvent)
    ch_elevState := make(chan singleelevator.ElevState)
	ch_cabRequests := make(chan []bool)
	
	// network
	ch_peerTransmitEnable := make(chan bool)
	ch_peerUpdate := make(chan peers.PeerUpdate)

	//systemview
	ch_initCabRequests := make(chan []bool)
	ch_hraInput := make(chan systemview.SystemAwareness)
	ch_hallRequests := make(chan [][2]bool)

	ch_setTransmitEnable := make(chan bool)
	ch_receivePeerUpdate := make(chan peers.PeerUpdate)
	ch_receiveNodeAwareness := make(chan systemview.NodeAwareness)
	ch_sendNodeAwareness := make(chan systemview.NodeAwareness)
	
	// hra
	ch_hraOutput := make(chan [][2]bool)
	
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go doortimer.CheckTimer(ch_door)


    
	go singleelevator.Fsm_elevator(drv_buttons, drv_floors, ch_door, ch_hraOutput, ch_initCabRequests, ch_completedHallRequests, ch_newHallRequests, ch_elevState)
	go network.Network(ch_sendNodeAwareness, ch_receiveNodeAwareness, ch_receivePeerUpdate, ch_setTransmitEnable)
	go peers.Receiver(12222, ch_peerUpdate)
	go peers.Transmitter(12223, config.LocalID, ch_peerTransmitEnable)

	go systemview.SystemView(ch_sendNodeAwareness, ch_receiveNodeAwareness, ch_receivePeerUpdate, ch_setTransmitEnable, ch_newHallRequests, 
		ch_completedHallRequests, ch_elevState, ch_hallRequests, ch_initCabRequests, ch_hraInput)

	go hallrequestassigner.AssignHallRequests(ch_hraInput, ch_hraOutput)
	go singleelevator.LampStateMachine(ch_hallRequests, ch_cabRequests)
}

