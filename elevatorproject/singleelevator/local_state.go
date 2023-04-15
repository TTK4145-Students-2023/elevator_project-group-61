package singleelevator

import (
	"elevatorproject/singleelevator/elevio"
)

const (
	Idle     = "Idle"
	Moving   = "Moving"
	DoorOpen = "DoorOpen"
)

type localElevState struct {
	LastFloor         int
	LastDirection     elevio.MotorDirection
	ElevatorBehaviour string
}

func (state *localElevState) InitLocalElevState() {
	state.LastFloor = -1
	state.LastDirection = elevio.MD_Up
	state.ElevatorBehaviour = Idle
}

func (state *localElevState) SetLastFloor(floor int) {
	state.LastFloor = floor
}

func (state *localElevState) SetDirection(dir elevio.MotorDirection) {
	switch dir {
	case elevio.MD_Up:
		state.LastDirection = elevio.MD_Up
		state.ElevatorBehaviour = Moving
	case elevio.MD_Down:
		state.LastDirection = elevio.MD_Down
		state.ElevatorBehaviour = Moving
	case elevio.MD_Stop:
	}
}

func (state *localElevState) SetElevatorBehaviour(behaviour string) {
	state.ElevatorBehaviour = behaviour
}

func (state localElevState) GetLastFloor() int {
	return state.LastFloor
}

func (state localElevState) GetLastDirection() elevio.MotorDirection {
	return state.LastDirection
}

func (state localElevState) GetElevatorBehaviour() string {
	return state.ElevatorBehaviour
}
