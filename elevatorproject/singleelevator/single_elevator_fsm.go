package singleelevator

import (
	"elevatorproject/config"
	"elevatorproject/singleelevator/doortimer"
	"elevatorproject/singleelevator/elevio"
	"fmt"
	"strings"
	// "time"
)

// TODO: Cleanup, kanskje endre fra sånne bools i fsmen til bare å bruke States. I tillegg se på hvordan bruke 
// 	 golang timers istedenfor door timer muligens. Da kan det kanskje bare brukes rett i denne modulen. 
//   Samt også brukes til mechanical errors.
//	 -1 floor mech error må resettes hver gang et floor treffes (stoppes). Og startes hver gang vi er i moving.
//	 Obstruction må startes hver gang det er obstruksjon og stoppes (cleares) når døren lukkes  (Dette gir mening
// 	 fordi det er på mange måter simulert obstruksjon om folk spammer åpne dør knappen og heisen blir stående i den etasjen).


// Helper struct
type ElevState struct {
	Behaviour   string
	Floor       int
	Direction   string
	CabRequests []bool
	IsAvailable bool
}

// Constants
const n_floors int = config.NumFloors

// Helper functions
func PrintElevState(state ElevState) {
	fmt.Printf("Behaviour: %s\n", state.Behaviour)
	fmt.Printf("Floor: %d\n", state.Floor)
	fmt.Printf("Direction: %s\n", state.Direction)
	fmt.Printf("CabRequests: %s\n", boolSliceToString(state.CabRequests))
	fmt.Printf("IsAvailable: %t\n", state.IsAvailable)
}

func boolSliceToString(arr []bool) string {
	return strings.Join(strings.Split(fmt.Sprintf("%v", arr), " "), ", ")
}

func StopAfterSensingFloor(floor int, elev_states States, active_orders Orders) bool {
	if floor == 0 || floor == n_floors-1 {
		return true
	}
	if !active_orders.AnyOrder() {
		return true
	}
	if active_orders.GetSpecificOrder(floor, elevio.BT_Cab) {
		return true
	}
	switch elev_states.GetLastDirection() {
	case elevio.MD_Up:
		if active_orders.GetSpecificOrder(floor, elevio.BT_HallUp) || !active_orders.AnyOrderPastFloorInDir(floor, elevio.MD_Up) {
			return true
		}
		return false
	case elevio.MD_Down:
		if active_orders.GetSpecificOrder(floor, elevio.BT_HallDown) || !active_orders.AnyOrderPastFloorInDir(floor, elevio.MD_Down) {
			return true
		}
		return false
	}
	return false
}

func BtnTypeToDir(btn_type elevio.ButtonType) elevio.MotorDirection {
	switch btn_type {
	case elevio.BT_HallUp:
		return elevio.MD_Up
	case elevio.BT_HallDown:
		return elevio.MD_Down
	}
	return elevio.MD_Stop
}

func DirToBtnType(dir elevio.MotorDirection) elevio.ButtonType {
	switch dir {
	case elevio.MD_Up:
		return elevio.BT_HallUp
	case elevio.MD_Down:
		return elevio.BT_HallDown
	}
	return elevio.BT_Cab
}

func StatesToHRAStates(states States, cab_requests []bool, isAvailable bool) ElevState {
	hra_states := ElevState{
		Behaviour:   "idle",
		Floor:       1,
		Direction:   "stop",
		CabRequests: make([]bool, n_floors),
		IsAvailable: true}

	// Floor
	if states.GetLastFloor() == -1 {
		hra_states.Floor = n_floors / 2
	} else {
		hra_states.Floor = states.GetLastFloor()
	}

	// Direction
	switch states.GetLastDirection() {
	case elevio.MD_Up:
		hra_states.Direction = "up"
	case elevio.MD_Down:
		hra_states.Direction = "down"
	}

	// Behaviour
	switch states.GetElevatorBehaviour() {
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

	// Cab requests
	copy(hra_states.CabRequests, cab_requests)

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
	for i := 0; i < n_floors; i++ {
		if a.CabRequests[i] != b.CabRequests[i] {
			return true
		}
	}
	return false
}

// Functions for handling events
func HandleFloorSensor(floor int, elev_states States, active_orders Orders) (States, Orders, bool, []elevio.ButtonEvent) {
	stop_bool := false
	elev_states.SetLastFloor(floor)

	remove_orders_list := make([]elevio.ButtonEvent, 0)

	if StopAfterSensingFloor(floor, elev_states, active_orders) {
		stop_bool = true
		elev_states.SetElevatorBehaviour("DoorOpen")
		if active_orders.GetSpecificOrder(floor, elevio.BT_Cab) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab})
		}
		if active_orders.GetSpecificOrder(floor, DirToBtnType(elev_states.GetLastDirection())) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: floor, Button: DirToBtnType(elev_states.GetLastDirection())})
		}
		if floor == 0 && active_orders.GetSpecificOrder(floor, elevio.BT_HallUp) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp})
		}
		if floor == n_floors-1 && active_orders.GetSpecificOrder(floor, elevio.BT_HallDown) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown})
		}
	}
	return elev_states, active_orders, stop_bool, remove_orders_list
}

