package singleelevator

import (
	"elevatorproject/config"
	"elevatorproject/singleelevator/elevator_timers"
	"elevatorproject/singleelevator/elevio"
	"fmt"
)

const nFloors int = config.NumFloors

// Struct to distribute elevator state to other modules
type ElevState struct {
	Behaviour   string
	Floor       int
	Direction   string
	IsAvailable bool
}

func (elev_state *ElevState) InitElevState() {
	elev_state.Behaviour = "idle"
	elev_state.Floor = 1
	elev_state.Direction = "stop"
	elev_state.IsAvailable = true
}

func stopAfterSensingFloor(floor int, elev_states localElevState, active_orders activeOrders) bool {
	if floor == 0 || floor == nFloors-1 {
		return true
	}
	if !active_orders.anyOrder() {
		return true
	}
	if active_orders.getSpecificOrder(floor, elevio.BT_Cab) {
		return true
	}
	switch elev_states.getLastDirection() {
	case elevio.MD_Up:
		if active_orders.getSpecificOrder(floor, elevio.BT_HallUp) || !active_orders.anyOrderPastFloorInDir(floor, elevio.MD_Up) {
			return true
		}
		return false
	case elevio.MD_Down:
		if active_orders.getSpecificOrder(floor, elevio.BT_HallDown) || !active_orders.anyOrderPastFloorInDir(floor, elevio.MD_Down) {
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

func localStateToElevState(states localElevState, isAvailable bool) ElevState {
	hra_states := ElevState{
		Behaviour:   "idle",
		Floor:       1,
		Direction:   "stop",
		IsAvailable: true}

	// Floor
	if states.getLastFloor() == -1 {
		hra_states.Floor = nFloors / 2
	} else {
		hra_states.Floor = states.getLastFloor()
	}

	// Direction
	switch states.getLastDirection() {
	case elevio.MD_Up:
		hra_states.Direction = "up"
	case elevio.MD_Down:
		hra_states.Direction = "down"
	}

	// Behaviour
	switch states.getElevatorBehaviour() {
	case "Idle":
		hra_states.Behaviour = "idle"
		hra_states.Direction = "stop"
	case "Moving":
		hra_states.Behaviour = "moving"
	case "DoorOpen":
		hra_states.Behaviour = "doorOpen"
		hra_states.Direction = "stop"
	default:
		panic("Invalid elevator behaviour")
	}

	// isAvailable
	hra_states.IsAvailable = isAvailable

	return hra_states
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
func handleFloorSensor(floor int, elev_states localElevState, active_orders activeOrders) (localElevState, activeOrders, []elevio.ButtonEvent) {
	elev_states.setLastFloor(floor)

	remove_orders_list := make([]elevio.ButtonEvent, 0)

	if stopAfterSensingFloor(floor, elev_states, active_orders) {
		elev_states.setElevatorBehaviour("DoorOpen")
		if active_orders.getSpecificOrder(floor, elevio.BT_Cab) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab})
		}
		if active_orders.getSpecificOrder(floor, dirToBtnType(elev_states.getLastDirection())) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: floor, Button: dirToBtnType(elev_states.getLastDirection())})
		}
		if floor == 0 && active_orders.getSpecificOrder(floor, elevio.BT_HallUp) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp})
		}
		if floor == nFloors-1 && active_orders.getSpecificOrder(floor, elevio.BT_HallDown) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown})
		}
	}
	return elev_states, active_orders, remove_orders_list
}

