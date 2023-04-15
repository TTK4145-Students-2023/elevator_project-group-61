package singleelevator

import (
	"elevatorproject/singleelevator/elevio"
)

const (
	Idle = "Idle"
	Moving = "Moving"
	DoorOpen = "DoorOpen"
)

type States struct {
	Last_floor     int
	Last_direction elevio.MotorDirection
	ElevatorBehaviour string
}

// Methods
func (states *States) InitStates() {
	states.Last_floor = -1
	states.Last_direction = elevio.MD_Up
	states.ElevatorBehaviour = Idle
}

func (states *States) SetLastFloor(floor int) {
	states.Last_floor = floor
}

func (states *States) SetDirection(dir elevio.MotorDirection) {
	switch dir {
	case elevio.MD_Up:
		states.Last_direction = elevio.MD_Up
		states.ElevatorBehaviour = Moving
	case elevio.MD_Down:
		states.Last_direction = elevio.MD_Down
		states.ElevatorBehaviour = Moving
	case elevio.MD_Stop:
	}
}

func (states *States) SetElevatorBehaviour(behaviour string) {
	states.ElevatorBehaviour = behaviour
}

func (states States) GetLastFloor() int {
	return states.Last_floor
}

func (states States) GetLastDirection() elevio.MotorDirection {
	return states.Last_direction
}

func (states States) GetElevatorBehaviour() string {
	return states.ElevatorBehaviour
}