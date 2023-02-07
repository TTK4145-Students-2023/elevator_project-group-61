package main

import (
	"ElevatorProject/door_timer"
	"ElevatorProject/elevio"
	"fmt"
	"math"
)

// TODO: Maybe add more class like structure and methods to structs and types! Like Morten said!

// ####################
// TYPES AND VARIABLES
// ####################
var n_floors int = 4

type States struct {
	// If this is changed, remember to change:
	// - InitStates()
	last_floor     int
	last_direction elevio.MotorDirection
	door_open      bool
	moving         bool
}

var Elevator_states States

type Orders struct {
	Up_orders   []bool
	Down_orders []bool
	Cab_orders  []bool
}

var Current_orders Orders

// ####################
// FUNCTIONS
// ####################
func InitOrders() { // RECREATED
	Current_orders.Cab_orders = make([]bool, n_floors)
	Current_orders.Up_orders = make([]bool, n_floors)
	Current_orders.Down_orders = make([]bool, n_floors)
	for i := 0; i < n_floors; i++ {
		Current_orders.Cab_orders[i] = false
		Current_orders.Up_orders[i] = false
		Current_orders.Down_orders[i] = false
	}
}

func InitLamps() {
	elevio.SetDoorOpenLamp(false)
	for i := 0; i < n_floors; i++ {
		for j := 0; j < 3; j++ {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, false)
		}
	}
}

func InitStates() {
	Elevator_states.last_floor = -1
	Elevator_states.last_direction = elevio.MD_Stop
	Elevator_states.door_open = false
	Elevator_states.moving = false
}

func init_elevator() {
	InitOrders()
	InitLamps()
	InitStates()
	elevio.SetMotorDirection(elevio.MD_Stop)
	if elevio.GetFloor() == -1 {
		elevio.SetMotorDirection(elevio.MD_Up)
		Elevator_states.last_direction = elevio.MD_Up
		Elevator_states.moving = true
	} else {
		Elevator_states.last_floor = elevio.GetFloor()
		elevio.SetFloorIndicator(Elevator_states.last_floor)
	}
}

// MINI FUNCTIONS START ################### MINI FUNCTIONS START ###################
func any_orders() bool { //RECREATED
	for i := 0; i < n_floors; i++ {
		if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
			return true
		}
	}
	return false
}

func any_orders_past_this_floor_in_direction(floor int, dir elevio.MotorDirection) bool { //RECREATED
	switch dir {
	case elevio.MD_Up:
		if floor == n_floors-1 {
			return false
		}
		for i := floor + 1; i < n_floors; i++ {
			if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
				return true
			}
		}
	case elevio.MD_Down:
		if floor == 0 {
			return false
		}
		for i := floor - 1; i > -1; i-- {
			if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
				return true
			}
		}
	}
	return false
}

func elevator_should_stop_after_sensing_floor(floor int) bool {
	if !any_orders() {
		return true
	}
	if Current_orders.Cab_orders[floor] {
		return true
	}
	switch Elevator_states.last_direction {
	case elevio.MD_Up:
		if Current_orders.Up_orders[floor] || !any_orders_past_this_floor_in_direction(floor, elevio.MD_Up) {
			return true
		}
		return false
	case elevio.MD_Down:
		if Current_orders.Down_orders[floor] || !any_orders_past_this_floor_in_direction(floor, elevio.MD_Down) {
			return true
		}
		return false
	}
	fmt.Println("elevator_should_stop_after_sensing_floor with dir = MD_Stop")
	return false
}

func add_order_to_system(btn elevio.ButtonEvent) {
	elevio.SetButtonLamp(btn.Button, btn.Floor, true)
	switch btn.Button {
	case elevio.BT_HallUp:
		Current_orders.Up_orders[btn.Floor] = true
	case elevio.BT_Cab:
		Current_orders.Cab_orders[btn.Floor] = true
	case elevio.BT_HallDown:
		Current_orders.Down_orders[btn.Floor] = true
	}
}

func find_closest_floor_order(floor int) int {
	closest_order := -1
	shortest_diff := n_floors
	for i := 0; i < n_floors; i++ {
		if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
			diff := math.Abs(float64(floor - i))
			if diff < float64(shortest_diff) {
				shortest_diff = int(diff)
				closest_order = i
			}
		}
	}
	if closest_order == -1 {
		panic("find_closest_floor_order returns -1, ie there are no orders.")
	}
	return closest_order
}

func remove_order_direction(floor int, dir elevio.MotorDirection) {
	if dir == elevio.MD_Up && Current_orders.Up_orders[floor] {
		Current_orders.Up_orders[floor] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	} 
	if dir == elevio.MD_Down && Current_orders.Down_orders[floor] {
		Current_orders.Down_orders[floor] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	}
	if dir == elevio.MD_Stop && Current_orders.Cab_orders[floor] {
		Current_orders.Cab_orders[floor] = false
		elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	}
}

func btn_type_to_direction(btn_type elevio.ButtonType) elevio.MotorDirection {
	switch btn_type {
	case elevio.BT_HallUp:
		return elevio.MD_Up
	case elevio.BT_HallDown:
		return elevio.MD_Down
	}
	return elevio.MD_Stop
}

