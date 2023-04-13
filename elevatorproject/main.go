package main

import (
	"elevatorproject/config"
	"elevatorproject/network/bcast"
	"elevatorproject/network/peers"
	"elevatorproject/nodeview"
	"elevatorproject/requestassigner"
	"elevatorproject/singleelevator"
	"elevatorproject/singleelevator/elevio"
	"elevatorproject/worldview"
	//"flag"
	"fmt"
)

func main() {
	// var localID string
	// var elevPort string

	// flag.StringVar(&localID, "id", "", "id for this peer")
	// flag.StringVar(&elevPort, "port", "", "port of this peer")
	// flag.Parse()

	// localIP := "localhost:" + elevPort

	// localIP := "localhost:15657"
	// localID := "elev1"

	// localIP := "localhost:17000"
	// localID := "elev2"

	localIP := "localhost:18000"
	localID := "elev3"

	fmt.Printf("Starter programmet")
	elevio.Init(localIP, config.NumFloors)

	//singleelevator
	ch_completedHallRequests := make(chan elevio.ButtonEvent)
	ch_newHallRequests := make(chan elevio.ButtonEvent)
	ch_elevState := make(chan singleelevator.ElevState)

	//worldview
	ch_hraInput := make(chan worldview.MyWorldView)
	ch_hallLamps := make(chan [][2]bool)
	ch_cabLamps := make(chan []bool)
	ch_remoteRequestView := make(chan nodeview.RemoteRequestView)
	ch_singleElevMode := make(chan bool)

	// hra
	ch_hallRequest := make(chan [][2]bool)
	ch_cabRequests := make(chan []bool)

	// network in
	ch_receive := make(chan nodeview.MyNodeView)
	ch_peerUpdate := make(chan peers.PeerUpdate)
	go peers.Receiver(13200, ch_peerUpdate)
	go bcast.Receiver(12100, ch_receive)

	// network out
	ch_transmit := make(chan nodeview.MyNodeView)
	ch_peerTransmitEnable := make(chan bool)
	go peers.Transmitter(13200, localID, ch_peerTransmitEnable)
	go bcast.Transmitter(12100, ch_transmit)

	// go routines
	go worldview.WorldView(ch_receive, ch_peerUpdate, ch_remoteRequestView, ch_hraInput, ch_singleElevMode, localID)
	go nodeview.NodeView(ch_transmit, ch_newHallRequests, ch_completedHallRequests, ch_elevState, ch_hallLamps, ch_cabLamps, ch_remoteRequestView, localID)

	go requestassigner.AssignRequests(ch_hraInput, ch_hallRequest, ch_cabRequests, localID)
	go singleelevator.LampStateMachine(ch_hallLamps, ch_cabLamps)

	fmt.Println("Starter opp singleelevator")
	go singleelevator.RunSingleElevator(ch_hallRequest, ch_cabRequests, ch_completedHallRequests, ch_newHallRequests, ch_elevState, ch_singleElevMode)

	for {
		select {
		default:
		}
	}
}
