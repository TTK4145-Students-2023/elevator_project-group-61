package main

import (
	"ElevatorProject/elevio"
)

func InitLamps() {
	elevio.SetDoorOpenLamp(false)
	for i := 0; i < n_floors; i++ {
		for j := 0; j < 3; j++ {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, false)
		}
	}
}