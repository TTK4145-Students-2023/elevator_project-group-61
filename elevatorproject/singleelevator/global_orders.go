package singleelevator

import "elevatorproject/singleelevator/elevio"

// What this module needs to do:
// - For all elevators know what hall orders they have and in what stage those orders are.
// - - The orders can be in "unassigned", "assigned", "completed" or "aborted" state.

const n_elevators int = 3

type OrderStage int

const (
	Unassigned OrderStage = 0
	Assigned   OrderStage = 1
	Completed  OrderStage = 2
	Aborted    OrderStage = 3
)

type GlobalOrders struct {
	Up_orders   []OrderStage
	Down_orders []OrderStage
}

func (globalorders *GlobalOrders) InitGlobalOrders() {
	globalorders.Up_orders = make([]OrderStage, NumFloors)
	globalorders.Down_orders = make([]OrderStage, NumFloors)
	for i := 0; i < NumFloors; i++ {
		globalorders.Up_orders[i] = 0
		globalorders.Down_orders[i] = 0
	}
}

type ElevatorOrdersMap map[int]GlobalOrders

func (elev_order_map *ElevatorOrdersMap) InitGlobalOrdersMap() {
	for i := 0; i < n_elevators; i++ {
		(*elev_order_map)[i] = GlobalOrders{Up_orders: make([]OrderStage, NumFloors), Down_orders: make([]OrderStage, NumFloors)}
		for j := 0; j < NumFloors; j++ {
			(*elev_order_map)[i].Up_orders[j] = 0
			(*elev_order_map)[i].Down_orders[j] = 0
		}
	}
}

func (elev_order_map ElevatorOrdersMap) CheckIfOrderCanBeAssigned(floor int, btn_type elevio.ButtonType) bool {
	switch btn_type {
	case elevio.BT_HallUp:
		for i := 0; i < n_elevators; i++ {
			if elev_order_map[i].Up_orders[floor] == Assigned || elev_order_map[i].Up_orders[floor] == Completed {
				return false
			}
		}
	case elevio.BT_HallDown:
		for i := 0; i < n_elevators; i++ {
			if elev_order_map[i].Down_orders[floor] == Assigned || elev_order_map[i].Down_orders[floor] == Completed {
				return false
			}
		}
	}
	return true
}

func (elev_order_map *ElevatorOrdersMap) SetOrderAssigned(id int, floor int, btn_type elevio.ButtonType) {
	if btn_type == elevio.BT_HallUp {
		(*elev_order_map)[id].Up_orders[floor] = Assigned
	}
	if btn_type == elevio.BT_HallDown {
		(*elev_order_map)[id].Down_orders[floor] = Assigned
	}
}

func (elev_order_map *ElevatorOrdersMap) SetOrderCompleted(id int, floor int, btn_type elevio.ButtonType) {
	if btn_type == elevio.BT_HallUp {
		(*elev_order_map)[id].Up_orders[floor] = Completed
	}
	if btn_type == elevio.BT_HallDown {
		(*elev_order_map)[id].Down_orders[floor] = Completed
	}
}

func (elev_order_map *ElevatorOrdersMap) UpdateElevatorFromNetworkSignal(id int, global_orders GlobalOrders) {
	(*elev_order_map)[id] = global_orders
}

// I don't know if these below here are useful
func (elev_order_map ElevatorOrdersMap) GetSpecificElevatorAssignedOrders(id int) Orders {
	assigned_orders := Orders{Up_orders: make([]bool, NumFloors), Down_orders: make([]bool, NumFloors), Cab_orders: make([]bool, NumFloors)}
	for i := 0; i < NumFloors; i++ {
		assigned_orders.Cab_orders[i] = false
		if elev_order_map[id].Up_orders[i] == Assigned {
			assigned_orders.Up_orders[i] = true
		}
		if elev_order_map[id].Down_orders[i] == Assigned {
			assigned_orders.Down_orders[i] = true
		}
	}
	return assigned_orders
}

func (old_global_orders GlobalOrders) GetDifferenceBetweenElevatorsOrders(new_global_orders GlobalOrders) GlobalOrders {
	changed_orders := GlobalOrders{Up_orders: make([]OrderStage, NumFloors), Down_orders: make([]OrderStage, NumFloors)}
	for i := 0; i < NumFloors; i++ {
		if old_global_orders.Up_orders[i] != new_global_orders.Up_orders[i] {
			changed_orders.Up_orders[i] = new_global_orders.Up_orders[i]
		}
		if old_global_orders.Down_orders[i] != new_global_orders.Down_orders[i] {
			changed_orders.Down_orders[i] = new_global_orders.Down_orders[i]
		}
	}
	return changed_orders
}

func (elev_order_map *ElevatorOrdersMap) UpdateGlobalMapFromLocalChange(id int, btn elevio.ButtonEvent, stage OrderStage) {
	if btn.Button == elevio.BT_HallUp {
		(*elev_order_map)[id].Up_orders[btn.Floor] = stage
	} else if btn.Button == elevio.BT_HallDown {
		(*elev_order_map)[id].Down_orders[btn.Floor] = stage
	}
}