func HandleNewRequests(hra [][2]bool, cab_order_floor int, elev_states States, active_orders Orders) (States, Orders, bool, bool, []elevio.ButtonEvent) {
	// If cab order
	if cab_order_floor != -1 {
		active_orders.SetOrder(cab_order_floor, elevio.BT_Cab, true)
	} else {
		// If HRA orders
		for i := 0; i < n_floors; i++ {
			active_orders.SetOrder(i, elevio.BT_HallUp, hra[i][0]) // First columns is up, second is down
			active_orders.SetOrder(i, elevio.BT_HallDown, hra[i][1])
		}
	}
	open_door_bool := false
	set_direction_bool := false
	remove_orders_list := make([]elevio.ButtonEvent, 0)

	switch elev_states.GetElevatorBehaviour() {
	case Moving:
	case DoorOpen:
		up_this_floor, down_this_floor, cab_this_floor := active_orders.GetOrdersInFloor(elev_states.GetLastFloor())
		if cab_this_floor {
			open_door_bool = true
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.GetLastFloor(), Button: elevio.BT_Cab})
		}
		if (up_this_floor && elev_states.GetLastDirection() == elevio.MD_Up) ||
			(elev_states.GetLastFloor() == 0 && up_this_floor) {
			open_door_bool = true
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.GetLastFloor(), Button: elevio.BT_HallUp})
		}
		if (down_this_floor && elev_states.GetLastDirection() == elevio.MD_Down) ||
			(elev_states.GetLastFloor() == n_floors-1 && down_this_floor) {
			open_door_bool = true
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.GetLastFloor(), Button: elevio.BT_HallDown})
		}
	case Idle:
		up_this_floor, down_this_floor, cab_this_floor := active_orders.GetOrdersInFloor(elev_states.GetLastFloor())
		order_in_this_floor := up_this_floor || down_this_floor || cab_this_floor
		orders_above := active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elevio.MD_Up)
		orders_below := active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elevio.MD_Down)
		if !active_orders.AnyOrder() {
			break
		} else {
			if order_in_this_floor {
				if cab_this_floor {
					open_door_bool = true
					elev_states.SetElevatorBehaviour("DoorOpen")
					remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.GetLastFloor(), Button: elevio.BT_Cab})
				}
				if (up_this_floor && elev_states.GetLastDirection() == elevio.MD_Up) ||
					(up_this_floor && !down_this_floor && !active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elevio.MD_Down)) ||
					(elev_states.GetLastFloor() == 0 && up_this_floor) {
					open_door_bool = true
					elev_states.SetElevatorBehaviour("DoorOpen")
					remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.GetLastFloor(), Button: elevio.BT_HallUp})
				}
				if (down_this_floor && elev_states.GetLastDirection() == elevio.MD_Down) ||
					(down_this_floor && !up_this_floor && !active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elevio.MD_Up)) ||
					(elev_states.GetLastFloor() == n_floors-1 && down_this_floor) {
					open_door_bool = true
					elev_states.SetElevatorBehaviour("DoorOpen")
					remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.GetLastFloor(), Button: elevio.BT_HallDown})
				}

			}
			if !order_in_this_floor || !open_door_bool {
				set_direction_bool = true
				elev_states.SetElevatorBehaviour("Moving")
				if (elev_states.GetLastDirection() == elevio.MD_Up && orders_above) || !orders_below {
					elev_states.SetDirection(elevio.MD_Up)
				} else if (elev_states.GetLastDirection() == elevio.MD_Down && orders_below) || !orders_above {
					elev_states.SetDirection(elevio.MD_Down)
				} else {
					panic("HandleNewRequests: No direction set")
				}
			}
		}
	}
	return elev_states, active_orders, open_door_bool, set_direction_bool, remove_orders_list
}

