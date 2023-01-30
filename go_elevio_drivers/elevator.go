package main

import (
	"fmt"
)

type ElevatorBehaviour int

const (
	Idle = iota,
	DoorOpen,
	Moving,
)

type Elevator struct {
	floor int
	direction Direction
	//behaviour
	ElevatorBehaviour behaviour
}
