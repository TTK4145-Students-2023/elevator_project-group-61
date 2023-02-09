package main

import (
	"ElevatorProject/elevio"
	"fmt"
)

func DelegateOrder(btn elevio.ButtonEvent) {
	// Send the order to others,
	// and possibly calculate who is going to have this order
	fmt.Print("Hey")
}

func DelegateStates(elev_states States) {
	// Update the others on your state
	fmt.Print("Hey man")
}

