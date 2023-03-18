package network

import (
	"elevatorproject/network/bcast"
	"elevatorproject/network/peers"
	"elevatorproject/nodeview"
	"fmt"
	"time"
)

// Network function going to take in ch_transmit, ch_receive, ch_peerUpdate, ch_peerTransmitEnable

func ReceiveNetwork(ch_receiveNodeView chan <- nodeview.MyNodeView,
	ch_receivePeerUpdate chan <- peers.PeerUpdate) {
		

		

		for {
			select {
			case nodeview := <- ch_receive:
				fmt.Println("receive_network: Mottar en node view")
				ch_receiveNodeView <- nodeview
			
			case peerUpdate := <- ch_peerUpdate:
				ch_receivePeerUpdate <- peerUpdate
			//default:
				//time.Sleep(100*time.Millisecond)

			}
			time.Sleep(100*time.Millisecond)
		}
	}