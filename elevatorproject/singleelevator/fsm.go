package singleelevator

import (
	"elevatorproject/config"
	"elevatorproject/singleelevator/elevio"
	"time"
)

const nFloors int = config.NumFloors
const standardDoorWait = 3 * time.Second
const initiateTime = 9999 * time.Second

// Struct to distribute elevator state to other modules
type ElevState struct {
	Behaviour   string
	Floor       int
	Direction   string
	IsAvailable bool
}

func (elevState *ElevState) InitElevState() {
	elevState.Behaviour = "idle"
	elevState.Floor = 1
	elevState.Direction = "stop"
	elevState.IsAvailable = true
}

func stopAfterSensingFloor(floor int, state localElevState, requests activeRequests) bool {
	if floor == 0 || floor == nFloors-1 {
		return true
	}
	if !requests.anyRequest() {
		return true
	}
	if requests.getSpecificRequest(floor, elevio.BT_Cab) {
		return true
	}
	switch state.getLastDirection() {
	case elevio.MD_Up:
		if requests.getSpecificRequest(floor, elevio.BT_HallUp) ||
			!requests.anyRequestPastFloorInDir(floor, elevio.MD_Up) {
			return true
		}
		return false
	case elevio.MD_Down:
		if requests.getSpecificRequest(floor, elevio.BT_HallDown) ||
			!requests.anyRequestPastFloorInDir(floor, elevio.MD_Down) {
			return true
		}
		return false
	}
	return false
}

func dirToBtnType(dir elevio.MotorDirection) elevio.ButtonType {
	switch dir {
	case elevio.MD_Up:
		return elevio.BT_HallUp
	case elevio.MD_Down:
		return elevio.BT_HallDown
	}
	return elevio.BT_Cab
}

func localStateToElevState(state localElevState, isAvailable bool) ElevState {
	elevState := ElevState{
		Behaviour:   "idle",
		Floor:       1,
		Direction:   "stop",
		IsAvailable: true}

	if state.getLastFloor() == -1 {
		elevState.Floor = nFloors / 2
	} else {
		elevState.Floor = state.getLastFloor()
	}

	switch state.getLastDirection() {
	case elevio.MD_Up:
		elevState.Direction = "up"
	case elevio.MD_Down:
		elevState.Direction = "down"
	}

	switch state.getElevatorBehaviour() {
	case Idle:
		elevState.Behaviour = "idle"
		elevState.Direction = "stop"
	case Moving:
		elevState.Behaviour = "moving"
	case DoorOpen:
		elevState.Behaviour = "doorOpen"
		elevState.Direction = "stop"
	default:
		panic("Invalid elevator behaviour")
	}

	elevState.IsAvailable = isAvailable

	return elevState
}

func hasElevStateChanged(a ElevState, b ElevState) bool {
	if a.Behaviour != b.Behaviour ||
		a.Floor != b.Floor ||
		a.Direction != b.Direction ||
		a.IsAvailable != b.IsAvailable {
		return true
	}
	return false
}

// Functions for handling events
func handleFloorSensor(
	floor int,
	state localElevState,
	requests activeRequests,
) (
	localElevState,
	activeRequests,
	[]elevio.ButtonEvent,
) {
	state.setLastFloor(floor)
	completedRequestsList := make([]elevio.ButtonEvent, 0)

	if stopAfterSensingFloor(floor, state, requests) {
		state.setElevatorBehaviour(DoorOpen)
		if requests.getSpecificRequest(floor, elevio.BT_Cab) {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab})
		}
		if requests.getSpecificRequest(floor, dirToBtnType(state.getLastDirection())) {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: floor, Button: dirToBtnType(state.getLastDirection())})
		}
		if floor == 0 && requests.getSpecificRequest(floor, elevio.BT_HallUp) {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp})
		}
		if floor == nFloors-1 && requests.getSpecificRequest(floor, elevio.BT_HallDown) {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown})
		}
	}
	return state, requests, completedRequestsList
}

