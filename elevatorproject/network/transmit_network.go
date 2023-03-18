package network

import (
	"elevatorproject/config"
	"elevatorproject/network/bcast"
	"elevatorproject/network/peers"
	"elevatorproject/nodeview"
	// "fmt"
	"time"
)


func TransmitNetwork(ch_sendMyNodeView <- chan nodeview.MyNodeView, ch_setTransmitEnable <- chan bool) {
	
	for {
		select {
		case myNodeView := <- ch_sendMyNodeView:
			ch_transmit <- myNodeView
		case setTransmitEnable := <- ch_setTransmitEnable:
			ch_peerTransmitEnable <- setTransmitEnable
		//default:
			//time.Sleep(100*time.Millisecond)

		}
		time.Sleep(100*time.Millisecond)
	}
}