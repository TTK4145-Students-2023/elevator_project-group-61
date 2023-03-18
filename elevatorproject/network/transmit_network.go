package network

import (
	"elevatorproject/config"
	"elevatorproject/network/bcast"
	"elevatorproject/network/peers"
	"elevatorproject/nodeview"
	"fmt"
)


func TransmitNetwork(ch_sendMyNodeView <- chan nodeview.MyNodeView, ch_setTransmitEnable <- chan bool) {
	ch_transmit := make(chan nodeview.MyNodeView)
	ch_peerTransmitEnable := make(chan bool)

	go peers.Transmitter(13200, config.LocalID, ch_peerTransmitEnable)
	go bcast.Transmitter(12100, ch_transmit)

	for {
		select {
		case myNodeView := <- ch_sendMyNodeView:
			fmt.Println("transmit_network: sending my nodeview")
			ch_transmit <- myNodeView
		case setTransmitEnable := <- ch_setTransmitEnable:
			ch_peerTransmitEnable <- setTransmitEnable
		//default:
			//time.Sleep(100*time.Millisecond)

		}
	}
}