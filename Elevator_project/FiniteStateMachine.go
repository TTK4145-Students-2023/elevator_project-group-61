package main

import (
	"ElevatorProject/door_timer"
	"ElevatorProject/elevio"
	"fmt"
)

// Global variables
var n_floors int = 4

var Active_orders Orders

var Elev_states States

// Functions
func InitElevator() {
	InitLamps()
	Active_orders.InitOrders()
	Elev_states.InitStates()
	if elevio.GetFloor() == -1 {
		Elev_states.SetDirection(elevio.MD_Up)
	} else {
		Elev_states.SetLastFloor(elevio.GetFloor())
	}
}

func StopAfterSensingFloor(floor int) bool {
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

func BtnTypeToDir(btn_type elevio.ButtonType) elevio.MotorDirection {
	switch btn_type {
	case elevio.BT_HallUp:
		return elevio.MD_Up
	case elevio.BT_HallDown:
		return elevio.MD_Down
	}
	return elevio.MD_Stop
}

// Functions for handling events
func HandleFloorSensor(floor int) {
	Elev_states.SetLastFloor(floor)
	if StopAfterSensingFloor(floor) {
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
	if !Active_orders.AnyOrder() {
		Elev_states.SetDoorOpen(false)
		return
	}
	if Active_orders.OrderInFloor(Elev_states.GetLastFloor()) && !Active_orders.AnyOrderPastFloorInDir(Elev_states.GetLastFloor(), Elev_states.GetLastDirection()) {
		door_timer.StartTimer()
		Active_orders.RemoveOrderDirection(Elev_states.last_floor, elevio.MD_Down)
		Active_orders.RemoveOrderDirection(Elev_states.last_floor, elevio.MD_Up)
		return
	}
	Elev_states.SetDoorOpen(false)
	if Elev_states.GetLastDirection() == elevio.MD_Stop {
		closest_order := Active_orders.FindClosestOrder(Elev_states.GetLastFloor())
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
	InitElevator()
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
