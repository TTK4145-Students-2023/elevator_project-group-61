package main

import (
	"ElevatorProject/door_timer"
	"ElevatorProject/elevio"
	"fmt"

	"golang.org/x/text/cases"
)

// Constants
const id = 1
const n_floors int = 4

// Helper functions
func StopAfterSensingFloor(floor int, elev_states States, active_orders Orders) bool {
	if floor == 0 || floor == n_floors-1 {
		return true
	}
	if !active_orders.AnyOrder() {
		return true
	}
	if active_orders.GetSpecificOrder(floor, elevio.BT_Cab) {
		return true
	}
	switch elev_states.GetLastDirection() {
	case elevio.MD_Up:
		if active_orders.GetSpecificOrder(floor, elevio.BT_HallUp) || !active_orders.AnyOrderPastFloorInDir(floor, elevio.MD_Up) {
			return true
		}
		return false
	case elevio.MD_Down:
		if active_orders.GetSpecificOrder(floor, elevio.BT_HallDown) || !active_orders.AnyOrderPastFloorInDir(floor, elevio.MD_Down) {
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
func HandleFloorSensor(floor int, elev_states States, active_orders Orders) (States, Orders, bool){ 
	stop_bool := false
	elev_states.SetLastFloor(floor) 
	if StopAfterSensingFloor(floor, elev_states, active_orders) {
		stop_bool = true
		elev_states.SetElevatorBehaviour("DoorOpen")
		active_orders.RemoveOrderDirection(floor, elevio.MD_Stop)
		active_orders.RemoveOrderDirection(floor, elev_states.GetLastDirection())
		// TODO: Delegate orders to other elevators (update global orders or similar)
		if floor == 0 {
			active_orders.RemoveOrderDirection(floor, elevio.MD_Up)
			// TODO: Delegate orders to other elevators (update global orders or similar)
		}
		if floor == n_floors-1 {
			active_orders.RemoveOrderDirection(floor, elevio.MD_Down)
			// TODO: Delegate orders to other elevators (update global orders or similar)
		}
	}
	return elev_states, active_orders, stop_bool
}

func HandleNewOrder(new_order elevio.ButtonEvent, elev_states States, active_orders Orders) (States, Orders, bool, bool) { 
	set_direction_bool := false
	open_door_bool := false
	active_orders.AddOrder(new_order)
	switch(elev_states.GetElevatorBehaviour()){
	case Moving:
		return elev_states, active_orders, set_direction_bool, open_door_bool
	case DoorOpen:
		if new_order.Floor == elev_states.GetLastFloor() &&
		(new_order.Button == elevio.BT_Cab || 
		BtnTypeToDir(new_order.Button) == elev_states.GetLastDirection() || 
		elev_states.GetLastFloor() == 0 || 
		elev_states.GetLastFloor() == n_floors-1) {
			open_door_bool = true
			active_orders.RemoveOrderDirection(new_order.Floor, BtnTypeToDir(new_order.Button))
		}
	case Idle:
		if new_order.Floor != elev_states.GetLastFloor() {
			set_direction_bool = true
			if new_order.Floor > elev_states.GetLastFloor() {
				elev_states.SetDirection(elevio.MD_Up)
				elev_states.SetElevatorBehaviour("Moving")
			} else {
				elev_states.SetDirection(elevio.MD_Down)
				elev_states.SetElevatorBehaviour("Moving")
			}
		} else {
			open_door_bool = true
			elev_states.SetElevatorBehaviour("DoorOpen")
			active_orders.RemoveOrderDirection(new_order.Floor, BtnTypeToDir(new_order.Button))
		}
	}
	return elev_states, active_orders, set_direction_bool, open_door_bool
}

func HandleDoorClosing(elev_states States, active_orders Orders) (States, Orders, bool, bool) {
	open_door_bool := false
	set_direction_bool := false
	if !active_orders.AnyOrder() {
		elev_states.SetElevatorBehaviour("Idle")
		return elev_states, active_orders, open_door_bool, set_direction_bool
	}
	if active_orders.OrderInFloor(elev_states.GetLastFloor()) && !active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elev_states.GetLastDirection()) {
		open_door_bool = true
		if elev_states.GetLastDirection() == elevio.MD_Up {
			active_orders.RemoveOrderDirection(elev_states.GetLastFloor(), elevio.MD_Down)
		} else {
			active_orders.RemoveOrderDirection(elev_states.GetLastFloor(), elevio.MD_Up)
		}
		return elev_states, active_orders, open_door_bool, set_direction_bool
	}
	set_direction_bool = true
	elev_states.SetElevatorBehaviour("Moving")
	orders_up_bool := active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elevio.MD_Up)
	orders_down_bool := active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elevio.MD_Down)
	if (elev_states.GetLastDirection() == elevio.MD_Up && orders_up_bool) || !orders_down_bool {
		elev_states.SetDirection(elevio.MD_Up)
	} else if (elev_states.GetLastDirection() == elevio.MD_Down && orders_down_bool) || !orders_up_bool {
		elev_states.SetDirection(elevio.MD_Down)
	}
	return elev_states, active_orders, open_door_bool, set_direction_bool
}

