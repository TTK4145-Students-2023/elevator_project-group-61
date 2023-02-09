package main

import (
	"ElevatorProject/door_timer"
	"ElevatorProject/elevio"
	"fmt"
)

var n_floors int = 4

// States struct and methods
type States struct {
	// If this is changed, remember to change:
	// - InitStates()
	last_floor     int
	last_direction elevio.MotorDirection
	door_open      bool
	moving         bool
}

func (states *States) SetLastFloor(floor int) {
	states.last_floor = floor
	elevio.SetFloorIndicator(floor)
}

func (states *States) SetDirection(dir elevio.MotorDirection) {
	switch dir {
	case elevio.MD_Up:
		states.last_direction = elevio.MD_Up
		states.moving = true
		elevio.SetMotorDirection(elevio.MD_Up)
	case elevio.MD_Down:
		states.last_direction = elevio.MD_Down
		states.moving = true
		elevio.SetMotorDirection(elevio.MD_Down)
	case elevio.MD_Stop:
		states.moving = false
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
}

func (states *States) SetDoorOpen(open bool) {
	states.door_open = open
	elevio.SetDoorOpenLamp(open)
}

func (states *States) InitStates() {
	states.last_floor = -1
	states.last_direction = elevio.MD_Stop
	states.door_open = false
	states.moving = false
}

func (states States) GetLastFloor() int {
	return states.last_floor
}

func (states States) GetLastDirection() elevio.MotorDirection {
	return states.last_direction
}

func (states States) GetDoorOpen() bool {
	return states.door_open
}

func (states States) GetMoving() bool {
	return states.moving
}

// Global variables
var Elev_states States

var Active_orders Orders

// Functions
func InitLamps() {
	elevio.SetDoorOpenLamp(false)
	for i := 0; i < n_floors; i++ {
		for j := 0; j < 3; j++ {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, false)
		}
	}
}

func init_elevator() {
	InitLamps()
	Active_orders.InitOrders()
	Elev_states.InitStates()
	if elevio.GetFloor() == -1 {
		Elev_states.SetDirection(elevio.MD_Up)
	} else {
		Elev_states.SetLastFloor(elevio.GetFloor())
	}
}

func elevator_should_stop_after_sensing_floor(floor int) bool {
	if !Active_orders.AnyOrder() {
		return true
	}
	if Active_orders.GetSpecificOrder(floor, elevio.BT_Cab) {
		return true
	}
	switch Elev_states.GetLastDirection() {
	case elevio.MD_Up:
		if Active_orders.GetSpecificOrder(floor, elevio.BT_HallUp) || !Active_orders.AnyOrderPastFloorInDir(floor, elevio.MD_Up) {
			return true
		}
		return false
	case elevio.MD_Down:
		if Active_orders.GetSpecificOrder(floor, elevio.BT_HallDown) || !Active_orders.AnyOrderPastFloorInDir(floor, elevio.MD_Down) {
			return true
		}
		return false
	}
	return false
}

// Functions for handling events
func HandleFloorSensor(floor int) {
	Elev_states.SetLastFloor(floor)
	if elevator_should_stop_after_sensing_floor(floor) {
		Elev_states.SetDirection(elevio.MD_Stop)
		Elev_states.SetDoorOpen(true)
		door_timer.StartTimer()
		Active_orders.RemoveOrderDirection(floor, elevio.MD_Stop)
		Active_orders.RemoveOrderDirection(floor, Elev_states.last_direction)
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
	if Elev_states.GetMoving() {
		return
	}
	if new_order.Floor != Elev_states.GetLastFloor() {
		if !Elev_states.GetDoorOpen() {
			if new_order.Floor > Elev_states.GetLastFloor() {
				Elev_states.SetDirection(elevio.MD_Up)
			} else {
				Elev_states.SetDirection(elevio.MD_Down)
			}
		}
		return
	}
	if new_order.Button == elevio.BT_Cab || BtnTypeToDir(new_order.Button) == Elev_states.GetLastDirection() {
		Elev_states.SetDoorOpen(true)
		door_timer.StartTimer()
		Active_orders.RemoveOrderDirection(new_order.Floor, BtnTypeToDir(new_order.Button))
		return
	}
	if Elev_states.GetLastDirection() != elevio.MD_Stop {
		return
	}
	Elev_states.SetDoorOpen(true)
	door_timer.StartTimer()
	Active_orders.RemoveOrderDirection(new_order.Floor, BtnTypeToDir(new_order.Button))
}

func HandleDoorClosing() {
	if elevio.GetObstruction() {
		door_timer.StartTimer()
		return
	}
	if  !Active_orders.AnyOrder() {
		Elev_states.SetDoorOpen(false)
		return
	}
	if Active_orders.OrderInFloor(Elev_states.GetLastFloor()) && !Active_orders.AnyOrderPastFloorInDir(Elev_states.GetLastFloor(), Elev_states.GetLastDirection()) {
		if Elev_states.GetLastDirection() == elevio.MD_Stop {
			panic("last_direction is MD_Stop, but there is an order in this floor. This should not happen! (HandleDoorClosing)")
		}
		Elev_states.SetDoorOpen(true)
		door_timer.StartTimer()
		if Elev_states.GetLastDirection() == elevio.MD_Up {
			Active_orders.RemoveOrderDirection(Elev_states.last_floor, elevio.MD_Down)
		} else {
			Active_orders.RemoveOrderDirection(Elev_states.last_floor, elevio.MD_Up)
		}
		return
	}
	if Elev_states.GetLastDirection() == elevio.MD_Stop {
		fmt.Println("HandleDoorsClosing with dir = MD_Stop")
		closest_order := Active_orders.FindClosestOrder(Elev_states.GetLastFloor())
		if closest_order == Elev_states.GetLastFloor() {
			panic("closest_order == last_floor, but there are no orders in this floor. This should not happen! (HandleDoorClosing2)")
		}
		if closest_order > Elev_states.GetLastFloor() {
			Elev_states.SetDirection(elevio.MD_Up)
			return
		} else if closest_order < Elev_states.GetLastFloor() {
			Elev_states.SetDirection(elevio.MD_Down)
			return
		}
	}
	orders_up_bool := Active_orders.AnyOrderPastFloorInDir(Elev_states.GetLastFloor(), elevio.MD_Up)
	orders_down_bool := Active_orders.AnyOrderPastFloorInDir(Elev_states.GetLastFloor(), elevio.MD_Down)
	if (Elev_states.GetLastDirection() == elevio.MD_Up && orders_up_bool) || !orders_down_bool {
		Elev_states.SetDirection(elevio.MD_Up)
		return
	}
	if (Elev_states.last_direction == elevio.MD_Down && orders_down_bool) || !orders_up_bool {
		Elev_states.SetDirection(elevio.MD_Down)
		return
	}
}

// Finite state machine
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
