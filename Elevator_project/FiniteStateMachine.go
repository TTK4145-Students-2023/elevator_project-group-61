package main

import (
	"ElevatorProject/door_timer"
	"ElevatorProject/elevio"
	"fmt"
)

// Elevator needs to react to:
// - Received order
// - Floor reached
// - Door timer ending
// - Not being initiated
// I can add stop and obstruction later

// ####################
// TYPES AND VARIABLES
// ####################

type States struct {
	// If this is changed, remember to change:
	// - InitStates()
	last_floor int
	last_direction elevio.MotorDirection
	door_open bool
}

var Elevator_states States

type Orders struct {
	Up_orders   [4]bool
	Down_orders [4]bool
	Cab_orders  [4]bool
}

var Current_orders Orders

// ####################
// FUNCTIONS
// ####################
func InitStates() {
	Elevator_states.last_floor = -1
	Elevator_states.last_direction = elevio.MD_Stop
	Elevator_states.door_open = false
}

func InitOrders() {
	Current_orders = Orders{
		[4]bool{false, false, false, false},
		[4]bool{false, false, false, false},
		[4]bool{false, false, false, false}}
}

func InitLamps() {
	elevio.SetDoorOpenLamp(false)
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, false)
		}
	}
}

func init_elevator() {
	// Initiate this elevator, that means:
	// - Turning off all btn lamps, turning off door lamp(close door)
	// - Resetting the state
	// - Resetting orders
	// - Make sure elevator is in a floor
	InitOrders()
	InitLamps()
	InitStates()

	// Make elevator move if it is not in a floor (it doesnt know where it is then)
	if elevio.GetFloor() == -1 {
		elevio.SetMotorDirection(elevio.MD_Up)
		Elevator_states.last_direction = elevio.MD_Up
		// Then I assume that when elevator hits a floor it will handle it from there
	} else {
		Elevator_states.last_floor = elevio.GetFloor()
		elevio.SetFloorIndicator(Elevator_states.last_floor)
	}
}

// MINI FUNCTIONS START ################### MINI FUNCTIONS START ###################
func any_orders() bool {
	for i := 0; i < 4; i++ {
		if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
			return true
		}
	}
	return false
}

func any_orders_past_this_floor_in_direction(floor int, dir elevio.MotorDirection) bool {
	switch dir {
	case elevio.MD_Up:
		// Det er ingen ordre forbi hvis dette er maks etasjen
		if floor == 3 {
			return false
		}
		for i := floor + 1; i < 4; i++ {
			if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
				return true
			}
		}
		// hvis det ikke er funnet noen ordre oppover forbi etasjen, finnes det ingen ordre forbi
		return false
	case elevio.MD_Down:
		if floor == 0 {
			return false
		}
		for i := floor - 1; i > -1; i-- {
			if Current_orders.Cab_orders[i] || Current_orders.Down_orders[i] || Current_orders.Up_orders[i] {
				return true
			}
		}
		return false
	}
	fmt.Println("Dette skal ikke skje, da har any orders past blitt kallet med elevator_last_direction som stop.")
	return false
}

func elevator_should_stop_after_sensing_floor(floor int) bool {
	// Antatt at Elevator_states.last_direction er opp eller ned, ikke stille
	// Det er antatt at heisen er i en etasje
	// det er antatt at heisen nettopp er kommet til en etasje (skal brukes i floor_sensored)
	// elevator should stop if there are no orders
	if !any_orders() {
		// hvis det ikke er noen ordre, stop
		return true
	}
	// hvis det er en cab order i etasjen, stop uansett
	if Current_orders.Cab_orders[floor] {
		return true
	}
	switch Elevator_states.last_direction {
	case elevio.MD_Up:
		// hvis heisen gikk opp, og det ikke er ordre videre i denne retningen etter floor, eller
		// ordren i denne etasjen faktisk er en "opp-ordre", stopp
		if Current_orders.Up_orders[floor] || !any_orders_past_this_floor_in_direction(floor, elevio.MD_Up) {
			return true
		}
		return false
	case elevio.MD_Down:
		// tilsvarende med ned
		if Current_orders.Down_orders[floor] || !any_orders_past_this_floor_in_direction(floor, elevio.MD_Down) {
			return true
		}
		return false
	}
	fmt.Println("elevator should stop - det finnes ordre, men enten så er last direction stop eller hmm.. videre")
	return false
}

func remove_orders_and_btn_lights(floor int) {
	// antar at floor er legit
	// fjerner ordre
	Current_orders.Cab_orders[floor] = false
	Current_orders.Down_orders[floor] = false
	Current_orders.Up_orders[floor] = false
	//skrur av lys
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
}

