package main

import (
	"ElevatorProject/elevio"
)

// 
// Types, variables, consts and structs
//
type Orders struct {
	Up_orders [4]bool
	Down_orders [4]bool
	Cab_orders [4]bool
}

var Current_orders Orders = Orders{
							[4]bool{false, false, false, false},
							[4]bool{false, false, false, false},
							[4]bool{false, false, false, false}}

//
// Functions
//
func InitOrders() {
	Current_orders = Orders{
		[4]bool{false, false, false, false},
		[4]bool{false, false, false, false},
		[4]bool{false, false, false, false}}
}

func CheckIfAnyOrders() bool{
	for i := 0 ; i < 4 ; i++ {
		if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
			return true
		}
	}
	return false
}	

func OrdersBelow(floor int) bool {
	if floor <= 0 {
		return false
	}
	for i := floor ; i > -1 ; i-- {
		if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
			return true
		}
	}
	return false
}

func OrdersAbove(floor int) bool {
	if floor >= 0 {
		return false
	}
	for i := floor ; i < 4 ; i++ {
		if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
			return true
		}
	}
	return false
}

func RemoveOrders(floor int) {
	if floor < 0 || floor > 3 {
		return
	}
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	Current_orders.Cab_orders[floor] = false
	Current_orders.Down_orders[floor] = false
	Current_orders.Up_orders[floor] = false
}

func UpdateOrders(button_press elevio.ButtonEvent) {
	switch button_press.Button {
	case elevio.BT_Cab:
		Current_orders.Cab_orders[button_press.Floor] = true
	case elevio.BT_HallDown:
		Current_orders.Down_orders[button_press.Floor] = true
	case elevio.BT_HallUp:
		Current_orders.Up_orders[button_press.Floor] = true
	}
}