func HandleDoorClosing(elev_states States, active_orders Orders) (States, Orders, bool, bool, []elevio.ButtonEvent) {
	open_door_bool := false
	set_direction_bool := false

	remove_orders_list := make([]elevio.ButtonEvent, 0)

	if !active_orders.AnyOrder() {
		elev_states.SetElevatorBehaviour("Idle")
		return elev_states, active_orders, open_door_bool, set_direction_bool, remove_orders_list
	}
	if active_orders.GetSpecificOrder(elev_states.GetLastFloor(), DirToBtnType(elev_states.GetLastDirection())) {
		open_door_bool = true
		if elev_states.GetLastDirection() == elevio.MD_Up {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.GetLastFloor(), Button: elevio.BT_HallUp})
		} else {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.GetLastFloor(), Button: elevio.BT_HallDown})
		}
		return elev_states, active_orders, open_door_bool, set_direction_bool, remove_orders_list
	} else if active_orders.OrderInFloor(elev_states.GetLastFloor()) && !active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elev_states.GetLastDirection()) {
		open_door_bool = true
		if elev_states.GetLastDirection() == elevio.MD_Up {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.GetLastFloor(), Button: elevio.BT_HallDown})
		} else {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{Floor: elev_states.GetLastFloor(), Button: elevio.BT_HallUp})
		}
		return elev_states, active_orders, open_door_bool, set_direction_bool, remove_orders_list
	}
	set_direction_bool = true
	elev_states.SetElevatorBehaviour("Moving")
	orders_up_bool := active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elevio.MD_Up)
	orders_down_bool := active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elevio.MD_Down)
	if (elev_states.GetLastDirection() == elevio.MD_Up && orders_up_bool) || !orders_down_bool {
		elev_states.SetDirection(elevio.MD_Up)
	} else if (elev_states.GetLastDirection() == elevio.MD_Down && orders_down_bool) || !orders_up_bool {
		elev_states.SetDirection(elevio.MD_Down)
	} else {
		panic("HandleDoorClosing: No direction set")
	}
	return elev_states, active_orders, open_door_bool, set_direction_bool, remove_orders_list
}

