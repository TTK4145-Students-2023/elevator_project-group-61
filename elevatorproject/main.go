package main

import (
	"elevatorproject/config"
	"elevatorproject/network/bcast"
	"elevatorproject/network/peers"
	"elevatorproject/peerview"
	"elevatorproject/requestassigner"
	"elevatorproject/singleelevator"
	"elevatorproject/singleelevator/elevio"
	"elevatorproject/worldview"
	"elevatorproject/lamps"
	// "flag"
	"fmt"
)

func main() {
	localID := "elev2"
	elevPort := "15657"
	// var localID string
	// var elevPort string

	// flag.StringVar(&localID, "id", "", "id of this peer")
	// flag.StringVar(&elevPort, "port", "", "port of this peer")
	// flag.Parse()	

	localIP := "localhost:" + elevPort

	fmt.Printf("Starting program")
	elevio.Init(localIP, config.NumFloors)

	//singleelevator
	ch_completedRequest := make(chan elevio.ButtonEvent)
	ch_newRequest := make(chan elevio.ButtonEvent)
	ch_elevState := make(chan singleelevator.ElevState)

	//worldview
	ch_myWorldView := make(chan worldview.MyWorldView)
	ch_remoteRequestView := make(chan peerview.RemoteRequestViews)
	ch_singleElevMode := make(chan bool)

	// peerview
	ch_hallLamps := make(chan [config.NumFloors][2]bool)
	ch_cabLamps := make(chan [config.NumFloors]bool)

	// requestassigner
	ch_hallRequests := make(chan [config.NumFloors][2]bool)
	ch_cabRequests := make(chan [config.NumFloors]bool)

	// network in
	ch_receive := make(chan peerview.MyPeerView)
	ch_peerUpdate := make(chan peers.PeerUpdate)
	go peers.Receiver(13200, ch_peerUpdate)
	go bcast.Receiver(12100, ch_receive)

	// network out
	ch_transmit := make(chan peerview.MyPeerView)
	ch_peerTransmitEnable := make(chan bool)
	go peers.Transmitter(13200, localID, ch_peerTransmitEnable)
	go bcast.Transmitter(12100, ch_transmit)

	// go routines
	go worldview.WorldView(ch_receive, ch_peerUpdate, ch_remoteRequestView, ch_myWorldView, ch_singleElevMode, localID)
	go peerview.PeerView(ch_transmit, ch_newRequest, ch_completedRequest, ch_elevState, ch_hallLamps, ch_cabLamps, ch_remoteRequestView, localID)

	go requestassigner.AssignRequests(ch_myWorldView, ch_hallRequests, ch_cabRequests, localID)
	go lamps.LampStateMachine(ch_hallLamps, ch_cabLamps)

	fmt.Println("Starter opp singleelevator")
	go singleelevator.RunSingleElevator(ch_hallRequests, ch_cabRequests, ch_singleElevMode, ch_completedRequest, ch_newRequest, ch_elevState)

	for {
		select {
		default:
		}
	}
}