func handleNewRequests(hra [config.NumFloors][2]bool, cab_call [config.NumFloors]bool, hra_or_cab string, elev_states localElevState, active_orders activeOrders) (localElevState, activeOrders, []elevio.ButtonEvent) {
	if hra_or_cab == "cab" {
		// If cab orders
		for i := 0; i < nFloors; i++ {
			active_orders.setOrder(i, elevio.BT_Cab, cab_call[i])
		}
	} else if hra_or_cab == "hra" {
		// If HRA orders
		for i := 0; i < nFloors; i++ {
			active_orders.setOrder(i, elevio.BT_HallUp, hra[i][0]) // First columns is up, second is down
			active_orders.setOrder(i, elevio.BT_HallDown, hra[i][1])
		}
	} else {
		panic("handleNewRequests: hra_or_cab must be 'cab' or 'hra'")
	}
	remove_orders_list := make([]elevio.ButtonEvent, 0)

	switch elev_states.getElevatorBehaviour() {
	case Moving:
	case DoorOpen:
		up_this_floor, down_this_floor, cab_this_floor := active_orders.getOrdersInFloor(elev_states.getLastFloor())
		if cab_this_floor {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.getLastFloor(), Button: elevio.BT_Cab})
		}
		if (up_this_floor && elev_states.getLastDirection() == elevio.MD_Up) ||
			(elev_states.getLastFloor() == 0 && up_this_floor) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.getLastFloor(), Button: elevio.BT_HallUp})
		}
		if (down_this_floor && elev_states.getLastDirection() == elevio.MD_Down) ||
			(elev_states.getLastFloor() == nFloors-1 && down_this_floor) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.getLastFloor(), Button: elevio.BT_HallDown})
		}
	case Idle:
		up_this_floor, down_this_floor, cab_this_floor := active_orders.getOrdersInFloor(elev_states.getLastFloor())
		order_in_this_floor := up_this_floor || down_this_floor || cab_this_floor
		orders_above := active_orders.anyOrderPastFloorInDir(elev_states.getLastFloor(), elevio.MD_Up)
		orders_below := active_orders.anyOrderPastFloorInDir(elev_states.getLastFloor(), elevio.MD_Down)
		if !active_orders.anyOrder() {
			break
		} else {
			if order_in_this_floor {
				if cab_this_floor {
					elev_states.setElevatorBehaviour("DoorOpen")
					remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.getLastFloor(), Button: elevio.BT_Cab})
				}
				if (up_this_floor && elev_states.getLastDirection() == elevio.MD_Up) ||
					(up_this_floor && !down_this_floor && !active_orders.anyOrderPastFloorInDir(elev_states.getLastFloor(), elevio.MD_Down)) ||
					(elev_states.getLastFloor() == 0 && up_this_floor) {
					elev_states.setElevatorBehaviour("DoorOpen")
					remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.getLastFloor(), Button: elevio.BT_HallUp})
				}
				if (down_this_floor && elev_states.getLastDirection() == elevio.MD_Down) ||
					(down_this_floor && !up_this_floor && !active_orders.anyOrderPastFloorInDir(elev_states.getLastFloor(), elevio.MD_Up)) ||
					(elev_states.getLastFloor() == nFloors-1 && down_this_floor) {
					elev_states.setElevatorBehaviour("DoorOpen")
					remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.getLastFloor(), Button: elevio.BT_HallDown})
				}

			}
			if !order_in_this_floor || elev_states.getElevatorBehaviour() != "DoorOpen" {
				elev_states.setElevatorBehaviour("Moving")
				if (elev_states.getLastDirection() == elevio.MD_Up && orders_above) || !orders_below {
					elev_states.setDirection(elevio.MD_Up)
				} else if (elev_states.getLastDirection() == elevio.MD_Down && orders_below) || !orders_above {
					elev_states.setDirection(elevio.MD_Down)
				} else {
					panic("HandleNewRequests: No direction set")
				}
			}
		}
	}
	return elev_states, active_orders, remove_orders_list
}

