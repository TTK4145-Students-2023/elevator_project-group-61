package main

import "time"

const n_floors int = 4
const m_elevators int = 3
const localID string = "0"

type RequestState int
type SingleElevatorMode bool // True if elevator is in single elevator mode
// All peers alive in a list of same type peers
type PeersAlive []string

const (
	RS_Unknown   RequestState = 0
	RS_NoOrder                = 1
	RS_Pending                = 2
	RS_Confirmed              = 3
)

type NodeAwareness struct {
	ID            string
	ElevatorState States
	HallRequests  [][2]RequestState // n number of floors
	CabRequests   map[string][]bool
}

type SystemAwareness struct {
	SystemElevState    map[string]States
	SystemHallRequests map[string][][2]RequestState
	SystemCabRequests  map[string][]bool
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
					if int(nodeView[row][col]) > max_count {
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
				RS_Pending_count := 0
				for _, nodeView := range systemHallRequests {
					if nodeView[row][col] == RS_Confirmed {
						myView[row][col] = RS_Confirmed
						break
					} else if nodeView[row][col] == RS_Pending {
						RS_Pending_count++
					}
				}
				// print len of systemHallRequests and
				//fmt.Println("len(systemHallRequests): ", len(systemHallRequests))
				if RS_Pending_count == len(systemHallRequests) {
					myView[row][col] = RS_Confirmed
				}
			case RS_Confirmed:
				for _, nodeView := range systemHallRequests {
					if nodeView[row][col] == RS_NoOrder {
						myView[row][col] = RS_NoOrder
						break
					}
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
func (nodeAwareness *NodeAwareness) ChangeNoOrderToUnknown() {
	for row := 0; row < len(nodeAwareness.HallRequests); row++ {
		for col := 0; col < len(nodeAwareness.HallRequests[row]); col++ {
			if nodeAwareness.HallRequests[row][col] == RS_NoOrder {
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
	systemAwareness.SystemElevState = make(map[string]States, m_elevators)
	systemAwareness.SystemHallRequests = make(map[string][][2]RequestState, m_elevators)
	systemAwareness.SystemCabRequests = make(map[string][]bool, m_elevators)
}

func SystemView(ch_transmit chan<- NodeAwareness,
	ch_receive <-chan NodeAwareness,
	ch_peerUpdate <-chan PeerUpdate,
	ch_peerTransmitEnable chan<- bool,
	ch_newHallRequest <-chan SpecificOrder,
	ch_compledtedHallRequest <-chan SpecificOrder,
	ch_cabRequests <-chan []bool,
	ch_elevState <-chan States,
	ch_hallRequests chan<- [][2]bool,
	ch_initCabRequests chan<- []bool,
	ch_hraInput chan<- SystemAwareness) {

	var myNodeAwareness NodeAwareness
	var systemAwareness SystemAwareness
	var singleElevatorMode SingleElevatorMode
	var peersAlive PeersAlive

	myNodeAwareness.InitNodeAwareness()
	systemAwareness.InitSystemAwareness()

	for {
		select {
		case peerUpdate := <-ch_peerUpdate:
			// Go to single elevator mode if at least one node left
			peersAlive = peerUpdate.Peers
			newPeer := peerUpdate.New
			peersLost := peerUpdate.Lost

			// Heller kjøre en metode som sletter her

			if len(peersAlive) <= 1 {
				singleElevatorMode = true
				// We must stop broadcasting our node awareness. Disable the channel
				ch_peerTransmitEnable <- false

			} else if singleElevatorMode && len(peersAlive) > 1 {
				singleElevatorMode = false
				// We must set all no order to unknown
				myNodeAwareness.ChangeNoOrderToUnknown()
				// Update system awareness hall requests
				systemAwareness.SystemHallRequests[localID] = myNodeAwareness.HallRequests
				ch_peerTransmitEnable <- true
				// else single elevator mode false
			} else {
				singleElevatorMode = false

			}

			for lostPeer := range peersLost {
				// TODO: If this node can be found in lostPeer, we should delete it from the systemAwareness
				delete(systemAwareness.SystemElevState, lostPeer)
				delete(systemAwareness.SystemHallRequests, lostPeer)
				delete(systemAwareness.SystemCabRequests, lostPeer)
			}

			// Here I can add if I am in an init state, I should send cab call of localID on channel init_cab_requests
			// This will be done in the init state of the elevator
		case nodeAwareness := <-ch_receive:
			nodeID := nodeAwareness.ID

			// Break out of case if IsPeerAlive returns false

			if peersAlive.IsPeerAlive(nodeID) == false || nodeID == localID {
				break
			}

			systemAwareness.SystemElevState[nodeID] = nodeAwareness.ElevatorState
			systemAwareness.SystemHallRequests[nodeID] = nodeAwareness.HallRequests
			systemAwareness.SystemCabRequests[nodeID] = nodeAwareness.CabRequests

			myNodeAwareness.CabRequests[nodeID] = nodeAwareness.CabRequests

			// Gjennomfører sammenlikningen
			hallRequests := updateMyHallRequestUnderstanding(systemAwareness.SystemHallRequests)

			// Her skal både myNodeAwareness og min id på systemAwareness oppdateres
			systemAwareness.SystemHallRequests[localID] = hallRequests
			myNodeAwareness.HallRequests = hallRequests

			// Etter dette må vi sende til panelmodulen for å kunne sette lys.

			ch_hallRequests <- hallRequests
			ch_hraInput <- systemAwareness

		case newHallRequest := <-ch_newHallRequest:
			// Her skal vi oppdatere vår egen hall request
			myNodeAwareness.HallRequests[newHallRequest.Floor][int(newHallRequest.Btn)] = RS_Pending
			systemAwareness.SystemHallRequests[localID] = myNodeAwareness.HallRequests

			// Denne trengs vel bare i singel elevator mode
			//ch_hallRequests <- myNodeAwareness.HallRequests
		case completedHallRequest := <-ch_compledtedHallRequest:
			// Her skal vi oppdatere vår egen hall request
			myNodeAwareness.HallRequests[completedHallRequest.Floor][int(completedHallRequest.Btn)] = RS_NoOrder
			systemAwareness.SystemHallRequests[localID] = myNodeAwareness.HallRequests

		case cabRequests := <-ch_cabRequests:
			// Her skal vi oppdatere vår egen cab request
			myNodeAwareness.CabRequests = cabRequests
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

// I singleElevatorMode never go to confirmed state. New orders will always be pending.