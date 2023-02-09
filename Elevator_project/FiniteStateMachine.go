package main

import (
	"ElevatorProject/door_timer"
	"ElevatorProject/elevio"
	"fmt"
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

var Active_orders Orders

// ####################
// FUNCTIONS
// ####################

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
	Active_orders.InitOrders()
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

func elevator_should_stop_after_sensing_floor(floor int) bool {
	if !Active_orders.AnyOrder() {
		return true
	}
	if Active_orders.Cab_orders[floor] {
		return true
	}
	switch Elevator_states.last_direction {
	case elevio.MD_Up:
		if Active_orders.Up_orders[floor] || !Active_orders.AnyOrderPastFloorInDir(floor, elevio.MD_Up) {
			return true
		}
		return false
	case elevio.MD_Down:
		if Active_orders.Down_orders[floor] || !Active_orders.AnyOrderPastFloorInDir(floor, elevio.MD_Down) {
			return true
		}
		return false
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
		Active_orders.RemoveOrderDirection(floor, elevio.MD_Stop)
		Active_orders.RemoveOrderDirection(floor, Elevator_states.last_direction)
		if floor == 0 {
			Active_orders.RemoveOrderDirection(0, elevio.MD_Up)
		}
		if floor == n_floors-1 {
			Active_orders.RemoveOrderDirection(n_floors-1, elevio.MD_Down)
		}
	}
}

func HandleNewOrder(new_order elevio.ButtonEvent) {
	Active_orders.AddOrder(new_order)
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
	if new_order.Button == elevio.BT_Cab || BtnTypeToDir(new_order.Button) == Elevator_states.last_direction {
		elevio.SetDoorOpenLamp(true)
		Elevator_states.door_open = true
		door_timer.StartTimer()
		Active_orders.RemoveOrderDirection(new_order.Floor, BtnTypeToDir(new_order.Button))
		return
	}
	if Elevator_states.last_direction != elevio.MD_Stop {
		return
	}
	elevio.SetDoorOpenLamp(true)
	Elevator_states.door_open = true
	door_timer.StartTimer()
	Active_orders.RemoveOrderDirection(new_order.Floor, BtnTypeToDir(new_order.Button))
}

func HandleDoorClosing() {
	if elevio.GetObstruction() {
		door_timer.StartTimer()
		return
	}
	if  !Active_orders.AnyOrder() {
		elevio.SetDoorOpenLamp(false)
		Elevator_states.door_open = false
		return
	}
	if Active_orders.OrderInFloor(Elevator_states.last_floor) && !Active_orders.AnyOrderPastFloorInDir(Elevator_states.last_floor, Elevator_states.last_direction) {
		if Elevator_states.last_direction == elevio.MD_Stop {
			panic("last_direction is MD_Stop, but there is an order in this floor. This should not happen! (HandleDoorClosing2)")
		}
		elevio.SetDoorOpenLamp(true)
		Elevator_states.door_open = true
		door_timer.StartTimer()
		if Elevator_states.last_direction == elevio.MD_Up {
			Active_orders.RemoveOrderDirection(Elevator_states.last_floor, elevio.MD_Down)
		} else {
			Active_orders.RemoveOrderDirection(Elevator_states.last_floor, elevio.MD_Up)
		}
		return
	}
	if Elevator_states.last_direction == elevio.MD_Stop {
		fmt.Println("HandleDoorsClosing with dir = MD_Stop")
		closest_order := Active_orders.FindClosestOrder(Elevator_states.last_floor)
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
	orders_up_bool := Active_orders.AnyOrderPastFloorInDir(Elevator_states.last_floor, elevio.MD_Up)
	orders_down_bool := Active_orders.AnyOrderPastFloorInDir(Elevator_states.last_floor, elevio.MD_Down)
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
