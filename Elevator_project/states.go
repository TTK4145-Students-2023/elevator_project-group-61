package main 

import (
	"ElevatorProject/elevio"
)

type States struct {
	last_floor     int
	last_direction elevio.MotorDirection
	door_open      bool
	moving         bool
}

// Methods
func (states *States) InitStates() {
	states.last_floor = -1
	states.last_direction = elevio.MD_Up
	states.door_open = false
	states.moving = false
}

func (states *States) SetLastFloor(floor int) {
	states.last_floor = floor
	elevio.SetFloorIndicator(floor)
	DelegateStates(Elev_states)
}

func (states *States) SetDirection(dir elevio.MotorDirection) {
	switch dir {
	case elevio.MD_Up:
		states.last_direction = elevio.MD_Up
		states.moving = true
		elevio.SetMotorDirection(elevio.MD_Up)
	case elevio.MD_Down:
		states.last_direction = elevio.MD_Down
		states.moving = true
		elevio.SetMotorDirection(elevio.MD_Down)
	case elevio.MD_Stop:
		states.moving = false
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
	DelegateStates(Elev_states)
}

func (states *States) SetDoorOpen(open bool) {
	states.door_open = open
	elevio.SetDoorOpenLamp(open)
	DelegateStates(Elev_states)
}

func (states States) GetLastFloor() int {
	return states.last_floor
}

func (states States) GetLastDirection() elevio.MotorDirection {
	return states.last_direction
}

func (states States) GetDoorOpen() bool {
	return states.door_open
}

func (states States) GetMoving() bool {
	return states.moving
}