package singleelevator

import (
	"elevatorproject/config"
	"elevatorproject/singleelevator/elevatortimers"
	"elevatorproject/singleelevator/elevio"
	"time"
)

const nFloors int = config.NumFloors

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

func stopAfterSensingFloor(floor int, state localElevState, orders activeOrders) bool {
	if floor == 0 || floor == nFloors-1 {
		return true
	}
	if !orders.anyOrder() {
		return true
	}
	if orders.getSpecificOrder(floor, elevio.BT_Cab) {
		return true
	}
	switch state.getLastDirection() {
	case elevio.MD_Up:
		if orders.getSpecificOrder(floor, elevio.BT_HallUp) ||
			!orders.anyOrderPastFloorInDir(floor, elevio.MD_Up) {
			return true
		}
		return false
	case elevio.MD_Down:
		if orders.getSpecificOrder(floor, elevio.BT_HallDown) ||
			!orders.anyOrderPastFloorInDir(floor, elevio.MD_Down) {
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

func diffElevStateStructs(a ElevState, b ElevState) bool {
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
	orders activeOrders,
) (
	localElevState, 
	activeOrders, 
	[]elevio.ButtonEvent,
) {
	state.setLastFloor(floor)
	completedOrdersList := make([]elevio.ButtonEvent, 0)

	if stopAfterSensingFloor(floor, state, orders) {
		state.setElevatorBehaviour(DoorOpen)
		if orders.getSpecificOrder(floor, elevio.BT_Cab) {
			completedOrdersList = append(completedOrdersList, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab})
		}
		if orders.getSpecificOrder(floor, dirToBtnType(state.getLastDirection())) {
			completedOrdersList = append(completedOrdersList, elevio.ButtonEvent{Floor: floor, Button: dirToBtnType(state.getLastDirection())})
		}
		if floor == 0 && orders.getSpecificOrder(floor, elevio.BT_HallUp) {
			completedOrdersList = append(completedOrdersList, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp})
		}
		if floor == nFloors-1 && orders.getSpecificOrder(floor, elevio.BT_HallDown) {
			completedOrdersList = append(completedOrdersList, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown})
		}
	}
	return state, orders, completedOrdersList
}

func handleNewRequests(
	hallRequests [nFloors][2]bool, 
	cabRequests [nFloors]bool, 
	cabOrHall string, 
	state localElevState, 
	orders activeOrders,
) (
	localElevState, 
	activeOrders, 
	[]elevio.ButtonEvent,
) {
	if cabOrHall == "cab" {
		for i := 0; i < nFloors; i++ {
			orders.setOrder(i, elevio.BT_Cab, cabRequests[i])
		}
	} else if cabOrHall == "hall" {
		for i := 0; i < nFloors; i++ {
			orders.setOrder(i, elevio.BT_HallUp, hallRequests[i][0])
			orders.setOrder(i, elevio.BT_HallDown, hallRequests[i][1])
		}
	} else {
		panic("handleNewRequests: hallOrCab must be 'cab' or 'hall'")
	}

	completedRequestsList := make([]elevio.ButtonEvent, 0)

	switch state.getElevatorBehaviour() {
	case Moving:
	case DoorOpen:
		upThisFloor, downThisFloor, cabThisFloor := orders.getOrdersInFloor(state.getLastFloor())
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
		upThisFloor, downThisFloor, cabThisFloor := orders.getOrdersInFloor(state.getLastFloor())
		orderInThisFloor := upThisFloor || downThisFloor || cabThisFloor
		orders_above := orders.anyOrderPastFloorInDir(state.getLastFloor(), elevio.MD_Up)
		orders_below := orders.anyOrderPastFloorInDir(state.getLastFloor(), elevio.MD_Down)
		if !orders.anyOrder() {
			break
		}
		if orderInThisFloor {
			if cabThisFloor {
				state.setElevatorBehaviour(DoorOpen)
				completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_Cab})
			}
			if (upThisFloor && state.getLastDirection() == elevio.MD_Up) ||
				(upThisFloor && !downThisFloor && !orders.anyOrderPastFloorInDir(state.getLastFloor(), elevio.MD_Down)) ||
				(state.getLastFloor() == 0 && upThisFloor) {
				state.setElevatorBehaviour(DoorOpen)
				completedRequestsList = append(completedRequestsList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallUp})
			}
			if (downThisFloor && state.getLastDirection() == elevio.MD_Down) ||
				(downThisFloor && !upThisFloor && !orders.anyOrderPastFloorInDir(state.getLastFloor(), elevio.MD_Up)) ||
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
	return state, orders, completedRequestsList
}

