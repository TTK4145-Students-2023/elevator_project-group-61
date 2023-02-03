package main

import (
	"ElevatorProject/elevio"
)

// 
// Types, variables, consts and structs
//
type Direction int

const (
	Direction_up Direction = 1
	Direction_still = 0
	Direction_down Direction = - 1	
)

type States struct {
	Last_floor int
	Last_direction Direction
}

var Elevator_states States = States{-1, 0}

//
// Functions
//
func InitElevatorStates() {
	Elevator_states = States{-1, 0}
}

func CheckStop(current_floor int) bool{ 
	if Current_orders.Cab_orders[current_floor] {
		return true
	}
	if Current_orders.Down_orders[current_floor] && Elevator_states.Last_direction == Direction_down {
		return true
	}
	if Current_orders.Up_orders[current_floor] && Elevator_states.Last_direction == Direction_up {
		return true
	}
	return false
}

func FindNextDirection(current_floor int) Direction {
	anyorder := CheckIfAnyOrders()
	if !anyorder {
		return Direction_still
	}
	if OrdersAbove(current_floor) && Elevator_states.Last_direction == Direction_up {
		return Direction_up
	}
	if OrdersBelow(current_floor) && Elevator_states.Last_direction == Direction_down {
		return Direction_down
	}
	return Direction_still
}


func fsm() {
	for {
		select {
		case floor := <-elevio.FloorReached:
			Elevator_states.Last_floor = floor
			if CheckStop(floor) {
				elevio.SetDoorOpenLamp(true)
				Current_orders.Cab_orders[floor] = false
				Current_orders.Down_orders[floor] = false
				Current_orders.Up_orders[floor] = false
				<-time.After(3 * time.Second)
				elevio.SetDoorOpenLamp(false)
			}
			Elevator_states.Last_direction = FindNextDirection(floor)
			elevio.SetMotorDirection(Elevator_states.Last_direction)
		case <-time.After(100 * time.Millisecond):
			Elevator_states.Last_direction = FindNextDirection(Elevator_states.Last_floor)
			elevio.SetMotorDirection(Elevator_states.Last_direction)
		}
	}
}