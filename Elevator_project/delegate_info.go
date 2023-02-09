package main

import (
	"ElevatorProject/elevio"
	"fmt"
	"math"
)

type GlobalOrders struct {
	Up_orders   []int
	Down_orders []int
	Cab_orders  []int
}

func (states *GlobalOrders) InitGlobalOrders() {
	states.Cab_orders = make([]int, n_floors)
	states.Up_orders = make([]int, n_floors)
	states.Down_orders = make([]int, n_floors)
	for i := 0; i < n_floors; i++ {
		states.Cab_orders[i] = 0
		states.Up_orders[i] = 0
		states.Down_orders[i] = 0
	}
}

// Variables
var alive_elevators = make([]bool, 3)
var elevators = make([]States, 3)

func InitGlobalShit() {
	alive_elevators[0] = true
	alive_elevators[1] = true
	alive_elevators[2] = true
	elevators[0].InitStates()
	elevators[1].InitStates()
	elevators[2].InitStates()
}



// Functions
func DelegateNewOrder(btn elevio.ButtonEvent) {
	lucky_winner := WhoShouldTakeNewOrder(btn, alive_elevators, elevators)
	SendOrderToElevator(btn, lucky_winner) //needs creation
	InformAboutHallCall(btn, true) //needs creation
}

func DelegateStates(elev_states States) {
	// Update the others on your state
	fmt.Print("Hey man")
}

func DelegateFinishedOrders(floor int, dir elevio.MotorDirection) {
	if dir == elevio.MD_Up {
		InformAboutHallCall(elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp}, false) //needs creation
	} else if dir == elevio.MD_Down {
		InformAboutHallCall(elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown}, false) //needs creation
	} 
}

func Optimizer(elev_states States, btn elevio.ButtonEvent) float64 {
	// Just an example function that will will give some time for a single elevator
	// to complete an order. (It just takes floors away from the order.)
	return math.Abs(float64(elev_states.GetLastFloor() - btn.Floor))
}

func WhoShouldTakeNewOrder(btn elevio.ButtonEvent, alive_elevators []bool, elevators []States) int {
	closest_elev := -1
	shortest := 1000.0
	for i := 0 ; i < len(alive_elevators); i++ {
		if alive_elevators[i] {
			time := Optimizer(elevators[i], btn)
			if time < shortest {
				shortest = time
				closest_elev = i
			}
		}
	}
	return closest_elev
}

