package singleelevator

import (
	"elevatorproject/singleelevator/doortimer"
	"elevatorproject/singleelevator/elevio"
	"fmt"
)

// Helper struct
type ElevState struct {
    Behavior    string     
    Floor       int          
    Direction   string      
    CabRequests []bool   
	isAvailable bool   
}

// Constants
const n_floors int = 4

// Helper functions
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

// type ElevState struct {
//     Behavior    string     
//     Floor       int          
//     Direction   string      
//     CabRequests []bool      
//	   isAvailable bool
// }

func StatesToHRAStates(states States, cab_requests []bool, isAvailable bool) ElevState {
	hra_states := ElevState{}

	// Floor
	if states.GetLastFloor() == -1 {
		hra_states.Floor = n_floors/2
	} else {
		hra_states.Floor = states.GetLastFloor()
	}

	// Direction
	switch (states.GetLastDirection()) {
	case elevio.MD_Up:
		hra_states.Direction = "up"
	case elevio.MD_Down:
		hra_states.Direction = "down"
	}

	// Behaviour
	switch (states.GetElevatorBehaviour()) {
	case "Idle":
		hra_states.Behavior = "idle"
		hra_states.Direction = "stop"
	case "Moving":
		hra_states.Behavior = "moving"
	case "DoorOpen":
		hra_states.Behavior = "doorOpen"
		hra_states.Direction = "stop"
	default:
		panic("Invalid elevator behaviour")
	}

	// Cab requests
	hra_states.CabRequests = cab_requests

	// isAvailable
	hra_states.isAvailable = isAvailable

	return hra_states
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
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{floor, elevio.BT_Cab})
		}
		if active_orders.GetSpecificOrder(floor, DirToBtnType(elev_states.GetLastDirection())) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{floor, DirToBtnType(elev_states.GetLastDirection())})
		}
		if floor == 0 && active_orders.GetSpecificOrder(floor, elevio.BT_HallUp) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{floor, elevio.BT_HallUp})
		}
		if floor == n_floors-1 && active_orders.GetSpecificOrder(floor, elevio.BT_HallDown) {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{floor, elevio.BT_HallDown})
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
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{elev_states.GetLastFloor(), elevio.BT_Cab})
		}
		if (up_this_floor && elev_states.GetLastDirection() == elevio.MD_Up) ||
			(elev_states.GetLastFloor() == 0 && up_this_floor) {
			open_door_bool = true
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{elev_states.GetLastFloor(), elevio.BT_HallUp})
		}
		if (down_this_floor && elev_states.GetLastDirection() == elevio.MD_Down) ||
			(elev_states.GetLastFloor() == n_floors-1 && down_this_floor) {
			open_door_bool = true
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{elev_states.GetLastFloor(), elevio.BT_HallDown})
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
					remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{elev_states.GetLastFloor(), elevio.BT_Cab})
				}
				if (up_this_floor && elev_states.GetLastDirection() == elevio.MD_Up) ||
					(up_this_floor && !down_this_floor && !active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elevio.MD_Down)) ||
					(elev_states.GetLastFloor() == 0 && up_this_floor) {
					open_door_bool = true
					elev_states.SetElevatorBehaviour("DoorOpen")
					remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{elev_states.GetLastFloor(), elevio.BT_HallUp})
				}
				if (down_this_floor && elev_states.GetLastDirection() == elevio.MD_Down) ||
					(down_this_floor && !up_this_floor && !active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elevio.MD_Up)) ||
					(elev_states.GetLastFloor() == n_floors-1 && down_this_floor) {
					open_door_bool = true
					elev_states.SetElevatorBehaviour("DoorOpen")
					remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{elev_states.GetLastFloor(), elevio.BT_HallDown})
				}

			}
			if !order_in_this_floor || !open_door_bool {
				set_direction_bool = true
				elev_states.SetElevatorBehaviour("Moving")
				if (elev_states.GetLastDirection() == elevio.MD_Up && orders_above) || !orders_below {
					elev_states.SetDirection(elevio.MD_Up)
					// fmt.Println("Moving up")
				} else if (elev_states.GetLastDirection() == elevio.MD_Down && orders_below) || !orders_above {
					elev_states.SetDirection(elevio.MD_Down)
					// fmt.Println("Moving down")
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
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{elev_states.GetLastFloor(), elevio.BT_HallUp})
		} else {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{elev_states.GetLastFloor(), elevio.BT_HallDown})
		}
		return elev_states, active_orders, open_door_bool, set_direction_bool, remove_orders_list
	} else if active_orders.OrderInFloor(elev_states.GetLastFloor()) && !active_orders.AnyOrderPastFloorInDir(elev_states.GetLastFloor(), elev_states.GetLastDirection()) {
		open_door_bool = true
		if elev_states.GetLastDirection() == elevio.MD_Up {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{elev_states.GetLastFloor(), elevio.BT_HallDown})
		} else {
			remove_orders_list = append(remove_orders_list, elevio.ButtonEvent{elev_states.GetLastFloor(), elevio.BT_HallUp})
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

func testSetLights(orders Orders) {
	for floor := 0; floor < n_floors; floor++ {
		if orders.Cab_orders[floor] {
			elevio.SetButtonLamp(elevio.BT_Cab, floor, true)
		} else {
			elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
		}
		if orders.Up_orders[floor] {
			elevio.SetButtonLamp(elevio.BT_HallUp, floor, true)
		} else {
			elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		}
		if orders.Down_orders[floor] {
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, true)
		} else {
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
		}
	}
}