func Fsm_elevator(ch_cab_lamps chan<- []bool,
	ch_btn <-chan elevio.ButtonEvent,
	ch_floor <-chan int,
	ch_door <-chan int,
	ch_hra <-chan [][2]bool,
	ch_init_cab_requests <-chan []bool,
	ch_completed_hall_req chan<- elevio.ButtonEvent,
	ch_new_hall_req chan<- elevio.ButtonEvent,
	ch_elevstate chan<- ElevState) {

	var Elev_states States
	var Active_orders Orders

	isAvailable := true

	// Timers for isAvailable
	// obstruction_timer := time.NewTimer(10*time.Second)
	// obstruction_timer.Stop()

	// mechanical_timer := time.NewTimer(7*time.Second)
	// mechanical_timer.Stop()

	// Initiate elevator
	Active_orders.InitOrders()
	InitLamps(Active_orders) //TODO: remove this because of fix by channel
	Elev_states.InitStates()

	var Old_ElevState ElevState = StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)

	if elevio.GetFloor() == -1 {
		Elev_states.SetElevatorBehaviour("Moving")
		elevio.SetMotorDirection(elevio.MD_Up)
	} else {
		Elev_states.SetLastFloor(elevio.GetFloor())
		elevio.SetFloorIndicator(elevio.GetFloor())
	}
	if diffElevStateStructs(Old_ElevState, StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)) {
		Old_ElevState = StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
		ch_elevstate <- Old_ElevState
	}

	// Finite state machine
	for {
		select {
		case hra := <-ch_hra:
			fmt.Println("HandleHRA")
			// fmt.Println(hra)
			var open_door_bool, set_direction_bool bool
			var remove_orders_list []elevio.ButtonEvent
			Elev_states, Active_orders, open_door_bool, set_direction_bool, remove_orders_list = HandleNewRequests(hra, -1, Elev_states, Active_orders)
			if open_door_bool {
				elevio.SetDoorOpenLamp(true)
				doortimer.StartTimer()
				for _, v := range remove_orders_list {
					if v.Button == elevio.BT_Cab {
						Active_orders.SetOrder(v.Floor, v.Button, false)
					} else {
						ch_completed_hall_req <- v
					}
				}
			} else {
				if set_direction_bool {
					elevio.SetMotorDirection(Elev_states.GetLastDirection())
				}
			}
			if diffElevStateStructs(Old_ElevState, StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)) {
				Old_ElevState = StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
				ch_elevstate <- Old_ElevState
			}
			ch_cab_lamps <- Active_orders.GetCabRequests()
		case floor := <-ch_floor:
			fmt.Println("HandleFloorSensor")
			elevio.SetFloorIndicator(floor)
			var stop_bool bool
			var remove_orders_list []elevio.ButtonEvent
			Elev_states, Active_orders, stop_bool, remove_orders_list = HandleFloorSensor(floor, Elev_states, Active_orders)
			if stop_bool {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				doortimer.StartTimer()
				for _, v := range remove_orders_list {
					if v.Button == elevio.BT_Cab {
						Active_orders.SetOrder(v.Floor, v.Button, false)
					} else {
						ch_completed_hall_req <- v
					}
				}
			}
			if diffElevStateStructs(Old_ElevState, StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)) {
				Old_ElevState = StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
				ch_elevstate <- Old_ElevState
			}
			ch_cab_lamps <- Active_orders.GetCabRequests()
		case <-ch_door:
			fmt.Println("HandleDoorClosing")
			if elevio.GetObstruction() {
				doortimer.StartTimer()
				fmt.Println("Obstruction detected")
				break
			}
			var open_door_bool, set_direction_bool bool
			var remove_orders_list []elevio.ButtonEvent
			Elev_states, Active_orders, open_door_bool, set_direction_bool, remove_orders_list = HandleDoorClosing(Elev_states, Active_orders)
			if open_door_bool {
				elevio.SetDoorOpenLamp(true)
				doortimer.StartTimer()
				for _, v := range remove_orders_list {
					if v.Button == elevio.BT_Cab {
						Active_orders.SetOrder(v.Floor, v.Button, false)
					} else {
						ch_completed_hall_req <- v
					}
				}
			} else {
				elevio.SetDoorOpenLamp(false)
				if set_direction_bool {
					elevio.SetMotorDirection(Elev_states.GetLastDirection())
				}
			}
			if diffElevStateStructs(Old_ElevState, StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)) {
				Old_ElevState = StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
				ch_elevstate <- Old_ElevState
			}
			ch_cab_lamps <- Active_orders.GetCabRequests()
		case btn_press := <-ch_btn:
			fmt.Println("HandleButtonEvent")
			if btn_press.Button == elevio.BT_Cab {
				var open_door_bool, set_direction_bool bool
				var remove_orders_list []elevio.ButtonEvent
				emtpy_hra := [][2]bool{}
				Elev_states, Active_orders, open_door_bool, set_direction_bool, remove_orders_list = HandleNewRequests(emtpy_hra, btn_press.Floor, Elev_states, Active_orders)
				if open_door_bool {
					elevio.SetDoorOpenLamp(true)
					doortimer.StartTimer()
					for _, v := range remove_orders_list {
						if v.Button == elevio.BT_Cab {
							Active_orders.SetOrder(v.Floor, v.Button, false)
						} else {
							ch_completed_hall_req <- v
						}
					}
				} else {
					if set_direction_bool {
						elevio.SetMotorDirection(Elev_states.GetLastDirection())
					}
				}
				if diffElevStateStructs(Old_ElevState, StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)) {
					Old_ElevState = StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
					ch_elevstate <- Old_ElevState
				}
				ch_cab_lamps <- Active_orders.GetCabRequests()
			} else {
				ch_new_hall_req <- elevio.ButtonEvent{Floor: btn_press.Floor, Button: btn_press.Button}
			}
		case initial_cab_requests := <-ch_init_cab_requests:
			// TODO: Implement
			fmt.Println("HandleInitCabRequests", initial_cab_requests[0]) // Just to ignore it during testing
		}
	}
}
