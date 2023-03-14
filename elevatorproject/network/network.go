package network

import (
	"elevatorproject/systemview"
	"elevatorproject/network/peers"
	"elevatorproject/config"
	"elevatorproject/network/bcast"
)

// Network function going to take in ch_transmit, ch_receive, ch_peerUpdate, ch_peerTransmitEnable

func Network(
	ch_sendNodeAwareness <- chan systemview.NodeAwareness,
	ch_receiveNodeAwareness chan <- systemview.NodeAwareness,
	ch_receivePeerUpdate chan <- peers.PeerUpdate,
	ch_setTransmitEnalble <- chan bool) {
	
	ch_transmit := make(chan systemview.NodeAwareness)
	ch_receive := make(chan systemview.NodeAwareness)
	ch_peerTransmitEnable := make(chan bool)
	ch_peerUpdate := make(chan peers.PeerUpdate)

	go peers.Transmitter(15647, config.LocalID, ch_peerTransmitEnable)
	go peers.Receiver(15647, ch_peerUpdate)
	go bcast.Receiver(16569, ch_receive)
	go bcast.Transmitter(16569, ch_transmit)

	for {
		select {
		case myNodeAwareness := <- ch_sendNodeAwareness:
			ch_transmit <- myNodeAwareness
		case nodeAwareness := <- ch_receive:
			ch_receiveNodeAwareness <- nodeAwareness
		case peerUpdate := <- ch_peerUpdate:
			ch_receivePeerUpdate <- peerUpdate
		case setTransitEnable := <- ch_setTransmitEnalble:
			ch_peerTransmitEnable <- setTransitEnable
		}
	}	
}