func Fsm_elevator(ch_btn <-chan elevio.ButtonEvent,
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

	// Initiate elevator
	fmt.Println("Initiate elevator")
	Active_orders.InitOrders()
	InitLamps(Active_orders) //TODO: remove this because of fix by channel
	Elev_states.InitStates()
	if elevio.GetFloor() == -1 {
		Elev_states.SetElevatorBehaviour("Moving")
		elevio.SetMotorDirection(elevio.MD_Up)
	} else {
		Elev_states.SetLastFloor(elevio.GetFloor())
	}
	ch_elevstate <- StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)

	// Finite state machine
	for {
		select {
		case hra := <-ch_hra:
			fmt.Println("HandleHRA")
			var open_door_bool, set_direction_bool bool
			var remove_orders_list []elevio.ButtonEvent
			Elev_states, Active_orders, open_door_bool, set_direction_bool, remove_orders_list = HandleNewRequests(hra, -1, Elev_states, Active_orders)
			testSetLights(Active_orders) // Just for testing
			ch_elevstate <- StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
			if open_door_bool {
				elevio.SetDoorOpenLamp(true)
				doortimer.StartTimer()
				// TODO: For loop to remove orders
				for _, v := range remove_orders_list {
					if v.Button == elevio.BT_Cab {
						Active_orders.SetOrder(v.Floor, v.Button, false)
						ch_elevstate <- StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
						UpdateCabLamps(Active_orders)
					} else {
						ch_completed_hall_req <- v
					}

				}
			} else {
				if set_direction_bool {
					elevio.SetMotorDirection(Elev_states.GetLastDirection())
				}
			}
		case floor := <-ch_floor:
			fmt.Println("HandleFloorSensor")
			elevio.SetFloorIndicator(floor)
			var stop_bool bool
			var remove_orders_list []elevio.ButtonEvent
			Elev_states, Active_orders, stop_bool, remove_orders_list = HandleFloorSensor(floor, Elev_states, Active_orders)
			testSetLights(Active_orders) // Just for testing
			ch_elevstate <- StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
			if stop_bool {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)
				doortimer.StartTimer() // maybe change to use golang timer

				// TODO: For loop to remove orders
				for _, v := range remove_orders_list {
					if v.Button == elevio.BT_Cab {
						Active_orders.SetOrder(v.Floor, v.Button, false)
						ch_elevstate <- StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
						UpdateCabLamps(Active_orders)
					} else {
						ch_completed_hall_req <- v
					}
				}
			}
		case <-ch_door:
			testSetLights(Active_orders) // Just for testing
			fmt.Println("HandleDoorClosing")
			if elevio.GetObstruction() {
				doortimer.StartTimer()
				break
			}
			var open_door_bool, set_direction_bool bool
			var remove_orders_list []elevio.ButtonEvent
			Elev_states, Active_orders, open_door_bool, set_direction_bool, remove_orders_list = HandleDoorClosing(Elev_states, Active_orders)
			testSetLights(Active_orders) // Just for testing
			ch_elevstate <- StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
			if open_door_bool {
				elevio.SetDoorOpenLamp(true)
				doortimer.StartTimer()
				for _, v := range remove_orders_list {
					if v.Button == elevio.BT_Cab {
						Active_orders.SetOrder(v.Floor, v.Button, false)
						ch_elevstate <- StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
						UpdateCabLamps(Active_orders)
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
		case btn_press := <-ch_btn:
			testSetLights(Active_orders) // Just for testing
			fmt.Println("HandleButtonEvent")
			if btn_press.Button == elevio.BT_Cab {
				var open_door_bool, set_direction_bool bool
				var remove_orders_list []elevio.ButtonEvent
				emtpy_hra := [][2]bool{}
				Elev_states, Active_orders, open_door_bool, set_direction_bool, remove_orders_list = HandleNewRequests(emtpy_hra, btn_press.Floor, Elev_states, Active_orders)
				testSetLights(Active_orders) // Just for testing
				ch_elevstate <- StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
				if open_door_bool {
					elevio.SetDoorOpenLamp(true)
					doortimer.StartTimer()
					for _, v := range remove_orders_list {
						if v.Button == elevio.BT_Cab {
							Active_orders.SetOrder(v.Floor, v.Button, false)
							ch_elevstate <- StatesToHRAStates(Elev_states, Active_orders.GetCabRequests(), isAvailable)
							UpdateCabLamps(Active_orders)
						} else {
							ch_completed_hall_req <- v
						}
					}
				} else {
					if set_direction_bool {
						elevio.SetMotorDirection(Elev_states.GetLastDirection())
					}
				}
			} else {
				ch_new_hall_req <- elevio.ButtonEvent{btn_press.Floor, btn_press.Button}

			}
		case initial_cab_requests := <-ch_init_cab_requests:
			// TODO: Implement
			fmt.Println("HandleInitCabRequests", initial_cab_requests[0]) // Just to ignore it during testing
		}
	}
}
