package singleelevator

import (
	"elevatorproject/singleelevator/elevio"
	"elevatorproject/config"
)

type Orders struct {
	Up_orders   []bool
	Down_orders []bool
	Cab_orders  []bool
}

func (orders *Orders) InitOrders() {
	orders.Cab_orders = make([]bool, config.NumFloors)
	orders.Up_orders = make([]bool, config.NumFloors)
	orders.Down_orders = make([]bool, config.NumFloors)
	for i := 0; i < config.NumFloors; i++ {
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
	for i := 0; i < config.NumFloors; i++ {
		if orders.Cab_orders[i] || orders.Up_orders[i] || orders.Down_orders[i] {
			return true
		}
	}
	return false
}

func (orders Orders) AnyOrderPastFloorInDir(floor int, dir elevio.MotorDirection) bool {
	switch dir {
	case elevio.MD_Up:
		if floor == config.NumFloors-1 {
			return false
		}
		for i := floor + 1; i < config.NumFloors; i++ {
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

func (orders Orders) GetCabRequests() [config.NumFloors]bool {
	var cab_requests [config.NumFloors]bool
	for i := 0; i < config.NumFloors; i++ {
		cab_requests[i] = orders.Cab_orders[i]
	}
	return cab_requests
}

func (orders Orders) GetHallRequests() [config.NumFloors][2]bool {
	var hall_requests [config.NumFloors][2]bool
	for i := 0; i < config.NumFloors; i++ {
		hall_requests[i][0] = orders.Up_orders[i]
		hall_requests[i][1] = orders.Down_orders[i]
	}
	return hall_requests
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

