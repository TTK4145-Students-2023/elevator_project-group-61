package systemview

import (
	"elevatorproject/config"
	"elevatorproject/network/peers"
	"elevatorproject/singleelevator"
	"elevatorproject/singleelevator/elevio"
	"fmt"
	"time"
)

type RequestState int

// All peers alive in a list of same type peers

const (
	RS_Unknown   RequestState = -1
	RS_NoOrder   RequestState = 0
	RS_Pending   RequestState = 1
	RS_Confirmed RequestState = 2
	RS_Completed RequestState = 3
)

type PeersAlive []string

type NodeAwareness struct {
	ID           string
	IsAvailable  bool
	ElevState    singleelevator.ElevState
	HallRequests [][2]RequestState // n number of floors
	CabRequests  map[string][]bool
}

type SystemAwareness struct {
	SystemElevState      map[string]singleelevator.ElevState
	SystemHallRequests   map[string][][2]RequestState
	SystemNodesAvailable map[string]bool
}

func updateMyHallRequestView(systemHallRequests map[string][][2]RequestState) [][2]RequestState {
	myView := systemHallRequests[config.LocalID]
	delete(systemHallRequests, config.LocalID)

	for row := 0; row < len(myView); row++ {
		for col := 0; col < len(myView[row]); col++ {
			hall_order := myView[row][col]

			switch hall_order {
			case RS_Unknown:
				max_count := int(hall_order)
				for _, nodeView := range systemHallRequests {
					if (int(nodeView[row][col]) > max_count) && nodeView[row][col] != RS_Completed {
						max_count = int(nodeView[row][col])
					}
				}
				myView[row][col] = RequestState(max_count)
			case RS_NoOrder:
				// Go to RS_Pending if any other node has RS_Pending
				for _, nodeView := range systemHallRequests {
					if nodeView[row][col] == RS_Pending {
						myView[row][col] = RS_Pending
						break
					}
				}
			case RS_Pending:
				pendingCount := 0
				for _, nodeView := range systemHallRequests {
					if nodeView[row][col] == RS_Confirmed {
						myView[row][col] = RS_Confirmed
						break
					} else if nodeView[row][col] == RS_Pending {
						pendingCount++
					}
				}
				if pendingCount == len(systemHallRequests) {
					myView[row][col] = RS_Confirmed
				}
			case RS_Confirmed:
				for _, nodeView := range systemHallRequests {
					// TODO: Check if or nodeView[row][col] == RS_Confirmed is needed
					if nodeView[row][col] == RS_Completed {
						myView[row][col] = RS_NoOrder
						break
					}
				}
			case RS_Completed:
				// Go to RS_NoOrder if all other nodes have anything else than RS_Confirmed
				noOrderCount := 0
				for _, nodeView := range systemHallRequests {
					if nodeView[row][col] != RS_Confirmed {
						noOrderCount++
					}
				}
				if noOrderCount == len(systemHallRequests) {
					myView[row][col] = RS_NoOrder
				}
			}
		}
	}
	return myView
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

// Function that changes all NoOrder to Unknown as a method of the NodeAwareness struct
func (nodeAwareness *NodeAwareness) ChangeNoOrderAndConfirmedToUnknown() {
	for row := 0; row < len(nodeAwareness.HallRequests); row++ {
		for col := 0; col < len(nodeAwareness.HallRequests[row]); col++ {
			if nodeAwareness.HallRequests[row][col] == RS_NoOrder || nodeAwareness.HallRequests[row][col] == RS_Confirmed {
				nodeAwareness.HallRequests[row][col] = RS_Unknown
			}
		}
	}
}

func (nodeAwareness *NodeAwareness) InitNodeAwareness() {
	nodeAwareness.ID = config.LocalID
	nodeAwareness.HallRequests = make([][2]RequestState, config.NumFloors)
	nodeAwareness.CabRequests = make(map[string][]bool)
}

func (systemAwareness *SystemAwareness) InitSystemAwareness() {
	systemAwareness.SystemElevState = make(map[string]singleelevator.ElevState, config.NumElevators)
	systemAwareness.SystemHallRequests = make(map[string][][2]RequestState, config.NumElevators)
}

// function that takes a [][2]RequestState as input and return [][2]bool
func convertHallRequestStateToBool(hallRequests [][2]RequestState, singleElevatorMode bool) [][2]bool {
	hallRequestsBool := make([][2]bool, len(hallRequests))
	for row := 0; row < len(hallRequests); row++ {
		for col := 0; col < len(hallRequests[row]); col++ {
			if hallRequests[row][col] == RS_Confirmed {
				hallRequestsBool[row][col] = true
			} else if (hallRequests[row][col] == RS_Pending) && singleElevatorMode {
				hallRequestsBool[row][col] = true
			} else {
				hallRequestsBool[row][col] = false
			}
		}
	}
	return hallRequestsBool
}

func printNodeAwareness(node NodeAwareness) {
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

func SystemView(ch_sendNodeAwareness chan<- NodeAwareness,
	ch_receiveNodeAwareness <-chan NodeAwareness,
	ch_receivePeerUpdate <-chan peers.PeerUpdate,
	ch_setTransmitEnable chan<- bool,
	ch_newHallRequest <-chan elevio.ButtonEvent,
	ch_compledtedHallRequest <-chan elevio.ButtonEvent,
	ch_elevState <-chan singleelevator.ElevState,
	ch_hallRequests chan<- [][2]bool,
	ch_initCabRequests chan<- []bool,
	ch_hraInput chan<- SystemAwareness) {

	var myNodeAwareness NodeAwareness
	var systemAwareness SystemAwareness
	var peersAlive PeersAlive
	var singleElevatorMode = true

	myNodeAwareness.InitNodeAwareness()
	systemAwareness.InitSystemAwareness()

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
				ch_setTransmitEnable <- false

			} else if singleElevatorMode && len(peersAlive) > 1 {
				singleElevatorMode = false
				// We must set all no order to unknown
				myNodeAwareness.ChangeNoOrderAndConfirmedToUnknown()
				// Update system awareness hall requests
				systemAwareness.SystemHallRequests[config.LocalID] = myNodeAwareness.HallRequests
				ch_setTransmitEnable <- true
				// else single elevator mode false
			} else if len(peersAlive) > 1 && singleElevatorMode {
				singleElevatorMode = false
			}

			for _, lostPeer := range peersLost {
				// If this node can be found in lostPeer, we should delete it from the systemAwareness
				delete(systemAwareness.SystemNodesAvailable, lostPeer)
				delete(systemAwareness.SystemElevState, lostPeer)
				delete(systemAwareness.SystemHallRequests, lostPeer)

			}

			// print peer update
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)

			// Here I can add if I am in an init state, I should send cab call of LocalID on channel init_cab_requests
			// This will be done in the init state of the elevator
		case nodeAwareness := <-ch_receiveNodeAwareness:
			nodeID := nodeAwareness.ID
			// Break out of case if IsPeerAlive returns false
			if !peersAlive.IsPeerAlive(nodeID) || nodeID == config.LocalID {
				break
			}
			systemAwareness.SystemNodesAvailable[nodeID] = nodeAwareness.IsAvailable
			systemAwareness.SystemElevState[nodeID] = nodeAwareness.ElevState
			systemAwareness.SystemHallRequests[nodeID] = nodeAwareness.HallRequests
			myNodeAwareness.CabRequests[nodeID] = nodeAwareness.ElevState.CabRequests

			// Gjennomfører sammenlikningen
			hallRequests := updateMyHallRequestView(systemAwareness.SystemHallRequests)

			// Her skal både myNodeAwareness og min id på systemAwareness oppdateres
			systemAwareness.SystemHallRequests[config.LocalID] = hallRequests
			myNodeAwareness.HallRequests = hallRequests

			// Send systemAwareness til hra modulen
			ch_hraInput <- systemAwareness

			// Etter dette må vi sende til panelmodulen for å kunne sette lys.
			ch_hallRequests <- convertHallRequestStateToBool(hallRequests, singleElevatorMode)

			// Debug print
			printNodeAwareness(nodeAwareness)

		case newHallRequest := <-ch_newHallRequest:
			// Her skal vi oppdatere vår egen hall request
			myNodeAwareness.HallRequests[newHallRequest.Floor][int(newHallRequest.Button)] = RS_Pending
			systemAwareness.SystemHallRequests[config.LocalID] = myNodeAwareness.HallRequests

			// Denne trengs vel bare i singel elevator mode

			if singleElevatorMode {
				ch_hallRequests <- convertHallRequestStateToBool(myNodeAwareness.HallRequests, singleElevatorMode)
				ch_hraInput <- systemAwareness
				// nice print of output of convertHallRequestStateToBool
			}
			// nice print to check if we are in single elevator mode
			fmt.Println("Single elevator mode: ", singleElevatorMode)
			// nice print of the new hall request
			//fmt.Println("New hall request: ", newHallRequest.Floor, newHallRequest.Button)
		case completedHallRequest := <-ch_compledtedHallRequest:
			// Her skal vi oppdatere vår egen hall request
			nextRS := RS_Completed
			if singleElevatorMode {
				nextRS = RS_NoOrder
			}
			myNodeAwareness.HallRequests[completedHallRequest.Floor][int(completedHallRequest.Button)] = nextRS
			systemAwareness.SystemHallRequests[config.LocalID] = myNodeAwareness.HallRequests

			// nice print of the completed hall request
			fmt.Println("Completed hall request: ", completedHallRequest.Floor, completedHallRequest.Button)

		case elevState := <-ch_elevState:
			// Her skal vi oppdatere vår egen elevstate
			myNodeAwareness.ElevState = elevState
			systemAwareness.SystemElevState[config.LocalID] = elevState

		case <-time.After(5 * time.Second):
			// Her skal vi sende vår egen nodeawareness på nettverket
			ch_sendNodeAwareness <- myNodeAwareness

			printNodeAwareness(myNodeAwareness)
		}
	}
}
