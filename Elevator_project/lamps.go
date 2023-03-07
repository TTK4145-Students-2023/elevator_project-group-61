package main

import (
	"ElevatorProject/elevio"
)

func LampStateMachine(ch_hall_requests chan [][2]bool, ch_cab_requests chan []bool) {
	//TODO: make
	
}


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

func UpdateSingleElevOrderLamps(orders Orders) {
	for floor_num := 0; floor_num < n_floors; floor_num++ {
		for i := 0; i < 3; i++ {
			if orders.GetSpecificOrder(floor_num, elevio.ButtonType(i)) {
				elevio.SetButtonLamp(elevio.ButtonType(i), floor_num, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(i), floor_num, false)
			}
		}
	}
}

func UpdateCabLamps(orders Orders) {
	for floor_num := 0; floor_num < n_floors; floor_num++ {
		if orders.GetSpecificOrder(floor_num, elevio.BT_Cab) {
			elevio.SetButtonLamp(elevio.BT_Cab, floor_num, true)
		} else {
			elevio.SetButtonLamp(elevio.BT_Cab, floor_num, false)
		}
	}
}