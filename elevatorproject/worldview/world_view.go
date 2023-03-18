package worldview

import (
	"elevatorproject/config"
	"elevatorproject/network/peers"
	"elevatorproject/nodeview"
	"elevatorproject/singleelevator"
	"fmt"
)

type PeersAlive []string

type MyWorldView struct {
	ElevStates     map[string]singleelevator.ElevState
	HallRequestViews   map[string][][2]nodeview.RequestState
	NodesAvailable map[string]bool
}

// Function that returns false if nodeID (input) is not in peersALive
func (peersAlive PeersAlive) IsPeerAlive(nodeID string) bool {
	for _, peer := range peersAlive {
		if peer == nodeID {
			return true
		}
	}
	return false
}

func (myWorldView *MyWorldView) initMyWorldView() {
	myWorldView.ElevStates = make(map[string]singleelevator.ElevState, config.NumElevators)
	myWorldView.HallRequestViews = make(map[string][][2]nodeview.RequestState, config.NumElevators)
	myWorldView.NodesAvailable = make(map[string]bool, config.NumElevators)
	myWorldView.NodesAvailable[config.LocalID] = true
}

func printNodeView(node nodeview.MyNodeView) {
	fmt.Printf("ID: %s\n", node.ID)
	fmt.Printf("IsAvailable: %v\n", node.IsAvailable)
	fmt.Printf("ElevState: Behaviour=%s Floor=%d Direction=%s CabRequests=%v IsAvailable=%v\n",
		node.ElevState.Behaviour, node.ElevState.Floor, node.ElevState.Direction, node.ElevState.CabRequests, node.ElevState.IsAvailable)
	fmt.Printf("HallRequests:\n")
	for i, requests := range node.HallRequests {
		fmt.Printf("  Floor %d: Up=%v Down=%v\n", i+1, requests[0], requests[1])
	}
	fmt.Printf("CabRequests:\n")
	for id, requests := range node.RemoteCabRequests {
		fmt.Printf("  Cab %s: %v\n", id, requests)
	}
}

func WorldView(ch_receiveNodeView <-chan nodeview.MyNodeView,
	ch_receivePeerUpdate <-chan peers.PeerUpdate,
	ch_setTransmitEnable chan <- bool,
	ch_initCabRequests chan <- []bool,
	ch_remoteRequestView chan <- nodeview.RemoteRequestView,
	ch_hraInput chan<- MyWorldView) {


	var myWorldView MyWorldView
	var peersAlive PeersAlive
	var remoteRequestView nodeview.RemoteRequestView

	myWorldView.initMyWorldView()
	remoteRequestView.InitRemoteRequestView()

	for {
		select {
		case peerUpdate := <-ch_receivePeerUpdate:
			fmt.Println("worldview: peerUpdate")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)

			peersAlive = peerUpdate.Peers
			peersLost := peerUpdate.Lost
		
			for _, lostPeer := range peersLost {
				// If this node can be found in lostPeer, we should delete it from the systemAwareness
				if lostPeer != config.LocalID {
					delete(myWorldView.NodesAvailable, lostPeer)
					delete(myWorldView.ElevStates, lostPeer)
					delete(myWorldView.HallRequestViews, lostPeer)

					delete(remoteRequestView.RemoteHallRequestViews, lostPeer)
					delete(remoteRequestView.RemoteCabRequests, lostPeer)
				}
			}
			// Here I can add if I am in an init state, I should send cab call of LocalID on channel init_cab_requests
			// This will be done in the init state of the elevator
		case nodeView := <-ch_receiveNodeView:
			fmt.Println("worldview: nodeView")
			fmt.Println("Received from ", nodeView.ID)

			nodeID := nodeView.ID

			// Break out of case if IsPeerAlive returns false
			if !peersAlive.IsPeerAlive(nodeID) {
				break
			}

			myWorldView.NodesAvailable[nodeID] = nodeView.IsAvailable
			myWorldView.ElevStates[nodeID] = nodeView.ElevState
			myWorldView.HallRequestViews[nodeID] = nodeView.HallRequests

			if nodeID != config.LocalID {
				remoteRequestView.RemoteHallRequestViews[nodeID] = nodeView.HallRequests
				remoteRequestView.RemoteCabRequests[nodeID] = nodeView.ElevState.CabRequests
			}

			ch_hraInput <- myWorldView
			ch_remoteRequestView <- remoteRequestView
		default:

		}
	}
}
