package main

import (
	"ElevatorProject/door_timer"
	"ElevatorProject/elevio"
	"fmt"
	"math"
)

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
func InitOrders() {
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
func any_orders() bool {
	for i := 0; i < n_floors; i++ {
		if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
			return true
		}
	}
	return false
}

func any_orders_past_this_floor_in_direction(floor int, dir elevio.MotorDirection) bool {
	if dir == elevio.MD_Stop {
		panic("any_orders_past_this_floor_in_direction: dir == MD_Stop")
	}
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

func remove_orders_and_btn_lights(floor int) {
	Current_orders.Cab_orders[floor] = false
	Current_orders.Down_orders[floor] = false
	Current_orders.Up_orders[floor] = false
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
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
		fmt.Println("find_closest_floor_order returns -1, ie there are no orders.")
	}
	return closest_order
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
		remove_orders_and_btn_lights(floor)
	}
}

func HandleNewOrder(new_order elevio.ButtonEvent) {
	if Elevator_states.moving {
		add_order_to_system(new_order)
		return
	}
	if new_order.Floor == Elevator_states.last_floor {
		elevio.SetDoorOpenLamp(true)
		Elevator_states.door_open = true
		door_timer.StartTimer()
		return
	}
	add_order_to_system(new_order)
	if Elevator_states.door_open {
		return
	}
	if new_order.Floor-elevio.GetFloor() > 0 {
		elevio.SetMotorDirection(elevio.MD_Up)
		Elevator_states.last_direction = elevio.MD_Up
		Elevator_states.moving = true
	} else {
		elevio.SetMotorDirection(elevio.MD_Down)
		Elevator_states.last_direction = elevio.MD_Down
		Elevator_states.moving = true
	}
}

func HandleDoorClosing() {
	if elevio.GetObstruction() {
		door_timer.StartTimer()
		return
	}
	elevio.SetDoorOpenLamp(false)
	Elevator_states.door_open = false
	if !any_orders() {
		return
	}
	if Elevator_states.last_direction == elevio.MD_Stop {
		fmt.Println("HandleDoorsClosing with dir = MD_Stop")
		closest_order := find_closest_floor_order(Elevator_states.last_floor)
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
		panic("HandleDoorClosing: dir = MD_Stop and order in current floor error!")
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

	// TODO:
	// Change logic of when orders are removed from the system if that is what is needed from
	// specification requirements. In other words:
	// Cab orders can be removed when opening the door.
	// Up and Down orders can only be removed if moving in that direction happens. So change some
	// removing orders logic and change where and when they are removed in the code.
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
