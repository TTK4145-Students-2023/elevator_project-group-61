package main

import (
	"elevatorproject/config"
	"elevatorproject/hallrequestassigner"
	"elevatorproject/network"
	"elevatorproject/network/peers"
	"elevatorproject/singleelevator"
	"elevatorproject/systemview"
	"elevatorproject/singleelevator/elevio"
)


func main() {
	numFloors := 4

    elevio.Init("localhost:15657", numFloors) 

    ch_completedHallRequests := make(chan elevio.ButtonEvent)
    ch_newHallRequests := make(chan elevio.ButtonEvent)
    ch_elevState := make(chan singleelevator.ElevState)
	// ch_cabRequests := make(chan []bool)
	ch_cabLamps := make(chan []bool)
	
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

	go network.Network(ch_sendNodeAwareness, ch_receiveNodeAwareness, ch_receivePeerUpdate, ch_setTransmitEnable)

	go systemview.SystemView(ch_sendNodeAwareness, ch_receiveNodeAwareness, ch_receivePeerUpdate, ch_setTransmitEnable, ch_newHallRequests, 
		ch_completedHallRequests, ch_elevState, ch_hallRequests, ch_initCabRequests, ch_hraInput)

	go hallrequestassigner.AssignHallRequests(ch_hraInput, ch_hraOutput)
	go singleelevator.LampStateMachine(ch_hallRequests, ch_cabLamps)
	singleelevator.RunSingleElevator(ch_cabLamps, ch_hraOutput, ch_initCabRequests, ch_completedHallRequests, ch_newHallRequests, ch_elevState)
}

