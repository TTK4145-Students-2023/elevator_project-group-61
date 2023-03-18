package network

import (
	"elevatorproject/network/bcast"
	"elevatorproject/network/peers"
	"elevatorproject/nodeview"
	"fmt"
)

// Network function going to take in ch_transmit, ch_receive, ch_peerUpdate, ch_peerTransmitEnable

func ReceiveNetwork(ch_receiveNodeView chan <- nodeview.MyNodeView,
	ch_receivePeerUpdate chan <- peers.PeerUpdate) {
		ch_receive := make(chan nodeview.MyNodeView)
		ch_peerUpdate := make(chan peers.PeerUpdate)

		go peers.Receiver(13200, ch_peerUpdate)
		go bcast.Receiver(12100, ch_receive)

		for {
			select {
			case nodeview := <- ch_receive:
				fmt.Println("receive_network: Mottar en node view")
				ch_receiveNodeView <- nodeview
			
			case peerUpdate := <- ch_peerUpdate:
				ch_receivePeerUpdate <- peerUpdate
			//default:

			}
		}
	}