func handleNewRequests(
	hallRequests [nFloors][2]bool,
	cabRequests [nFloors]bool,
	cabOrHall string,
	state localElevState,
	requests activeRequests,
) (
	localElevState,
	activeRequests,
	[]elevio.ButtonEvent,
) {
	if cabOrHall == "cab" {
		for i := 0; i < nFloors; i++ {
			requests.setRequest(i, elevio.BT_Cab, cabRequests[i])
		}
	} else if cabOrHall == "hall" {
		for i := 0; i < nFloors; i++ {
			requests.setRequest(i, elevio.BT_HallUp, hallRequests[i][0])
			requests.setRequest(i, elevio.BT_HallDown, hallRequests[i][1])
		}
	} else {
		panic("handleNewRequests: hallOrCab must be 'cab' or 'hall'")
	}

	completedRequestsList := make([]elevio.ButtonEvent, 0)

	switch state.getElevatorBehaviour() {
	case Moving:
	case DoorOpen:
		upThisFloor, downThisFloor, cabThisFloor := requests.getRequestsInFloor(state.getLastFloor())
		if cabThisFloor {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_Cab})
		}
		if (upThisFloor && state.getLastDirection() == elevio.MD_Up) ||
			(state.getLastFloor() == 0 && upThisFloor) {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallUp})
		}
		if (downThisFloor && state.getLastDirection() == elevio.MD_Down) ||
			(state.getLastFloor() == nFloors-1 && downThisFloor) {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallDown})
		}
	case Idle:
		upThisFloor, downThisFloor, cabThisFloor := requests.getRequestsInFloor(state.getLastFloor())
		orderInThisFloor := upThisFloor || downThisFloor || cabThisFloor
		orders_above := requests.anyRequestPastFloorInDir(state.getLastFloor(), elevio.MD_Up)
		orders_below := requests.anyRequestPastFloorInDir(state.getLastFloor(), elevio.MD_Down)
		if !requests.anyRequest() {
			break
		}
		if orderInThisFloor {
			if cabThisFloor {
				state.setElevatorBehaviour(DoorOpen)
				completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_Cab})
			}
			if (upThisFloor && state.getLastDirection() == elevio.MD_Up) ||
				(upThisFloor && !downThisFloor && !requests.anyRequestPastFloorInDir(state.getLastFloor(), elevio.MD_Down)) ||
				(state.getLastFloor() == 0 && upThisFloor) {
				state.setElevatorBehaviour(DoorOpen)
				completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallUp})
			}
			if (downThisFloor && state.getLastDirection() == elevio.MD_Down) ||
				(downThisFloor && !upThisFloor && !requests.anyRequestPastFloorInDir(state.getLastFloor(), elevio.MD_Up)) ||
				(state.getLastFloor() == nFloors-1 && downThisFloor) {
				state.setElevatorBehaviour(DoorOpen)
				completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallDown})
			}

		}
		if !orderInThisFloor || state.getElevatorBehaviour() != DoorOpen {
			state.setElevatorBehaviour(Moving)
			if (state.getLastDirection() == elevio.MD_Up && orders_above) || !orders_below {
				state.setDirection(elevio.MD_Up)
			} else if (state.getLastDirection() == elevio.MD_Down && orders_below) || !orders_above {
				state.setDirection(elevio.MD_Down)
			} else {
				panic("handleNewRequests: No direction set")
			}
		}
	}
	return state, requests, completedRequestsList
}

func handleDoorClosing(state localElevState, requests activeRequests) (localElevState, activeRequests, []elevio.ButtonEvent) {
	completedRequestsList := make([]elevio.ButtonEvent, 0)

	if !requests.anyRequest() {
		state.setElevatorBehaviour(Idle)
		return state, requests, completedRequestsList
	}
	if requests.getSpecificRequest(state.getLastFloor(), dirToBtnType(state.getLastDirection())) {
		if state.getLastDirection() == elevio.MD_Up {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallUp})
		} else {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallDown})
		}
		return state, requests, completedRequestsList
	} else if requests.requestInFloor(state.getLastFloor()) && !requests.anyRequestPastFloorInDir(state.getLastFloor(), state.getLastDirection()) {
		if state.getLastDirection() == elevio.MD_Up {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallDown})
		} else {
			completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallUp})
		}
		return state, requests, completedRequestsList
	}
	state.setElevatorBehaviour(Moving)
	ordersUpBool := requests.anyRequestPastFloorInDir(state.getLastFloor(), elevio.MD_Up)
	ordersDownBool := requests.anyRequestPastFloorInDir(state.getLastFloor(), elevio.MD_Down)
	if (state.getLastDirection() == elevio.MD_Up && ordersUpBool) || !ordersDownBool {
		state.setDirection(elevio.MD_Up)
	} else if (state.getLastDirection() == elevio.MD_Down && ordersDownBool) || !ordersUpBool {
		state.setDirection(elevio.MD_Down)
	} else {
		panic("handleDoorClosing: No direction set")
	}
	return state, requests, completedRequestsList
}

