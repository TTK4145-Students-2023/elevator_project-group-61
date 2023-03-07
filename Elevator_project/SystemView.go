package main

import (
	"ElevatorProject/elevio"
	"fmt"
)

const m_elevators int = 3
const localID string = "0"

type RequestState int

const (
	RS_Unkwown RequestState = 0
	RS_NoOrder = 1
	RS_Pending = 2
	RS_Confirmed = 3
)

type NodeAwareness struct {
	ID string
	ElevatorState ElevState
	HallRequests [n_floors][]RequestState    // n number of floors
	CabRequests map[string] []bool
}

type SystemAwareness struct {
	SystemElevState map[string] ElevState
	SystemHallRequests map[string] [n_floors][]RequestState
	SystemCabRequests map[string] []bool
}

// Denne funksjonen skal ta inn systemHallRequests og oppdatere denne nodens understanding basert på hva som 
// ligger der

func updateMyHallRequestUnderstanding(systemHallRequests map[string][][2]RequestState) [][2]RequestState {
	
	return
}

func SystemView(ch_receiver <- chan NodeAwareness, ch_transmit chan <- NodeAwareness, ch_hallCalls 
	<- chan int, ch_ElevState <- chan ElevState, 
	ch_transmit chan <- NodeStatus, ch_alive <- chan []int, ch_hrainfo chan <- etellerannet ) {

		// Her skal alle lokale variabler lages
		myNodeAwareness := NodeAwareness{
			ID : LocalID,
			ElevatorState: ,
			HallRequsts : ,
			CabRequests : ,
		}

		// Må definere alive-listen her
		systemAwareness := make(map[int] NodeAwareness, m)

		for {
			select {
			case nodeAwareness := <- ch_reciever:
				nodeID := nodeAwareness.ID
				// Oppdatere akkurat denne nodens elevstate i system awareness
				systemAwareness.SystemElevstate[nodeID] = nodeAwareness.ElevatorState
				systemAwareness.SystemHallRequests[nodeID] = nodeAwareness.HallRequests
				systemAwareness.SystemCabRequests[nodeID] = nodeAwareness.CabRequestsch_ElevState
				myNodeAwareness.CabRequests[nodeID] = nodeAwareness.CabRequests

				// Gjennomfører sammenlikningen
				hallRequests := updateMyHallRequestUnderstanding(systemAwareness.SystemHallRequests)
				
				// Her skal både myNodeAwareness og min id på systemAwareness oppdateres

				systemAwareness.SystemHallRequests[localID] = hallRequests 
				myNodeAwareness.HallRequests = hallRequests


				// Etter oppdatering skal vi sende resultatet på nettverket
				ch_transmit <- myNodeAwareness

				// Etter dette må vi sende til panelmodulen for å kunne sette lys.

				ch_panel <- hallRequests

			case hall_call := <- ch_hallCalls:
				// Skrive hva som skjer hvis det kommer en ny hall call
				ch_transmit <- node_awareness
			case cab_call := <- ch_cabRequests:
				// Skrive lett hva som skjer når cab-call kommer inn
				ch_transmit <- node_awareness
			}

			case // Her må vi håndtere sletting fra systemAwareness hvis den en heis dør.
		}
	 }
