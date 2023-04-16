package lamps

import (
	"elevatorproject/singleelevator/elevio"
	"elevatorproject/config"
)

func LampStateMachine(
	ch_hallRequests <-chan [config.NumFloors][2]bool, 
	ch_cabRequests  <-chan [config.NumFloors]bool,
) {
	elevio.SetDoorOpenLamp(false)
	for floor_num := 0; floor_num < config.NumFloors; floor_num++ {
		for i := 0; i < 3; i++ {
			elevio.SetButtonLamp(elevio.ButtonType(i), floor_num, false)
		}
	}

	for {
		select {
		case hall_requests := <-ch_hallRequests:
			for floor_num := 0; floor_num < config.NumFloors; floor_num++ {
				for i := 0; i < 2; i++ {
					if hall_requests[floor_num][i] {
						elevio.SetButtonLamp(elevio.ButtonType(i), floor_num, true)
					} else {
						elevio.SetButtonLamp(elevio.ButtonType(i), floor_num, false)
					}
				}
			}
		case cab_requests := <-ch_cabRequests:
			for floor_num := 0; floor_num < config.NumFloors; floor_num++ {
				if cab_requests[floor_num] {
					elevio.SetButtonLamp(elevio.BT_Cab, floor_num, true)
				} else {
					elevio.SetButtonLamp(elevio.BT_Cab, floor_num, false)
				}
			}
		}
	}
}