func hasLocalElevStateChanged(a localElevState, b localElevState) bool {
	if a.getLastFloor() != b.getLastFloor() {
		return true
	}
	if a.getLastDirection() != b.getLastDirection() {
		return true
	}
	if a.getElevatorBehaviour() != b.getElevatorBehaviour() {
		return true
	}
	return false
}

func fsmElevator(
	ch_btn <-chan elevio.ButtonEvent,
	ch_floor <-chan int,
	ch_hallRequests <-chan [nFloors][2]bool,
	ch_cabRequests <-chan [nFloors]bool,
	ch_singleElevMode <-chan bool,
	ch_completedRequest chan<- elevio.ButtonEvent,
	ch_newRequest chan<- elevio.ButtonEvent,
	ch_elevState chan<- ElevState,
) {
	// Elevator variables
	var myLocalState localElevState
	var myActiveRequests activeRequests
	isAvailable := false
	singleElevMode := true
	isObstructed := false

	// For error handling
	errorTimer := time.Now().UnixMilli()
	var oldLocalState localElevState
	oldLocalState.initLocalElevState()

	// Door timer
	doorTimer := time.NewTimer(initiateTime)

	// Initiate elevator
	myActiveRequests.initRequests()
	myLocalState.initLocalElevState()
	var oldElevState ElevState = localStateToElevState(myLocalState, isAvailable)
	if elevio.GetFloor() == -1 {
		myLocalState.setElevatorBehaviour(Moving)
		elevio.SetMotorDirection(elevio.MD_Up)
	} else {
		myLocalState.setLastFloor(elevio.GetFloor())
		elevio.SetFloorIndicator(elevio.GetFloor())
		isAvailable = true
	}
	if hasElevStateChanged(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
		oldElevState = localStateToElevState(myLocalState, isAvailable)
		ch_elevState <- oldElevState
	}

	// Main loop
	for {
		select {
		case btnPress := <-ch_btn:
			ch_newRequest <- elevio.ButtonEvent{Floor: btnPress.Floor, Button: btnPress.Button}
			if singleElevMode {
				myActiveRequests.setRequest(btnPress.Floor, btnPress.Button, true)

				cabReqList := [nFloors]bool{false, false, false, false}
				hallReqList := [nFloors][2]bool{
					{false, false},
					{false, false},
					{false, false},
					{false, false},
				}
				cabOrHall := "cab"

				if btnPress.Button != elevio.BT_Cab {
					cabOrHall = "hall"
					hallReqList = myActiveRequests.getHallRequests()
				} else {
					cabReqList = myActiveRequests.getCabRequests()
				}
				var completedRequestsList []elevio.ButtonEvent
				myLocalState, myActiveRequests, completedRequestsList = handleNewRequests(hallReqList,
																							cabReqList,
																							cabOrHall,
																							myLocalState,
																							myActiveRequests,
																						)
				if myLocalState.getElevatorBehaviour() == DoorOpen {
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(standardDoorWait)
					for _, request := range completedRequestsList {
						myActiveRequests.setRequest(request.Floor, request.Button, false)
						ch_completedRequest <- request
					}
				} else if myLocalState.getElevatorBehaviour() == Moving && oldElevState.Direction == "stop" {
					elevio.SetMotorDirection(myLocalState.getLastDirection())
				}

				// Distribute state
				if hasElevStateChanged(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
					oldElevState = localStateToElevState(myLocalState, isAvailable)
					ch_elevState <- oldElevState
				}
			}

		case floor := <-ch_floor:
			elevio.SetFloorIndicator(floor)
			var completedRequestsList []elevio.ButtonEvent
			myLocalState, myActiveRequests, completedRequestsList = handleFloorSensor(floor, myLocalState, myActiveRequests)
			if myLocalState.getElevatorBehaviour() == DoorOpen {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				doorTimer.Reset(standardDoorWait)
				for _, request := range completedRequestsList {
					if singleElevMode {
						myActiveRequests.setRequest(request.Floor, request.Button, false)
					}
					ch_completedRequest <- request
				}
			}

			// Distribute state
			if hasElevStateChanged(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
				oldElevState = localStateToElevState(myLocalState, isAvailable)
				ch_elevState <- oldElevState
			}

		case <-doorTimer.C:
			if elevio.GetObstruction() {
				doorTimer.Reset(standardDoorWait)
				isObstructed = true
				break
			}
			isObstructed = false
			var completedRequestsList []elevio.ButtonEvent
			myLocalState, myActiveRequests, completedRequestsList = handleDoorClosing(myLocalState, myActiveRequests)
			if myLocalState.getElevatorBehaviour() == DoorOpen {
				elevio.SetDoorOpenLamp(true)
				doorTimer.Reset(standardDoorWait)
				for _, request := range completedRequestsList {
					if singleElevMode {
						myActiveRequests.setRequest(request.Floor, request.Button, false)
					}
					ch_completedRequest <- request
				}
			} else {
				elevio.SetDoorOpenLamp(false)
				if myLocalState.getElevatorBehaviour() == Moving {
					elevio.SetMotorDirection(myLocalState.getLastDirection())
				}
			}

			// Distribute state
			if hasElevStateChanged(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
				oldElevState = localStateToElevState(myLocalState, isAvailable)
				ch_elevState <- oldElevState
			}

		case hallRequests := <-ch_hallRequests:
			var completedRequestsList []elevio.ButtonEvent
			emtpyCabList := [nFloors]bool{false, false, false, false}
			myLocalState, myActiveRequests, completedRequestsList = handleNewRequests(hallRequests,
				emtpyCabList,
				"hall",
				myLocalState,
				myActiveRequests,
			)
			if myLocalState.getElevatorBehaviour() == DoorOpen {
				elevio.SetDoorOpenLamp(true)
				doorTimer.Reset(standardDoorWait)
				for _, request := range completedRequestsList {
					if singleElevMode {
						myActiveRequests.setRequest(request.Floor, request.Button, false)
					}
					ch_completedRequest <- request
				}
			} else if myLocalState.getElevatorBehaviour() == Moving && oldElevState.Direction == "stop" {
				elevio.SetMotorDirection(myLocalState.getLastDirection())
			}

			// Distribute state
			if hasElevStateChanged(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
				oldElevState = localStateToElevState(myLocalState, isAvailable)
				ch_elevState <- oldElevState
			}

		case cab := <-ch_cabRequests:
			var completedRequestsList []elevio.ButtonEvent
			emptyHallList := [nFloors][2]bool{
				{false, false},
				{false, false},
				{false, false},
				{false, false},
			}
			myLocalState, myActiveRequests, completedRequestsList = handleNewRequests(emptyHallList,
				cab,
				"cab",
				myLocalState,
				myActiveRequests,
			)
			if myLocalState.getElevatorBehaviour() == DoorOpen {
				elevio.SetDoorOpenLamp(true)
				doorTimer.Reset(standardDoorWait)
				for _, request := range completedRequestsList {
					if singleElevMode {
						myActiveRequests.setRequest(request.Floor, request.Button, false)
					}
					ch_completedRequest <- request
				}
			} else if myLocalState.getElevatorBehaviour() == Moving && oldElevState.Direction == "stop" {
				elevio.SetMotorDirection(myLocalState.getLastDirection())
			}

			// Distribute state
			if hasElevStateChanged(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
				oldElevState = localStateToElevState(myLocalState, isAvailable)
				ch_elevState <- oldElevState
			}

		case singleBool := <-ch_singleElevMode:
			singleElevMode = singleBool
		case <-time.After(25 * time.Millisecond):
			// Error handling
			if hasLocalElevStateChanged(oldLocalState, myLocalState) {
				// Reset timer if local state has changed
				errorTimer = time.Now().UnixMilli()
				oldLocalState = myLocalState
			}
			if myActiveRequests.anyRequest() {
				if time.Now().UnixMilli()-errorTimer > 7500 {
					// If there are requests and timer has run out, make elevator unavailable
					isAvailable = false
				} else {
					isAvailable = true && !isObstructed
				}
			} else {
				errorTimer = time.Now().UnixMilli()
				isAvailable = true && !isObstructed
			}

			// Distribute state
			if hasElevStateChanged(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
				oldElevState = localStateToElevState(myLocalState, isAvailable)
				ch_elevState <- oldElevState
			}
		}
	}
}