func handleDoorClosing(elev_states localElevState, active_orders activeOrders) (localElevState, activeOrders, []elevio.ButtonEvent) {
	remove_orders_list := make([]elevio.ButtonEvent, 0)

	if !active_orders.anyOrder() {
		elev_states.setElevatorBehaviour("Idle")
		return elev_states, active_orders, remove_orders_list
	}
	if active_orders.getSpecificOrder(elev_states.getLastFloor(), dirToBtnType(elev_states.getLastDirection())) {
		if elev_states.getLastDirection() == elevio.MD_Up {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.getLastFloor(), Button: elevio.BT_HallUp})
		} else {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.getLastFloor(), Button: elevio.BT_HallDown})
		}
		return elev_states, active_orders, remove_orders_list
	} else if active_orders.orderInFloor(elev_states.getLastFloor()) && !active_orders.anyOrderPastFloorInDir(elev_states.getLastFloor(), elev_states.getLastDirection()) {
		if elev_states.getLastDirection() == elevio.MD_Up {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.getLastFloor(), Button: elevio.BT_HallDown})
		} else {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.getLastFloor(), Button: elevio.BT_HallUp})
		}
		return elev_states, active_orders, remove_orders_list
	}
	elev_states.setElevatorBehaviour("Moving")
	orders_up_bool := active_orders.anyOrderPastFloorInDir(elev_states.getLastFloor(), elevio.MD_Up)
	orders_down_bool := active_orders.anyOrderPastFloorInDir(elev_states.getLastFloor(), elevio.MD_Down)
	if (elev_states.getLastDirection() == elevio.MD_Up && orders_up_bool) || !orders_down_bool {
		elev_states.setDirection(elevio.MD_Up)
	} else if (elev_states.getLastDirection() == elevio.MD_Down && orders_down_bool) || !orders_up_bool {
		elev_states.setDirection(elevio.MD_Down)
	} else {
		panic("HandleDoorClosing: No direction set")
	}
	return elev_states, active_orders, remove_orders_list
}

func changeInBehaviour(oldState ElevState, currentState localElevState) bool {
	behaviour_translation := "idle"
	switch currentState.getElevatorBehaviour() {
	case "Idle":
		behaviour_translation = "idle"
	case "Moving":
		behaviour_translation = "moving"
	case "DoorOpen":
		behaviour_translation = "doorOpen"
	default:
		panic("Invalid elevator behaviour")
	}
	return oldState.Behaviour != behaviour_translation
}

// Fikse spam error
// Skal startes når døren åpner seg og det finnes ordre i andre etasjer for denne heisen
// Skal stoppes når heisen kommer seg av gårde, altså et sted i closing doors.

// TODO: Må legge til elevator_timers.StopSpamTimer() noen steder!

// Ny error metode fra Nicholas:
// Kun sjekke endringer i state.
// Hvis det finnes ordre og heisen blir for lenge i en state skal den settes til isAvailable = false.

// Timeren må resettes hver gang heisen blir satt i en ny state.
// Timeren skal stoppes hver gang det ikke er ordre.
// Timeren skal startes hver gang det går til å finnes en ordre.

