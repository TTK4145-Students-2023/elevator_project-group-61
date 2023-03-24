package singleelevator

import (
	"elevatorproject/singleelevator/elevio"
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

func (orders Orders) GetCabRequests() []bool {
	return orders.Cab_orders
}

func (orders *Orders) SetOrder(floor int, btn elevio.ButtonType, value bool) {
	switch btn {
	case elevio.BT_HallUp:
		orders.Up_orders[floor] = value
	case elevio.BT_HallDown:
		orders.Down_orders[floor] = value
	case elevio.BT_Cab:
		orders.Cab_orders[floor] = value
	}
}

func (orders Orders) OrderInFloor(floor int) bool {
	if orders.Up_orders[floor] || orders.Down_orders[floor] || orders.Cab_orders[floor] {
		return true
	}
	return false
}

func (orders Orders) GetOrdersInFloor(floor int) (bool, bool, bool) {
	return orders.Up_orders[floor], orders.Down_orders[floor], orders.Cab_orders[floor]
}