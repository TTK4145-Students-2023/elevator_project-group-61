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
	lastFloor         int
	lastDirection     elevio.MotorDirection
	elevatorBehaviour string
}

func (state *localElevState) initLocalElevState() {
	state.lastFloor = -1
	state.lastDirection = elevio.MD_Up
	state.elevatorBehaviour = Idle
}

func (state *localElevState) setLastFloor(floor int) {
	state.lastFloor = floor
}

func (state *localElevState) setDirection(dir elevio.MotorDirection) {
	switch dir {
	case elevio.MD_Up:
		state.lastDirection = elevio.MD_Up
		state.elevatorBehaviour = Moving
	case elevio.MD_Down:
		state.lastDirection = elevio.MD_Down
		state.elevatorBehaviour = Moving
	case elevio.MD_Stop:
	}
}

func (state *localElevState) setElevatorBehaviour(behaviour string) {
	state.elevatorBehaviour = behaviour
}

func (state localElevState) getLastFloor() int {
	return state.lastFloor
}

func (state localElevState) getLastDirection() elevio.MotorDirection {
	return state.lastDirection
}

func (state localElevState) getElevatorBehaviour() string {
	return state.elevatorBehaviour
}