func Fsm_elevator(
	ch_btn <-chan elevio.ButtonEvent,
	ch_floor <-chan int,
	ch_door <-chan int,
	ch_error <-chan int,
	ch_hra <-chan [config.NumFloors][2]bool,
	ch_cab_requests <-chan [config.NumFloors]bool,
	ch_single_mode <-chan bool,
	ch_completed_request chan<- elevio.ButtonEvent,
	ch_new_request chan<- elevio.ButtonEvent,
	ch_elevstate chan<- ElevState,
) {

	var elevState localElevState
	var activeOrders activeOrders
	isAvailable := false

	singleElevMode := true

	// Initiate elevator
	activeOrders.initOrders()
	elevState.initLocalElevState()
	var oldElevInfo ElevState = localStateToElevState(elevState, isAvailable)
	if elevio.GetFloor() == -1 {
		elevState.setElevatorBehaviour("Moving")
		elevio.SetMotorDirection(elevio.MD_Up)
	} else {
		elevState.setLastFloor(elevio.GetFloor())
		elevio.SetFloorIndicator(elevio.GetFloor())
		isAvailable = true
	}
	if diffElevStateStructs(oldElevInfo, localStateToElevState(elevState, isAvailable)) {
		oldElevInfo = localStateToElevState(elevState, isAvailable)
		ch_elevstate <- oldElevInfo
	}

	// Finite state machine
	for {
		select {
		case hra := <-ch_hra: // SPM: Mottas det HRA om heisen er i single elev mode??
			//fmt.Println("HandleHRA")
			var remove_orders_list []elevio.ButtonEvent
			empty_cab_list := [config.NumFloors]bool{false, false, false, false}
			elevState, activeOrders, remove_orders_list = handleNewRequests(hra, empty_cab_list, "hra", elevState, activeOrders)
			if elevState.getElevatorBehaviour() == "DoorOpen" {
				elevio.SetDoorOpenLamp(true)
				elevator_timers.StartDoorTimer()
				for _, v := range remove_orders_list {
					if singleElevMode {
						activeOrders.setOrder(v.Floor, v.Button, false)
					}
					ch_completed_request <- v
				}
			} else if elevState.getElevatorBehaviour() == "Moving" && oldElevInfo.Direction == "stop" { // Vil dette forårsake "hakkete" oppførsel? Må sjekkes. (Endret til å sjekke oldElevInfo)
				elevio.SetMotorDirection(elevState.getLastDirection())
			}
			// Eventuelt sende full info til Nodeview og sende cabstatus til lampmodul, oppdatere isAvailable
			// Error handling start
			// -- Hvis det ikke er noen ordre skal timeren stoppes.
			// -- -- isAvailable settes true.
			// -- Hvis timeren er i gang og det skjer en endring i behaviour så skal timeren resettes.
			// -- -- isAvailable settes true.
			// -- Hvis timeren ikke er i gang og det er ordre i systemet skal timeren startes.
			if elevator_timers.GetErrorCounter() == -1 {
				if activeOrders.anyOrder() {
					elevator_timers.StartErrorTimer()
					isAvailable = true
				}
			} else {
				if !activeOrders.anyOrder() {
					elevator_timers.StopErrorTimer()
					isAvailable = true
				} else if changeInBehaviour(oldElevInfo, elevState) {
					elevator_timers.StartErrorTimer()
					isAvailable = true
				}
			}
			// Error handling end
			if diffElevStateStructs(oldElevInfo, localStateToElevState(elevState, isAvailable)) {
				oldElevInfo = localStateToElevState(elevState, isAvailable)
				ch_elevstate <- oldElevInfo
			}
		case floor := <-ch_floor:
			fmt.Println("HandleFloorSensor")
			elevio.SetFloorIndicator(floor)
			var remove_orders_list []elevio.ButtonEvent
			elevState, activeOrders, remove_orders_list = handleFloorSensor(floor, elevState, activeOrders)
			if elevState.getElevatorBehaviour() == "DoorOpen" {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				elevator_timers.StartDoorTimer()
				for _, v := range remove_orders_list {
					if singleElevMode {
						activeOrders.setOrder(v.Floor, v.Button, false)
					}
					ch_completed_request <- v
				}
			}
			// Eventuelt sende full info til Nodeview og sende cabstatus til lampmodul, oppdatere isAvailable
			// Error handling start
			// -- Hvis det ikke er noen ordre skal timeren stoppes.
			// -- -- isAvailable settes true.
			// -- Hvis timeren er i gang og det skjer en endring i behaviour så skal timeren resettes.
			// -- -- isAvailable settes true.
			// -- Hvis timeren ikke er i gang og det er ordre i systemet skal timeren startes.
			if elevator_timers.GetErrorCounter() == -1 {
				if activeOrders.anyOrder() {
					elevator_timers.StartErrorTimer()
					isAvailable = true
				}
			} else {
				if !activeOrders.anyOrder() {
					elevator_timers.StopErrorTimer()
					isAvailable = true
				} else if changeInBehaviour(oldElevInfo, elevState) {
					elevator_timers.StartErrorTimer()
					isAvailable = true
				}
			}
			// Error handling end
			if diffElevStateStructs(oldElevInfo, localStateToElevState(elevState, isAvailable)) {
				oldElevInfo = localStateToElevState(elevState, isAvailable)
				ch_elevstate <- oldElevInfo
			}
		case <-ch_door:
			//fmt.Println("HandleDoorClosing")
			if elevio.GetObstruction() {
				elevator_timers.StartDoorTimer()
				fmt.Println("Obstruction detected")
				break
			}
			var remove_orders_list []elevio.ButtonEvent
			elevState, activeOrders, remove_orders_list = handleDoorClosing(elevState, activeOrders)
			if elevState.getElevatorBehaviour() == "DoorOpen" {
				elevio.SetDoorOpenLamp(true)
				elevator_timers.StartDoorTimer()
				for _, v := range remove_orders_list {
					if singleElevMode {
						activeOrders.setOrder(v.Floor, v.Button, false)
					}
					ch_completed_request <- v
				}
			} else {
				elevio.SetDoorOpenLamp(false)
				if elevState.getElevatorBehaviour() == "Moving" {
					elevio.SetMotorDirection(elevState.getLastDirection())
				}
			}
			// Eventuelt sende full info til Nodeview og sende cabstatus til lampmodul, oppdatere isAvailable
			// Error handling start
			// -- Hvis det ikke er noen ordre skal timeren stoppes.
			// -- -- isAvailable settes true.
			// -- Hvis timeren er i gang og det skjer en endring i behaviour så skal timeren resettes.
			// -- -- isAvailable settes true.
			// -- Hvis timeren ikke er i gang og det er ordre i systemet skal timeren startes.
			if elevator_timers.GetErrorCounter() == -1 {
				if activeOrders.anyOrder() {
					elevator_timers.StartErrorTimer()
					isAvailable = true
				}
			} else {
				if !activeOrders.anyOrder() {
					elevator_timers.StopErrorTimer()
					isAvailable = true
				} else if changeInBehaviour(oldElevInfo, elevState) {
					elevator_timers.StartErrorTimer()
					isAvailable = true
				}
			}
			// Error handling end
			if diffElevStateStructs(oldElevInfo, localStateToElevState(elevState, isAvailable)) {
				oldElevInfo = localStateToElevState(elevState, isAvailable)
				ch_elevstate <- oldElevInfo
			}
		case btn_press := <-ch_btn:
			//fmt.Println("HandleButtonEvent")
			ch_new_request <- elevio.ButtonEvent{Floor: btn_press.Floor, Button: btn_press.Button}
			if singleElevMode {
				activeOrders.setOrder(btn_press.Floor, btn_press.Button, true)

				cab_req_list := [config.NumFloors]bool{false, false, false, false}
				hall_req_list := [config.NumFloors][2]bool{{false, false}, {false, false}, {false, false}, {false, false}}
				cab_or_hra := "cab"

				if btn_press.Button != elevio.BT_Cab {
					cab_or_hra = "hra"
					hall_req_list = activeOrders.getHallRequests()
				} else {
					cab_req_list = activeOrders.getCabRequests()
				}
				var remove_orders_list []elevio.ButtonEvent
				elevState, activeOrders, remove_orders_list = handleNewRequests(hall_req_list, cab_req_list, cab_or_hra, elevState, activeOrders)
				if elevState.getElevatorBehaviour() == "DoorOpen" {
					elevio.SetDoorOpenLamp(true)
					elevator_timers.StartDoorTimer()
					for _, v := range remove_orders_list {
						activeOrders.setOrder(v.Floor, v.Button, false)
						ch_completed_request <- v
					}
				} else if elevState.getElevatorBehaviour() == "Moving" && oldElevInfo.Direction == "stop" { // Vil dette forårsake "hakkete" oppførsel? Må sjekkes. (Endret til å sjekke oldElevInfo)
					elevio.SetMotorDirection(elevState.getLastDirection())
				}
				// Eventuelt sende full info til Nodeview og sende cabstatus til lampmodul, oppdatere isAvailable
				// Error handling start
				// -- Hvis det ikke er noen ordre skal timeren stoppes.
				// -- -- isAvailable settes true.
				// -- Hvis timeren er i gang og det skjer en endring i behaviour så skal timeren resettes.
				// -- -- isAvailable settes true.
				// -- Hvis timeren ikke er i gang og det er ordre i systemet skal timeren startes.
				if elevator_timers.GetErrorCounter() == -1 {
					if activeOrders.anyOrder() {
						elevator_timers.StartErrorTimer()
						isAvailable = true
					}
				} else {
					if !activeOrders.anyOrder() {
						elevator_timers.StopErrorTimer()
						isAvailable = true
					} else if changeInBehaviour(oldElevInfo, elevState) {
						elevator_timers.StartErrorTimer()
						isAvailable = true
					}
				}
				// Error handling end
				if diffElevStateStructs(oldElevInfo, localStateToElevState(elevState, isAvailable)) {
					oldElevInfo = localStateToElevState(elevState, isAvailable)
					ch_elevstate <- oldElevInfo
				}
			} else {
				ch_new_request <- elevio.ButtonEvent{Floor: btn_press.Floor, Button: btn_press.Button}
			}
		case <-ch_error:
			fmt.Println("HandleError")
			isAvailable = false
			oldElevInfo = localStateToElevState(elevState, isAvailable)
			ch_elevstate <- oldElevInfo
		case single_bool := <-ch_single_mode:
			singleElevMode = single_bool
		case cab := <-ch_cab_requests:
			fmt.Println("HandleCabRequests")
			var remove_orders_list []elevio.ButtonEvent
			empty_hra_list := [config.NumFloors][2]bool{{false, false}, {false, false}, {false, false}, {false, false}}
			elevState, activeOrders, remove_orders_list = handleNewRequests(empty_hra_list, cab, "cab", elevState, activeOrders)
			if elevState.getElevatorBehaviour() == "DoorOpen" {
				elevio.SetDoorOpenLamp(true)
				elevator_timers.StartDoorTimer()
				for _, v := range remove_orders_list {
					if singleElevMode { // SPM: Er denne i det hele tatt nødvendig? Altså kan man få cab liste i single elev?
						activeOrders.setOrder(v.Floor, v.Button, false)
					}
					ch_completed_request <- v
				}
			} else if elevState.getElevatorBehaviour() == "Moving" && oldElevInfo.Direction == "stop" {
				elevio.SetMotorDirection(elevState.getLastDirection())
			}
			// Eventuelt sende full info til Nodeview og sende cabstatus til lampmodul, oppdatere isAvailable
			// Error handling start
			// -- Hvis det ikke er noen ordre skal timeren stoppes.
			// -- -- isAvailable settes true.
			// -- Hvis timeren er i gang og det skjer en endring i behaviour så skal timeren resettes.
			// -- -- isAvailable settes true.
			// -- Hvis timeren ikke er i gang og det er ordre i systemet skal timeren startes.
			if elevator_timers.GetErrorCounter() == -1 {
				if activeOrders.anyOrder() {
					elevator_timers.StartErrorTimer()
					isAvailable = true
				}
			} else {
				if !activeOrders.anyOrder() {
					elevator_timers.StopErrorTimer()
					isAvailable = true
				} else if changeInBehaviour(oldElevInfo, elevState) {
					elevator_timers.StartErrorTimer()
					isAvailable = true
				}
			}
			// Error handling end
			if diffElevStateStructs(oldElevInfo, localStateToElevState(elevState, isAvailable)) {
				oldElevInfo = localStateToElevState(elevState, isAvailable)
				ch_elevstate <- oldElevInfo
			}
		}
	}
}
