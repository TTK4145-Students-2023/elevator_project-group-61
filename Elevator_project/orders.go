package main

import (
	"ElevatorProject/elevio"
	"math"
)

type Orders struct {
	Up_orders   []bool
	Down_orders []bool
	Cab_orders  []bool
}

// Methods
func (orders *Orders) InitOrders() {
	orders.Cab_orders = make([]bool, n_floors)
	orders.Up_orders = make([]bool, n_floors)
	orders.Down_orders = make([]bool, n_floors)
	for i := 0; i < n_floors; i++ {
		orders.Cab_orders[i] = false
		orders.Up_orders[i] = false
		orders.Down_orders[i] = false
	}
}

func (orders Orders) GetSpecificOrder(floor int, btn elevio.ButtonType) bool {
	switch btn {
	case elevio.BT_HallUp:
		return orders.Up_orders[floor]
	case elevio.BT_HallDown:
		return orders.Down_orders[floor]
	case elevio.BT_Cab:
		return orders.Cab_orders[floor]
	}
	return false
}

func (orders Orders) AnyOrder() bool {
	for i := 0; i < n_floors; i++ {
		if orders.Cab_orders[i] || orders.Up_orders[i] || orders.Down_orders[i] {
			return true
		}
	}
	return false
}

func (orders Orders) AnyOrderPastFloorInDir(floor int, dir elevio.MotorDirection) bool {
	switch dir {
	case elevio.MD_Up:
		if floor == n_floors-1 {
			return false
		}
		for i := floor + 1; i < n_floors; i++ {
			if orders.Cab_orders[i] || orders.Down_orders[i] || orders.Up_orders[i] {
				return true
			}
		}
	case elevio.MD_Down:
		if floor == 0 {
			return false
		}
		for i := floor - 1; i > -1; i-- {
			if orders.Cab_orders[i] || orders.Down_orders[i] || orders.Up_orders[i] {
				return true
			}
		}
	}
	return false
}

func (orders *Orders) AddOrder(btn elevio.ButtonEvent) {
	elevio.SetButtonLamp(btn.Button, btn.Floor, true)
	switch btn.Button {
	case elevio.BT_HallUp:
		orders.Up_orders[btn.Floor] = true
	case elevio.BT_HallDown:
		orders.Down_orders[btn.Floor] = true
	case elevio.BT_Cab:
		orders.Cab_orders[btn.Floor] = true
	}
}

func (orders Orders) FindClosestOrder(floor int) int {
	closest_order := -1
	shortest_diff := n_floors
	for i := 0; i < n_floors; i++ {
		if orders.Cab_orders[i] || orders.Down_orders[i] || orders.Up_orders[i] {
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

func (orders *Orders) RemoveOrderDirection(floor int, dir elevio.MotorDirection) {
	if dir == elevio.MD_Up && Active_orders.Up_orders[floor] {
		Active_orders.Up_orders[floor] = false
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	}
	if dir == elevio.MD_Down && Active_orders.Down_orders[floor] {
		Active_orders.Down_orders[floor] = false
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	}
	if dir == elevio.MD_Stop && Active_orders.Cab_orders[floor] {
		Active_orders.Cab_orders[floor] = false
		elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	}
}

func (orders Orders) OrderInFloor(floor int) bool {
	if orders.Up_orders[floor] || orders.Down_orders[floor] || orders.Cab_orders[floor] {
		return true
	}
	return false
}