func HandleButtonEvent(btn elevio.ButtonEvent) { 
	if btn.Button == elevio.BT_Cab {
		HandleNewOrder(btn)
		return
	}
	if Elev_states.GetMoving() || btn.Floor != Elev_states.GetLastFloor(){
		DelegateNewOrder(btn)
		return
	}
	if BtnTypeToDir(btn.Button) == Elev_states.GetLastDirection() || Elev_states.GetLastFloor() == 0 || Elev_states.GetLastFloor() == n_floors-1 {
		HandleNewOrder(btn)
		return
	}
	DelegateNewOrder(btn)
}

func HandleFinishedOrder(floor int, dir elevio.MotorDirection) {
	if dir == elevio.MD_Stop {
		Active_orders.RemoveOrderDirection(floor, dir)
		return
	}
	Active_orders.RemoveOrderDirection(floor, dir)
	DelegateFinishedOrders(floor, dir) 
}

// TODO: Change the parameters to use arrows
func Fsm_elevator(ch_btn chan elevio.ButtonEvent, ch_floor chan int, ch_door chan int, ch_new_order chan elevio.ButtonEvent) {
	var Elev_states States
	var Active_orders Orders

	// Initiate elevator
	Active_orders.InitOrders()
	InitLamps(Active_orders)
	Elev_states.InitStates()
	if elevio.GetFloor() == -1 {
		Elev_states.SetDirection(elevio.MD_Up)
	} else {
		Elev_states.SetLastFloor(elevio.GetFloor())
	}

	// Finite state machine
	// TODO: Add order lights functionality for all events
	// TODO: Handle order delegation, both when adding and removing.
	// TODO: That means in HandleFloorSensor, HandleNewOrder, 
	for {
		select {
		case floor := <-ch_floor:
			fmt.Println("HandleFloorSensor")
			elevio.SetFloorIndicator(floor)
			var stop_bool bool
			Elev_states, Active_orders, stop_bool = HandleFloorSensor(floor, Elev_states, Active_orders)
			if stop_bool {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				door_timer.StartTimer() // maybe change to use golang timer
			}
		case new_order := <-ch_new_order: 
			fmt.Println("HandleNewOrder")
			var set_direction_bool, open_door_bool bool
			Elev_states, Active_orders, set_direction_bool, open_door_bool = HandleNewOrder(new_order, Elev_states, Active_orders)
			if set_direction_bool {
				elevio.SetMotorDirection(Elev_states.GetLastDirection())
			}
			if open_door_bool {
				elevio.SetDoorOpenLamp(true)
				door_timer.StartTimer()
			}
		case <-ch_door:
			fmt.Println("HandleDoorClosing")
			if elevio.GetObstruction() {
				door_timer.StartTimer()
				break
			}
			var set_direction_bool, open_door_bool bool
			Elev_states, Active_orders, set_direction_bool, open_door_bool = HandleDoorClosing(Elev_states, Active_orders)
			if set_direction_bool {
				elevio.SetMotorDirection(Elev_states.GetLastDirection())
			}
			if open_door_bool {
				elevio.SetDoorOpenLamp(true)
				door_timer.StartTimer()
			} else {
				elevio.SetDoorOpenLamp(false)
			}

		case btn_press := <-ch_btn:
			//TODO: Finish this one, (done the others I think)
			fmt.Println("HandleButtonEvent")
			HandleButtonEvent(btn_press)
		}
	}
}
