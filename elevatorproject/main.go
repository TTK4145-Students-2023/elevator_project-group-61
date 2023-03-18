package main

import (
	"elevatorproject/config"
	"elevatorproject/hallrequestassigner"
	"elevatorproject/network"
	"elevatorproject/network/peers"
	"elevatorproject/singleelevator"
	"elevatorproject/singleelevator/elevio"
	"elevatorproject/nodeview"
	"elevatorproject/worldview"
	"fmt"
)

func main() {

	fmt.Printf("Starter programmet")
	elevio.Init("localhost:15657", config.NumFloors)

	//singleelevator
	ch_completedHallRequests := make(chan elevio.ButtonEvent)
	ch_newHallRequests := make(chan elevio.ButtonEvent)
	ch_elevState := make(chan singleelevator.ElevState)
	ch_cabLamps := make(chan []bool)

	//worldview
	ch_initCabRequests := make(chan []bool)
	ch_hraInput := make(chan worldview.MyWorldView)
	ch_hallRequests := make(chan [][2]bool)
	ch_remoteRequestView := make(chan nodeview.RemoteRequestView)

	// receive_network
	ch_receivePeerUpdate := make(chan peers.PeerUpdate)
	ch_receiveNodeView := make(chan nodeview.MyNodeView)
	
	// transmit_network
	ch_sendMyNodeView := make(chan nodeview.MyNodeView)
	ch_setTransmitEnable := make(chan bool)


	// hra
	ch_hraOutput := make(chan [][2]bool)

	// go routines
	go network.TransmitNetwork(ch_sendMyNodeView, ch_setTransmitEnable)
	go network.ReceiveNetwork(ch_receiveNodeView, ch_receivePeerUpdate)

	go worldview.WorldView(ch_receiveNodeView, ch_receivePeerUpdate, ch_setTransmitEnable, ch_initCabRequests, ch_remoteRequestView, ch_hraInput)
	go nodeview.NodeView(ch_sendMyNodeView, ch_newHallRequests, ch_completedHallRequests, ch_elevState, ch_hallRequests, ch_remoteRequestView)


	go hallrequestassigner.AssignHallRequests(ch_hraInput, ch_hraOutput)
	go singleelevator.LampStateMachine(ch_hallRequests, ch_cabLamps)
	
	fmt.Println("Starter opp singleelevator")
	singleelevator.RunSingleElevator(ch_cabLamps, ch_hraOutput, ch_initCabRequests, ch_completedHallRequests, ch_newHallRequests, ch_elevState)

}
