package main

import (
	"ElevatorProject/elevio"
)

func InitLamps(active_orders Orders) {
	elevio.SetDoorOpenLamp(false)
	for floor_num := 0; floor_num < n_floors; floor_num++ {
		for i := 0; i < 3; i++ {
			if active_orders.GetSpecificOrder(floor_num, elevio.ButtonType(i)) {
				elevio.SetButtonLamp(elevio.ButtonType(i), floor_num, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(i), floor_num, false)
			}
		}
	}
}

func UpdateLamps() {
	// Should update lamps from global orders
	// TODO: Implement later
}