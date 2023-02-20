package main

import (

)

type GlobalOrders struct {
	Up_orders   []int
	Down_orders []int
	Cab_orders  []int
}

func (states *GlobalOrders) InitGlobalOrders() {
	states.Cab_orders = make([]int, n_floors)
	states.Up_orders = make([]int, n_floors)
	states.Down_orders = make([]int, n_floors)
	for i := 0; i < n_floors; i++ {
		states.Cab_orders[i] = 0
		states.Up_orders[i] = 0
		states.Down_orders[i] = 0
	}
}