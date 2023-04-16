package singleelevator

import (
	"elevatorproject/config"
	"elevatorproject/singleelevator/elevio"
)

type activeRequests struct {
	upOrders   []bool
	downOrders []bool
	cabOrders  []bool
}

func (requests *activeRequests) initRequests() {
	requests.cabOrders = make([]bool, config.NumFloors)
	requests.upOrders = make([]bool, config.NumFloors)
	requests.downOrders = make([]bool, config.NumFloors)
	for i := 0; i < config.NumFloors; i++ {
		requests.cabOrders[i] = false
		requests.upOrders[i] = false
		requests.downOrders[i] = false
	}
}

func (requests activeRequests) getSpecificRequest(floor int, btn elevio.ButtonType) bool {
	switch btn {
	case elevio.BT_HallUp:
		return requests.upOrders[floor]
	case elevio.BT_HallDown:
		return requests.downOrders[floor]
	case elevio.BT_Cab:
		return requests.cabOrders[floor]
	}
	return false
}

func (requests activeRequests) anyRequest() bool {
	for i := 0; i < config.NumFloors; i++ {
		if requests.cabOrders[i] || requests.upOrders[i] || requests.downOrders[i] {
			return true
		}
	}
	return false
}

func (requests activeRequests) anyRequestPastFloorInDir(floor int, dir elevio.MotorDirection) bool {
	switch dir {
	case elevio.MD_Up:
		if floor == config.NumFloors-1 {
			return false
		}
		for i := floor + 1; i < config.NumFloors; i++ {
			if requests.cabOrders[i] || requests.downOrders[i] || requests.upOrders[i] {
				return true
			}
		}
	case elevio.MD_Down:
		if floor == 0 {
			return false
		}
		for i := floor - 1; i > -1; i-- {
			if requests.cabOrders[i] || requests.downOrders[i] || requests.upOrders[i] {
				return true
			}
		}
	}
	return false
}

func (requests activeRequests) getCabRequests() [config.NumFloors]bool {
	var cab_requests [config.NumFloors]bool
	for i := 0; i < config.NumFloors; i++ {
		cab_requests[i] = requests.cabOrders[i]
	}
	return cab_requests
}

func (requests activeRequests) getHallRequests() [config.NumFloors][2]bool {
	var hall_requests [config.NumFloors][2]bool
	for i := 0; i < config.NumFloors; i++ {
		hall_requests[i][0] = requests.upOrders[i]
		hall_requests[i][1] = requests.downOrders[i]
	}
	return hall_requests
}

func (requests *activeRequests) setRequest(floor int, btn elevio.ButtonType, value bool) {
	switch btn {
	case elevio.BT_HallUp:
		requests.upOrders[floor] = value
	case elevio.BT_HallDown:
		requests.downOrders[floor] = value
	case elevio.BT_Cab:
		requests.cabOrders[floor] = value
	}
}

func (requests activeRequests) requestInFloor(floor int) bool {
	if requests.upOrders[floor] || requests.downOrders[floor] || requests.cabOrders[floor] {
		return true
	}
	return false
}

func (requests activeRequests) getRequestsInFloor(floor int) (bool, bool, bool) {
	return requests.upOrders[floor], requests.downOrders[floor], requests.cabOrders[floor]
}
