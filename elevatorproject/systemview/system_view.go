package systemview

import (
	"elevatorproject/network/peers"
	"elevatorproject/singleelevator"
	"elevatorproject/singleelevator/elevio"
	"time"
)

const n_floors int = 4
const m_elevators int = 3
const localID string = "0"

type RequestState int

// All peers alive in a list of same type peers

const (
	RS_Unknown   RequestState = -1
	RS_NoOrder                = 0
	RS_Pending                = 1
	RS_Confirmed              = 2
	RS_Completed              = 3
)

type PeersAlive []string

type NodeAwareness struct {
	ID            string
	IsAvailable   bool
	ElevatorState singleelevator.States
	HallRequests  [][2]RequestState // n number of floors
	CabRequests   map[string][]bool
}

type SystemAwareness struct {
	SystemElevState      map[string]singleelevator.States
	SystemHallRequests   map[string][][2]RequestState
	SystemCabRequests    map[string][]bool
	SystemNodesAvailable map[string]bool
}

func updateMyHallRequestView(systemHallRequests map[string][][2]RequestState) [][2]RequestState {
	myView := systemHallRequests[localID]
	delete(systemHallRequests, localID)

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
	nodeAwareness.ID = localID
	nodeAwareness.ElevatorState.InitStates()
	nodeAwareness.HallRequests = make([][2]RequestState, n_floors)
	nodeAwareness.CabRequests = make(map[string][]bool)
}

func (systemAwareness *SystemAwareness) InitSystemAwareness() {
	systemAwareness.SystemElevState = make(map[string]singleelevator.States, m_elevators)
	systemAwareness.SystemHallRequests = make(map[string][][2]RequestState, m_elevators)
	systemAwareness.SystemCabRequests = make(map[string][]bool, m_elevators)
}

// function that takes a [][2]RequestState as input and return [][2]bool
func convertHallRequestStateToBool(hallRequests [][2]RequestState, singleElevatorMode bool) [][2]bool {
	hallRequestsBool := make([][2]bool, len(hallRequests))
	for row := 0; row < len(hallRequests); row++ {
		for col := 0; col < len(hallRequests[row]); col++ {
			if hallRequests[row][col] == RS_Confirmed {
				hallRequestsBool[row][col] = true
			} else if (hallRequests[row][col] == RS_Pending) && singleElevatorMode {
			} else {
				hallRequestsBool[row][col] = false
			}
		}
	}
	return hallRequestsBool
}

func SystemView(ch_transmit chan<- NodeAwareness,
	ch_receive <-chan NodeAwareness,
	ch_peerUpdate <-chan peers.PeerUpdate,
	ch_peerTransmitEnable chan<- bool,
	ch_newHallRequest <-chan elevio.ButtonEvent,
	ch_compledtedHallRequest <-chan elevio.ButtonEvent,
	ch_cabRequests <-chan []bool,
	ch_elevState <-chan singleelevator.States,
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
		case peerUpdate := <-ch_peerUpdate:
			// Go to single elevator mode if at least one node left
			peersAlive = peerUpdate.Peers
			peersLost := peerUpdate.Lost
			// Heller kjøre en metode som sletter her
			if len(peersAlive) <= 1 && !singleElevatorMode {
				singleElevatorMode = true
				// We must stop broadcasting our node awareness. Disable the channel
				ch_peerTransmitEnable <- false

			} else if singleElevatorMode && len(peersAlive) > 1 {
				singleElevatorMode = false
				// We must set all no order to unknown
				myNodeAwareness.ChangeNoOrderAndConfirmedToUnknown()
				// Update system awareness hall requests
				systemAwareness.SystemHallRequests[localID] = myNodeAwareness.HallRequests
				ch_peerTransmitEnable <- true
				// else single elevator mode false
			} else if len(peersAlive) > 1 && singleElevatorMode {
				singleElevatorMode = false
			}
			
			for _, lostPeer := range peersLost {
				// If this node can be found in lostPeer, we should delete it from the systemAwareness
				delete(systemAwareness.SystemNodesAvailable, lostPeer)
				delete(systemAwareness.SystemElevState, lostPeer)
				delete(systemAwareness.SystemHallRequests, lostPeer)
				delete(systemAwareness.SystemCabRequests, lostPeer)
			}

			// Here I can add if I am in an init state, I should send cab call of localID on channel init_cab_requests
			// This will be done in the init state of the elevator
		case nodeAwareness := <-ch_receive:
			nodeID := nodeAwareness.ID
			// Break out of case if IsPeerAlive returns false
			if !peersAlive.IsPeerAlive(nodeID) || nodeID == localID {
				break
			}
			systemAwareness.SystemNodesAvailable[nodeID] = nodeAwareness.IsAvailable
			systemAwareness.SystemElevState[nodeID] = nodeAwareness.ElevatorState
			systemAwareness.SystemHallRequests[nodeID] = nodeAwareness.HallRequests
			systemAwareness.SystemCabRequests[nodeID] = nodeAwareness.CabRequests[nodeID]

			myNodeAwareness.CabRequests[nodeID] = nodeAwareness.CabRequests[nodeID]

			// Gjennomfører sammenlikningen
			hallRequests := updateMyHallRequestView(systemAwareness.SystemHallRequests)

			// Her skal både myNodeAwareness og min id på systemAwareness oppdateres
			systemAwareness.SystemHallRequests[localID] = hallRequests
			myNodeAwareness.HallRequests = hallRequests

			// Send systemAwareness til hra modulen
			ch_hraInput <- systemAwareness

			// Etter dette må vi sende til panelmodulen for å kunne sette lys.
			ch_hallRequests <- convertHallRequestStateToBool(hallRequests, singleElevatorMode)

		case newHallRequest := <-ch_newHallRequest:
			// Her skal vi oppdatere vår egen hall request
			myNodeAwareness.HallRequests[newHallRequest.Floor][int(newHallRequest.Button)] = RS_Pending
			systemAwareness.SystemHallRequests[localID] = myNodeAwareness.HallRequests

			// Denne trengs vel bare i singel elevator mode
			if singleElevatorMode {
				ch_hallRequests <- convertHallRequestStateToBool(myNodeAwareness.HallRequests, singleElevatorMode)
			}
		case completedHallRequest := <-ch_compledtedHallRequest:
			// Her skal vi oppdatere vår egen hall request
			nextRS := RS_Completed
			if singleElevatorMode {
				nextRS = RS_NoOrder
			}
			myNodeAwareness.HallRequests[completedHallRequest.Floor][int(completedHallRequest.Button)] = nextRS
			systemAwareness.SystemHallRequests[localID] = myNodeAwareness.HallRequests

		case cabRequests := <-ch_cabRequests:
			// Her skal vi oppdatere vår egen cab request
			myNodeAwareness.CabRequests[localID] = cabRequests
			systemAwareness.SystemCabRequests[localID] = cabRequests

		case elevState := <-ch_elevState:
			// Her skal vi oppdatere vår egen elevstate
			myNodeAwareness.ElevatorState = elevState
			systemAwareness.SystemElevState[localID] = elevState

		case <-time.After(15 * time.Millisecond):
			// Her skal vi sende vår egen nodeawareness på nettverket
			ch_transmit <- myNodeAwareness
		}
	}
}