func add_order_to_system(btn elevio.ButtonEvent) {
	elevio.SetButtonLamp(btn.Button, btn.Floor, true)
	switch btn.Button {
	case elevio.BT_HallUp:
		Current_orders.Up_orders[btn.Floor] = true
	case elevio.BT_Cab:
		Current_orders.Cab_orders[btn.Floor] = true
	case elevio.BT_HallDown:
		Current_orders.Down_orders[btn.Floor] = true
	}
}

// MINI FUNCTIONS STOP  ################### MINI FUNCTIONS STOP  ###################

func HandleFloorSensor(floor int) {
	// Skal uansett skru på floor light, og sette last floor
	elevio.SetFloorIndicator(floor)
	Elevator_states.last_floor = floor
	// Antar at heisen er i en etasje, tar altså ikke hensyn til array index out of range
	// If a floor is hit then:
	// - Check if elevator should stop
	if elevator_should_stop_after_sensing_floor(floor) {
		// heisen skal stoppe, MEN JEG SETTER IKKE LAST DIRECTION TIL STOP
		elevio.SetMotorDirection(elevio.MD_Stop)
		// Door må åpnes - altså lyset må skrus på
		elevio.SetDoorOpenLamp(true)
		Elevator_states.door_open = true
		// timer må startes for døren
		door_timer.StartTimer()
		// ordre i denne etasjen må fjernes og lys må fjernes fra knappene i denne etasjen
		remove_orders_and_btn_lights(floor)
	}
}

func HandleNewOrder(new_order elevio.ButtonEvent) {
	if elevio.GetFloor() == -1 {
		add_order_to_system(new_order)
		return
	}
	if new_order.Floor == elevio.GetFloor() {
		elevio.SetDoorOpenLamp(true)
		Elevator_states.door_open = true
		door_timer.StartTimer()
		return
	}
	switch Elevator_states.door_open {
	case false:
		add_order_to_system(new_order)
		if new_order.Floor-elevio.GetFloor() > 0 {
			elevio.SetMotorDirection(elevio.MD_Up)
			Elevator_states.last_direction = elevio.MD_Up
		} else {
			elevio.SetMotorDirection(elevio.MD_Down)
			Elevator_states.last_direction = elevio.MD_Down
		}
		return
	case true:
		// altså er døren åpen og ordren er ikke i denne etasjen:
		// - bare legge til ordren i systemet (og regne med at nå døren lukker seg behandles det)
		add_order_to_system(new_order)
		return
	}
}

func HandleDoorClosing() {
	// Dette er også det eneste stedet jeg tar hensyn til obstruksjon
	// Dersom det er obstruksjon, restart timer og return
	if elevio.GetObstruction() {
		door_timer.StartTimer()
		return
	}
	// Må stenge dører og alt med det å gjøre:
	elevio.SetDoorOpenLamp(false)
	Elevator_states.door_open = false
	// hvis det ikke er noen ordre skal ingenting gjøres
	if !any_orders() {
		return
	}
	// hvis heisen var på vei opp og det er ordre over denne etasjen, gå opp:
	orders_up := any_orders_past_this_floor_in_direction(Elevator_states.last_floor, elevio.MD_Up)
	orders_down := any_orders_past_this_floor_in_direction(Elevator_states.last_floor, elevio.MD_Down)
	if Elevator_states.last_direction == elevio.MD_Up && orders_up {
		elevio.SetMotorDirection(elevio.MD_Up)
		return
	}
	// og motsatt sjekkes:
	if Elevator_states.last_direction == elevio.MD_Down && orders_down {
		elevio.SetMotorDirection(elevio.MD_Down)
		return
	}
	// ellers er ingen av de tilfellene, men det er fremdeles en ordre i systemet, gå i retning av ordren
	// det betyr at det må være i motsatt retning av last direction
	if Elevator_states.last_direction == elevio.MD_Up {
		elevio.SetMotorDirection(elevio.MD_Down)
		Elevator_states.last_direction = elevio.MD_Down
		return
	}
	// motsatt
	elevio.SetMotorDirection(elevio.MD_Up)
	Elevator_states.last_direction = elevio.MD_Up
	// return
}

func Fsm_elevator(ch_order chan elevio.ButtonEvent, ch_floor chan int, ch_door chan int) {
	// Iniate elevator
	init_elevator()

	for {
		select {
		case floor := <-ch_floor:
			fmt.Println("HandleFloorSensor")
			HandleFloorSensor(floor)
		case new_order := <-ch_order:
			fmt.Println("HandleNewOrder")
			HandleNewOrder(new_order)
		case <-ch_door:
			fmt.Println("HandleDoorClosing")
			HandleDoorClosing()
		}
	}
}