func handleDoorClosing(state localElevState, orders activeOrders) (localElevState, activeOrders, []elevio.ButtonEvent) {
	completedOrdersList := make([]elevio.ButtonEvent, 0)

	if !orders.anyOrder() {
		state.setElevatorBehaviour(Idle)
		return state, orders, completedOrdersList
	}
	if orders.getSpecificOrder(state.getLastFloor(), dirToBtnType(state.getLastDirection())) {
		if state.getLastDirection() == elevio.MD_Up {
			completedOrdersList = append(completedOrdersList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallUp})
		} else {
			completedOrdersList = append(completedOrdersList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallDown})
		}
		return state, orders, completedOrdersList
	} else if orders.orderInFloor(state.getLastFloor()) && !orders.anyOrderPastFloorInDir(state.getLastFloor(), state.getLastDirection()) {
		if state.getLastDirection() == elevio.MD_Up {
			completedOrdersList = append(completedOrdersList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallDown})
		} else {
			completedOrdersList = append(completedOrdersList, elevio.ButtonEvent{Floor: state.getLastFloor(), Button: elevio.BT_HallUp})
		}
		return state, orders, completedOrdersList
	}
	state.setElevatorBehaviour(Moving)
	ordersUpBool := orders.anyOrderPastFloorInDir(state.getLastFloor(), elevio.MD_Up)
	ordersDownBool := orders.anyOrderPastFloorInDir(state.getLastFloor(), elevio.MD_Down)
	if (state.getLastDirection() == elevio.MD_Up && ordersUpBool) || !ordersDownBool {
		state.setDirection(elevio.MD_Up)
	} else if (state.getLastDirection() == elevio.MD_Down && ordersDownBool) || !ordersUpBool {
		state.setDirection(elevio.MD_Down)
	} else {
		panic("handleDoorClosing: No direction set")
	}
	return state, orders, completedOrdersList
}

