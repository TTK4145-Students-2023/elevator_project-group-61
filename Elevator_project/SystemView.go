package main

import (
	"ElevatorProject/elevio"
	"fmt"
)

const m int = m
const ID int = 0

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
	HallRequests [n][2]RequestState    // n number of floors
	CabRequests map[string] []bool
}


type SystemAwareness struct {
	map[string] [n][2]RequestState,
	map[string] ElevState
	map[string] []bool
}


func SystemView(ch_reciever <- chan NodeAwareness, ch_transmit chan <- NodeAwareness, ch_hallCalls 
	<- chan int, ch_ElevState <- chan ElevState, 
	ch_transmit chan <- NodeStatus, ch_alive <- chan []int, ch_hrainfo chan <- ) {

		// Her skal alle lokale variabler lages
		nodeAwareness := NodeAwareness{
			ID : ID,
			ElevatorState: ,
			HallRequsts : ,
			CabRequests : ,
		}

		SystemAwareness := make(map[int] NodeAwareness, m)

		
		for {
			select {
			case node_status := <- ch_reciever:
				// Hva skjer hvis det kommer en ny node_status på nettverket
				
				// Lage noe som overskriver det lokale mappet
				// Kalle en funksjon som tar dette lokale mappet som input

				ch_transmit <- node_awareness
			case hall_call := <- ch_hallCalls:
				// Skrive hva som skjer hvis det kommer en ny hall call
				ch_transmit <- node_awareness
			case cab_call := <- ch_cabCall:
				// Skrive lett hva som skjer når cab-call kommer inn
				ch_transmit <- node_awareness
			}

		}
	 }