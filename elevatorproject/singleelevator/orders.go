package singleelevator

import (
	"elevatorproject/config"
	"elevatorproject/singleelevator/elevio"
)

type activeOrders struct {
	upOrders   []bool
	downOrders []bool
	cabOrders  []bool
}

func (orders *activeOrders) initOrders() {
	orders.cabOrders = make([]bool, config.NumFloors)
	orders.upOrders = make([]bool, config.NumFloors)
	orders.downOrders = make([]bool, config.NumFloors)
	for i := 0; i < config.NumFloors; i++ {
		orders.cabOrders[i] = false
		orders.upOrders[i] = false
		orders.downOrders[i] = false
	}
}

func (orders activeOrders) getSpecificOrder(floor int, btn elevio.ButtonType) bool {
	switch btn {
	case elevio.BT_HallUp:
		return orders.upOrders[floor]
	case elevio.BT_HallDown:
		return orders.downOrders[floor]
	case elevio.BT_Cab:
		return orders.cabOrders[floor]
	}
	return false
}

func (orders activeOrders) anyOrder() bool {
	for i := 0; i < config.NumFloors; i++ {
		if orders.cabOrders[i] || orders.upOrders[i] || orders.downOrders[i] {
			return true
		}
	}
	return false
}

func (orders activeOrders) anyOrderPastFloorInDir(floor int, dir elevio.MotorDirection) bool {
	switch dir {
	case elevio.MD_Up:
		if floor == config.NumFloors-1 {
			return false
		}
		for i := floor + 1; i < config.NumFloors; i++ {
			if orders.cabOrders[i] || orders.downOrders[i] || orders.upOrders[i] {
				return true
			}
		}
	case elevio.MD_Down:
		if floor == 0 {
			return false
		}
		for i := floor - 1; i > -1; i-- {
			if orders.cabOrders[i] || orders.downOrders[i] || orders.upOrders[i] {
				return true
			}
		}
	}
	return false
}

func (orders activeOrders) getCabRequests() [config.NumFloors]bool {
	var cab_requests [config.NumFloors]bool
	for i := 0; i < config.NumFloors; i++ {
		cab_requests[i] = orders.cabOrders[i]
	}
	return cab_requests
}

func (orders activeOrders) getHallRequests() [config.NumFloors][2]bool {
	var hall_requests [config.NumFloors][2]bool
	for i := 0; i < config.NumFloors; i++ {
		hall_requests[i][0] = orders.upOrders[i]
		hall_requests[i][1] = orders.downOrders[i]
	}
	return hall_requests
}

func (orders *activeOrders) setOrder(floor int, btn elevio.ButtonType, value bool) {
	switch btn {
	case elevio.BT_HallUp:
		orders.upOrders[floor] = value
	case elevio.BT_HallDown:
		orders.downOrders[floor] = value
	case elevio.BT_Cab:
		orders.cabOrders[floor] = value
	}
}

func (orders activeOrders) orderInFloor(floor int) bool {
	if orders.upOrders[floor] || orders.downOrders[floor] || orders.cabOrders[floor] {
		return true
	}
	return false
}

func (orders activeOrders) getOrdersInFloor(floor int) (bool, bool, bool) {
	return orders.upOrders[floor], orders.downOrders[floor], orders.cabOrders[floor]
}