func checkChangeLocalStates(a localElevState, b localElevState) bool {
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
	ch_btn 				<-chan elevio.ButtonEvent,
	ch_floor 			<-chan int,
	ch_door 			<-chan int,
	ch_hallRequests     <-chan [nFloors][2]bool,
	ch_cabRequests      <-chan [nFloors]bool,
	ch_singleElevMode   <-chan bool,
	ch_completedRequest chan<- elevio.ButtonEvent,
	ch_newRequest       chan<- elevio.ButtonEvent,
	ch_elevState        chan<- ElevState,
) {

	var myLocalState localElevState
	var myActiveOrders activeOrders
	isAvailable := false
	singleElevMode := true
	errorTimer := time.Now().UnixMilli()
	var oldLocalState localElevState
	oldLocalState.initLocalElevState()

	// Initiate elevator
	myActiveOrders.initOrders()
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
	if diffElevStateStructs(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
		oldElevState = localStateToElevState(myLocalState, isAvailable)
		ch_elevState <- oldElevState
	}

	// Main loop
	for {
		select {
		case btnPress := <-ch_btn:
			ch_newRequest <- elevio.ButtonEvent{Floor: btnPress.Floor, Button: btnPress.Button}
			if singleElevMode {
				myActiveOrders.setOrder(btnPress.Floor, btnPress.Button, true)

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
					hallReqList = myActiveOrders.getHallRequests()
				} else {
					cabReqList = myActiveOrders.getCabRequests()
				}
				var completedOrdersList []elevio.ButtonEvent
				myLocalState, myActiveOrders, completedOrdersList = handleNewRequests(hallReqList, 
																						cabReqList, 
																						cabOrHall, 
																						myLocalState, 
																						myActiveOrders,
																					)
				if myLocalState.getElevatorBehaviour() == DoorOpen {
					elevio.SetDoorOpenLamp(true)
					elevatortimers.StartDoorTimer()
					for _, order := range completedOrdersList {
						myActiveOrders.setOrder(order.Floor, order.Button, false)
						ch_completedRequest <- order
					}
				} else if myLocalState.getElevatorBehaviour() == Moving && oldElevState.Direction == "stop" {
					elevio.SetMotorDirection(myLocalState.getLastDirection())
				}

				// Distribute state
				if diffElevStateStructs(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
					oldElevState = localStateToElevState(myLocalState, isAvailable)
					ch_elevState <- oldElevState
				}

			} else {
				ch_newRequest <- elevio.ButtonEvent{Floor: btnPress.Floor, Button: btnPress.Button}
			}

		case floor := <-ch_floor:
			elevio.SetFloorIndicator(floor)
			var completedOrdersList []elevio.ButtonEvent
			myLocalState, myActiveOrders, completedOrdersList = handleFloorSensor(floor, myLocalState, myActiveOrders)
			if myLocalState.getElevatorBehaviour() == DoorOpen {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				elevatortimers.StartDoorTimer()
				for _, order := range completedOrdersList {
					if singleElevMode {
						myActiveOrders.setOrder(order.Floor, order.Button, false)
					}
					ch_completedRequest <- order
				}
			}
			
			// Distribute state
			if diffElevStateStructs(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
				oldElevState = localStateToElevState(myLocalState, isAvailable)
				ch_elevState <- oldElevState
			}

		case <-ch_door:
			if elevio.GetObstruction() {
				elevatortimers.StartDoorTimer()
				break
			}
			var completedOrdersList []elevio.ButtonEvent
			myLocalState, myActiveOrders, completedOrdersList = handleDoorClosing(myLocalState, myActiveOrders)
			if myLocalState.getElevatorBehaviour() == DoorOpen {
				elevio.SetDoorOpenLamp(true)
				elevatortimers.StartDoorTimer()
				for _, order := range completedOrdersList {
					if singleElevMode {
						myActiveOrders.setOrder(order.Floor, order.Button, false)
					}
					ch_completedRequest <- order
				}
			} else {
				elevio.SetDoorOpenLamp(false)
				if myLocalState.getElevatorBehaviour() == Moving {
					elevio.SetMotorDirection(myLocalState.getLastDirection())
				}
			}

			// Distribute state
			if diffElevStateStructs(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
				oldElevState = localStateToElevState(myLocalState, isAvailable)
				ch_elevState <- oldElevState
			}

		case hallRequests := <-ch_hallRequests:
			var completedOrdersList []elevio.ButtonEvent
			emtpyCabList := [nFloors]bool{false, false, false, false}
			myLocalState, myActiveOrders, completedOrdersList = handleNewRequests(hallRequests, 
																					emtpyCabList, 
																					"hall", 
																					myLocalState, 
																					myActiveOrders,
																				)
			if myLocalState.getElevatorBehaviour() == DoorOpen {
				elevio.SetDoorOpenLamp(true)
				elevatortimers.StartDoorTimer()
				for _, order := range completedOrdersList {
					if singleElevMode {
						myActiveOrders.setOrder(order.Floor, order.Button, false)
					}
					ch_completedRequest <- order
				}
			} else if myLocalState.getElevatorBehaviour() == Moving && oldElevState.Direction == "stop" {
				elevio.SetMotorDirection(myLocalState.getLastDirection())
			}
			
			// Distribute state
			if diffElevStateStructs(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
				oldElevState = localStateToElevState(myLocalState, isAvailable)
				ch_elevState <- oldElevState
			}

		case cab := <-ch_cabRequests:
			var completedOrdersList []elevio.ButtonEvent
			emptyHallList := [nFloors][2]bool{
										{false, false}, 
										{false, false}, 
										{false, false}, 
										{false, false},
									}
			myLocalState, myActiveOrders, completedOrdersList = handleNewRequests(emptyHallList, 
																					cab, 
																					"cab", 
																					myLocalState, 
																					myActiveOrders,
																				)
			if myLocalState.getElevatorBehaviour() == DoorOpen {
				elevio.SetDoorOpenLamp(true)
				elevatortimers.StartDoorTimer()
				for _, order := range completedOrdersList {
					if singleElevMode {
						myActiveOrders.setOrder(order.Floor, order.Button, false)
					}
					ch_completedRequest <- order
				}
			} else if myLocalState.getElevatorBehaviour() == Moving && oldElevState.Direction == "stop" {
				elevio.SetMotorDirection(myLocalState.getLastDirection())
			}

			// Distribute state
			if diffElevStateStructs(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
				oldElevState = localStateToElevState(myLocalState, isAvailable)
				ch_elevState <- oldElevState
			}
		
		case singleBool := <-ch_singleElevMode:
			singleElevMode = singleBool
		case <-time.After(25 * time.Millisecond):
			if checkChangeLocalStates(oldLocalState, myLocalState) {
				errorTimer = time.Now().UnixMilli()
				oldLocalState = myLocalState
			}
			if myActiveOrders.anyOrder() {
				if time.Now().UnixMilli()-errorTimer > 7500 {
					isAvailable = false
				} else {
					isAvailable = true
				}
			} else {
				errorTimer = time.Now().UnixMilli()
				isAvailable = true
			}
			if diffElevStateStructs(oldElevState, localStateToElevState(myLocalState, isAvailable)) {
				oldElevState = localStateToElevState(myLocalState, isAvailable)
				ch_elevState <- oldElevState
			}
		}
	}
}
