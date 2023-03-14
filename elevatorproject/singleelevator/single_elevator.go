package singleelevator 

import (
    "elevatorproject/singleelevator/doortimer"
    "elevatorproject/singleelevator/elevio"
    "fmt"
    "math/rand"
    "time"
)

func RunSingleElevator(){

    numFloors := 4

    elevio.Init("10.100.23.27:15657", numFloors) 
    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)   
	ch_door := make(chan int)
    ch_hra := make(chan [][2]bool)
    ch_init_cab_requests := make(chan []bool)
    // ch_cab_requests := make(chan []bool)
    ch_completed_hall_requests := make(chan elevio.ButtonEvent)
    ch_new_hall_requests := make(chan elevio.ButtonEvent)
    ch_elevstate := make(chan ElevState)
    // Newest channels
    // ch_mechanical_error := make(chan bool)
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go doortimer.CheckTimer(ch_door)

    go testMakeHRA(ch_hra)
    go testReceiveStuff(ch_completed_hall_requests, ch_new_hall_requests, ch_elevstate)
	
    Fsm_elevator(drv_buttons, drv_floors, ch_door, ch_hra, ch_init_cab_requests, ch_completed_hall_requests, ch_new_hall_requests, ch_elevstate)
}

func testMakeHRA(orders_to_send chan [][2]bool){
    for {
        time.Sleep(2*time.Second)
        hra_orders := make([][2]bool, 4)
        for i := 0; i < 4; i++ {
            for j := 0; j < 2; j++ {
                if !(i == 0 && j == 1) && !(i == 3 && j == 0) {
                    hra_orders[i][j] = rand.Intn(2) == 1
                }
            }
        }
        fmt.Println("HRA orders: ", hra_orders)
        orders_to_send <- hra_orders
    }
}

func testReceiveStuff(ch_completed_hall_requests chan elevio.ButtonEvent, ch_new_hall_requests chan elevio.ButtonEvent, ch_elevstate chan ElevState){
    for {
        select {
        case completed_hall_request := <-ch_completed_hall_requests:
            fmt.Println("Completed hall request: ", completed_hall_request)
        case new_hall_request := <-ch_new_hall_requests:
            fmt.Println("New hall request: ", new_hall_request)
        case elevstate := <-ch_elevstate:
            fmt.Println("Elevator state: ", elevstate)
        }
    }
}
