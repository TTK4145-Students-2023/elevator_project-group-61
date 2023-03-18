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
	"elevatorproject/network/bcast"
	"elevatorproject/network/peers"
	"fmt"
	"time"
)

func main() {

	fmt.Printf("Starter programmet")
	elevio.Init(config.LocalIP, config.NumFloors)

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


	// hra
	ch_hraOutput := make(chan [][2]bool)

	// network in
	ch_receive := make(chan nodeview.MyNodeView)
	ch_peerUpdate := make(chan peers.PeerUpdate)
	go peers.Receiver(13200, ch_peerUpdate)
	go bcast.Receiver(12100, ch_receive)

	// network out
	ch_transmit := make(chan nodeview.MyNodeView)
	ch_peerTransmitEnable := make(chan bool)
	go peers.Transmitter(13200, config.LocalID, ch_peerTransmitEnable)
	go bcast.Transmitter(12100, ch_transmit)


	// go routines

	go worldview.WorldView(ch_receive, ch_peerUpdate, ch_peerTransmitEnable, ch_initCabRequests, ch_remoteRequestView, ch_hraInput)
	go nodeview.NodeView(ch_transmit, ch_newHallRequests, ch_completedHallRequests, ch_elevState, ch_hallRequests, ch_remoteRequestView)


	go hallrequestassigner.AssignHallRequests(ch_hraInput, ch_hraOutput)
	go singleelevator.LampStateMachine(ch_hallRequests, ch_cabLamps)

	fmt.Println("Starter opp singleelevator")
	go singleelevator.RunSingleElevator(ch_cabLamps, ch_hraOutput, ch_initCabRequests, ch_completedHallRequests, ch_newHallRequests, ch_elevState)

	for {
		select {
		default:
		}
	}

}
