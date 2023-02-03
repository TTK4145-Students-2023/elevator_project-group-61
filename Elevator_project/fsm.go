package main

import (
	"ElevatorProject/door_timer"
	"ElevatorProject/elevio"
	"fmt"
	"time"
)

func TestFunction() {
	fmt.Println("Hey there buddy")
}

//
// Types, variables, consts and structs
//

type States struct {
	Last_floor     int
	Last_direction elevio.MotorDirection
	Door_open      bool
}

var Elevator_states States = States{-1, 0, false}

// Functions
func InitElevatorStates() {
	Elevator_states = States{-1, 0, false}
}

func CheckStop(current_floor int) bool {
	if !CheckIfAnyOrders() {
		return true
	}
	if current_floor < 0 || current_floor > 3 {
		fmt.Println("Error: Floor out of range")
		return false
	}
	if Current_orders.Cab_orders[current_floor] {
		return true
	}
	if Current_orders.Down_orders[current_floor] && Elevator_states.Last_direction == elevio.MD_Down {
		return true
	}
	if Current_orders.Up_orders[current_floor] && Elevator_states.Last_direction == elevio.MD_Up {
		return true
	}
	return false
}

func FindNextDirection(current_floor int) elevio.MotorDirection {
	anyorder := CheckIfAnyOrders()
	if !anyorder {
		fmt.Println("No orders")
		return elevio.MD_Stop
	}
	if OrdersAbove(current_floor) && (Elevator_states.Last_direction == elevio.MD_Up || Elevator_states.Last_direction == elevio.MD_Stop) {
		fmt.Println("Orders above")
		return elevio.MD_Up
	}
	if OrdersBelow(current_floor) && (Elevator_states.Last_direction == elevio.MD_Down || Elevator_states.Last_direction == elevio.MD_Stop) {
		fmt.Println("Orders below")
		return elevio.MD_Down
	}
	return elevio.MD_Stop
}

func InitFSM() {
	InitElevatorStates()
	if elevio.GetFloor() != 3 {
		elevio.SetMotorDirection(elevio.MD_Up)
		Elevator_states.Last_direction = elevio.MD_Up
	} else if elevio.GetFloor() != 0 {
		elevio.SetMotorDirection(elevio.MD_Down)
		Elevator_states.Last_direction = elevio.MD_Down
	}
}

func HandleButton(button elevio.ButtonEvent) {
	elevio.SetButtonLamp(button.Button, button.Floor, true)
	if button.Button == elevio.BT_Cab {
		Current_orders.Cab_orders[button.Floor] = true
	}
	if button.Button == elevio.BT_HallUp {
		Current_orders.Up_orders[button.Floor] = true
	}
	if button.Button == elevio.BT_HallDown {
		Current_orders.Down_orders[button.Floor] = true
	}
	if Elevator_states.Last_direction == elevio.MD_Stop {
		elevio.SetMotorDirection(FindNextDirection(Elevator_states.Last_floor))
	}
}

func HandleFloorSensor(floor int) {
	fmt.Println("Floor sensor activated")
	elevio.SetFloorIndicator(floor)
	Elevator_states.Last_floor = floor
	if CheckStop(floor) {
		fmt.Println("Stop")
		RemoveOrders(floor)
		elevio.SetMotorDirection(elevio.MD_Stop)
		Elevator_states.Last_direction = elevio.MD_Stop
		Elevator_states.Door_open = true
		elevio.SetDoorOpenLamp(true)
		door_timer.StartTimer()
	}
}

func HandleDefault() {
	if Elevator_states.Last_direction == elevio.MD_Stop {
		elevio.SetMotorDirection(FindNextDirection(Elevator_states.Last_floor))
	}
}

func FinalStateMachine(ch_buttons chan elevio.ButtonEvent, ch_floorsensor chan int, ch_timer chan int) {
	InitFSM()
	for {
		select {
		case button := <-ch_buttons:
			HandleButton(button)
		case floor := <-ch_floorsensor:
			HandleFloorSensor(floor)
		case <-ch_timer:
			Elevator_states.Door_open = false
			elevio.SetDoorOpenLamp(false)
		default:
			HandleDefault()
			time.Sleep(100 * time.Millisecond)
		}
	}
}
