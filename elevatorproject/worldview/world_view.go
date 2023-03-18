package worldview

import (
	"elevatorproject/config"
	"elevatorproject/network/peers"
	"elevatorproject/nodeview"
	"elevatorproject/singleelevator"
	"elevatorproject/singleelevator/elevio"
	"fmt"
	"time"
)

type PeersAlive []string

type MyWorldView struct {
	SystemElevState      map[string]singleelevator.ElevState
	SystemHallRequests   map[string][][2]nodeview.RequestState
	SystemNodesAvailable map[string]bool
}

// Function that returns false if nodeID (input) is not in peersALive
func (peersAlive PeersAlive) IsPeerAlive(nodeID string) bool {
	for _, peer := range peersAlive {
		if peer == nodeID {
			return true
		}
	}
	return false
}

func (systemAwareness *MyWorldView) InitSystemAwareness() {
	systemAwareness.SystemElevState = make(map[string]singleelevator.ElevState, config.NumElevators)
	systemAwareness.SystemHallRequests = make(map[string][][2]RequestState, config.NumElevators)
	systemAwareness.SystemNodesAvailable = make(map[string]bool, config.NumElevators)
	systemAwareness.SystemNodesAvailable[config.LocalID] = true
}

func printNodeView(node nodeview.MyNodeView) {
	fmt.Printf("ID: %s\n", node.ID)
	fmt.Printf("IsAvailable: %v\n", node.IsAvailable)
	fmt.Printf("ElevState: Behaviour=%s Floor=%d Direction=%s CabRequests=%v IsAvailable=%v\n",
		node.ElevState.Behaviour, node.ElevState.Floor, node.ElevState.Direction, node.ElevState.CabRequests, node.ElevState.IsAvailable)
	fmt.Printf("HallRequests:\n")
	for i, requests := range node.HallRequests {
		fmt.Printf("  Floor %d: Up=%v Down=%v\n", i+1, requests[0], requests[1])
	}
	fmt.Printf("CabRequests:\n")
	for id, requests := range node.CabRequests {
		fmt.Printf("  Cab %s: %v\n", id, requests)
	}
}

func WorldView(ch_sendNodeView chan<- NodeView,
	ch_receiveNodeView <-chan NodeView,
	ch_receivePeerUpdate <-chan peers.PeerUpdate,
	ch_setTransmitEnable chan<- bool,
	ch_newHallRequest <-chan elevio.ButtonEvent,
	ch_compledtedHallRequest <-chan elevio.ButtonEvent,
	ch_elevState <-chan singleelevator.ElevState,
	ch_hallRequests chan<- [][2]bool,
	ch_initCabRequests chan<- []bool,
	ch_hraInput chan<- MyWorldView) {

	var myNodeView NodeView
	var systemAwareness MyWorldView
	var peersAlive PeersAlive
	var singleElevatorMode = true

	myNodeView.InitNodeView()
	systemAwareness.InitSystemAwareness()
// Veldig viktig å huske på at i singelelevmode så må hallcalls bli sendt til hraInput
	for {
		select {
		case peerUpdate := <-ch_receivePeerUpdate:
			// Go to single elevator mode if at least one node left
			peersAlive = peerUpdate.Peers
			peersLost := peerUpdate.Lost
			// Heller kjøre en metode som sletter her
			if len(peersAlive) <= 1 && !singleElevatorMode {
				singleElevatorMode = true
				// We must stop broadcasting our node awareness. Disable the channel
				//ch_setTransmitEnable <- false

			} else if singleElevatorMode && len(peersAlive) > 1 {
				singleElevatorMode = false
				// We must set all no order to unknown
				myNodeView.ChangeNoOrderAndConfirmedToUnknown()
				// Update system awareness hall requests
				systemAwareness.SystemHallRequests[config.LocalID] = myNodeView.HallRequests
				fmt.Println("Kobler oss tilbake på nettverket igjen")
				//ch_setTransmitEnable <- true
				// else single elevator mode false
			} //else if len(peersAlive) > 1 && singleElevatorMode {
			//singleElevatorMode = false
			//}

			for _, lostPeer := range peersLost {
				// If this node can be found in lostPeer, we should delete it from the systemAwareness
				if lostPeer != config.LocalID {
					delete(systemAwareness.SystemNodesAvailable, lostPeer)
					delete(systemAwareness.SystemElevState, lostPeer)
					delete(systemAwareness.SystemHallRequests, lostPeer)
				}
			}

			// print peer update
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)

			// Here I can add if I am in an init state, I should send cab call of LocalID on channel init_cab_requests
			// This will be done in the init state of the elevator
		case NodeView := <-ch_receiveNodeView:
			fmt.Println("Received from ", NodeView.ID)
			fmt.Println(NodeView.HallRequests)

			nodeID := NodeView.ID
			// Break out of case if IsPeerAlive returns false
			if !peersAlive.IsPeerAlive(nodeID) || nodeID == config.LocalID {
				break
			}
			systemAwareness.SystemNodesAvailable[nodeID] = NodeView.IsAvailable
			systemAwareness.SystemElevState[nodeID] = NodeView.ElevState
			systemAwareness.SystemHallRequests[nodeID] = NodeView.HallRequests
			myNodeView.CabRequests[nodeID] = NodeView.ElevState.CabRequests

			// Gjennomfører sammenlikningen
			hallRequests := updateMyHallRequestView(systemAwareness.SystemHallRequests)

			// Her skal både myNodeView og min id på systemAwareness oppdateres
			systemAwareness.SystemHallRequests[config.LocalID] = hallRequests
			myNodeView.HallRequests = hallRequests

			// Send systemAwareness til hra modulen
			ch_hraInput <- systemAwareness

			// Etter dette må vi sende til panelmodulen for å kunne sette lys.
			ch_hallRequests <- convertHallRequestStateToBool(hallRequests, singleElevatorMode)

			// Debug print

		case elevState := <-ch_elevState:
			// Her skal vi oppdatere vår egen elevstate
			//fmt.Println("Inside getting elevState channel case")
			myNodeView.ElevState = elevState
			systemAwareness.SystemElevState[config.LocalID] = elevState

			fmt.Println(systemAwareness.SystemElevState[config.LocalID])

		case <-time.After(50 * time.Millisecond):
			// Her skal vi sende vår egen NodeView på nettverket
			ch_sendNodeView <- myNodeView

			//printNodeView(myNodeView)
		}
	}
}