func order_in_floor(floor int) bool {
	if Current_orders.Up_orders[floor] || Current_orders.Down_orders[floor] || Current_orders.Cab_orders[floor] {
		return true
	}
	return false
}
// MINI FUNCTIONS STOP  ################### MINI FUNCTIONS STOP  ###################

func HandleFloorSensor(floor int) {
	elevio.SetFloorIndicator(floor)
	Elevator_states.last_floor = floor
	if elevator_should_stop_after_sensing_floor(floor) {
		elevio.SetMotorDirection(elevio.MD_Stop)
		Elevator_states.moving = false
		elevio.SetDoorOpenLamp(true)
		Elevator_states.door_open = true
		door_timer.StartTimer()
		remove_order_direction(floor, elevio.MD_Stop)
		remove_order_direction(floor, Elevator_states.last_direction)
		if floor == 0 {
			remove_order_direction(0, elevio.MD_Up)
		}
		if floor == n_floors - 1 {
			remove_order_direction(n_floors - 1, elevio.MD_Down)
		}	
	}
}

func HandleNewOrder(new_order elevio.ButtonEvent) {
	add_order_to_system(new_order)
	if Elevator_states.moving {
		return
	}
	if new_order.Floor != Elevator_states.last_floor {
		if !Elevator_states.door_open { 
			if new_order.Floor > Elevator_states.last_floor {
				elevio.SetMotorDirection(elevio.MD_Up)
				Elevator_states.last_direction = elevio.MD_Up
				Elevator_states.moving = true
			} else {
				elevio.SetMotorDirection(elevio.MD_Down)
				Elevator_states.last_direction = elevio.MD_Down
				Elevator_states.moving = true
			}	
		}
		return
	}
	if (new_order.Button == elevio.BT_Cab || btn_type_to_direction(new_order.Button) == Elevator_states.last_direction) {
		elevio.SetDoorOpenLamp(true)
		Elevator_states.door_open = true
		door_timer.StartTimer()
		remove_order_direction(new_order.Floor, btn_type_to_direction(new_order.Button))
		return
	}
	if Elevator_states.last_direction != elevio.MD_Stop {
		return
	}
	elevio.SetDoorOpenLamp(true)
	Elevator_states.door_open = true
	door_timer.StartTimer()
	remove_order_direction(new_order.Floor, btn_type_to_direction(new_order.Button))
}

func HandleDoorClosing() {
	if elevio.GetObstruction() {
		door_timer.StartTimer()
		return
	}
	if !any_orders() {
		elevio.SetDoorOpenLamp(false)
		Elevator_states.door_open = false
		return
	}
	if order_in_floor(Elevator_states.last_floor) && !any_orders_past_this_floor_in_direction(Elevator_states.last_floor, Elevator_states.last_direction) {
		if Elevator_states.last_direction == elevio.MD_Stop {
			panic("last_direction is MD_Stop, but there is an order in this floor. This should not happen! (HandleDoorClosing2)")
		}
		elevio.SetDoorOpenLamp(true)
		Elevator_states.door_open = true
		door_timer.StartTimer()
		if Elevator_states.last_direction == elevio.MD_Up {
			remove_order_direction(Elevator_states.last_floor, elevio.MD_Down)
		} else {
			remove_order_direction(Elevator_states.last_floor, elevio.MD_Up)
		}
		return
	}
	if Elevator_states.last_direction == elevio.MD_Stop {
		fmt.Println("HandleDoorsClosing with dir = MD_Stop")
		closest_order := find_closest_floor_order(Elevator_states.last_floor)
		if closest_order == Elevator_states.last_floor {
			panic("closest_order == last_floor, but there are no orders in this floor. This should not happen! (HandleDoorClosing2)")
		}
		if closest_order > Elevator_states.last_floor {
			elevio.SetMotorDirection(elevio.MD_Up)
			Elevator_states.last_direction = elevio.MD_Up
			Elevator_states.moving = true
			return
		} else if closest_order < Elevator_states.last_floor {
			elevio.SetMotorDirection(elevio.MD_Down)
			Elevator_states.last_direction = elevio.MD_Down
			Elevator_states.moving = true
			return
		}
	}
	orders_up_bool := any_orders_past_this_floor_in_direction(Elevator_states.last_floor, elevio.MD_Up)
	orders_down_bool := any_orders_past_this_floor_in_direction(Elevator_states.last_floor, elevio.MD_Down)
	if (Elevator_states.last_direction == elevio.MD_Up && orders_up_bool) || !orders_down_bool {
		elevio.SetMotorDirection(elevio.MD_Up)
		Elevator_states.last_direction = elevio.MD_Up
		Elevator_states.moving = true
		return
	}
	if (Elevator_states.last_direction == elevio.MD_Down && orders_down_bool) || !orders_up_bool {
		elevio.SetMotorDirection(elevio.MD_Down)
		Elevator_states.last_direction = elevio.MD_Down
		Elevator_states.moving = true
		return
	}
}

func Fsm_elevator(ch_order chan elevio.ButtonEvent, ch_floor chan int, ch_door chan int) {
	init_elevator()
	for {
		select {
		case floor := <-ch_floor:
			fmt.Println("HandleFloorSensor")
			HandleFloorSensor(floor)
		case new_order := <-ch_order:
			fmt.Println("HandleNewOrder")
			HandleNewOrder(new_order)
		case <-ch_door:
			fmt.Println("HandleDoorClosing")
			HandleDoorClosing()
		}
	}
}
