package main

import (
	"ElevatorProject/elevio"
	"encoding/json"
	"io/ioutil"
)

var elevator_orders_filename = "elevator_orders_backup.json"

type Orders struct {
	Up_orders   []bool
	Down_orders []bool
	Cab_orders  []bool
}

// Save and load to and from file functionality
func SaveElevatorOrdersToFile(orders Orders) error {
    data, err := json.MarshalIndent(orders, "", "    ")
    if err != nil {
        return err
    }
    err = ioutil.WriteFile(elevator_orders_filename, data, 0644)
    if err != nil {
        return err
    }
    return nil
}

func LoadElevatorOrdersFromFile() (Orders, error) {
    var orders Orders
    data, err := ioutil.ReadFile(elevator_orders_filename)
    if err != nil {
        return orders, err
    }
    err = json.Unmarshal(data, &orders)
    if err != nil {
        return orders, err
    }
    return orders, nil
}

// Methods
func (orders *Orders) InitOrders() {
	orders.Cab_orders = make([]bool, n_floors)
	orders.Up_orders = make([]bool, n_floors)
	orders.Down_orders = make([]bool, n_floors)
	// Try to load old orders from file
	some_orders, err := LoadElevatorOrdersFromFile()
	if err == nil {
		orders = &some_orders
		return
	}
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
	SaveElevatorOrdersToFile(*orders)
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
	SaveElevatorOrdersToFile(*orders)
}

func (orders Orders) OrderInFloor(floor int) bool {
	if orders.Up_orders[floor] || orders.Down_orders[floor] || orders.Cab_orders[floor] {
		return true
	}
	return false
}


