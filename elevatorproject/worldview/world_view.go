package worldview

import (
	"elevatorproject/config"
	"elevatorproject/network/peers"
	"elevatorproject/peerview"
	"elevatorproject/singleelevator"
	"fmt"
)

type PeersAlive []string

type MyWorldView struct {
	ElevStates      map[string]singleelevator.ElevState
	HallRequestView [config.NumFloors][2]peerview.RequestState
	CabRequests     map[string][config.NumFloors]peerview.RequestState
}

func copyMyWorldView(myWorldView MyWorldView) MyWorldView {
	var copyWorldView MyWorldView
	copyWorldView.ElevStates = make(map[string]singleelevator.ElevState, config.NumElevators)
	copyWorldView.CabRequests = make(map[string][config.NumFloors]peerview.RequestState, config.NumElevators)

	for key, value := range myWorldView.ElevStates {
		copyWorldView.ElevStates[key] = value
	}
	for key, value := range myWorldView.CabRequests {
		copyWorldView.CabRequests[key] = value
	}
	copyWorldView.HallRequestView = myWorldView.HallRequestView
	return copyWorldView
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

func (myWorldView *MyWorldView) initMyWorldView(localID string) {
	myWorldView.ElevStates = make(map[string]singleelevator.ElevState, config.NumElevators)
	myWorldView.HallRequestView = [config.NumFloors][2]peerview.RequestState{}

	var elevState singleelevator.ElevState
	elevState.InitElevState()
	myWorldView.ElevStates[localID] = elevState

	// init cab requests
	myWorldView.CabRequests = make(map[string][config.NumFloors]peerview.RequestState, config.NumElevators)
	// TODO: Sjekke om denne er nødvendig
	myWorldView.CabRequests[localID] = [config.NumFloors]peerview.RequestState{}
}

func WorldView(ch_receiveNodeView <-chan peerview.MyPeerView,
	ch_receivePeerUpdate <-chan peers.PeerUpdate,
	ch_remoteRequestView chan<- peerview.RemoteRequestView,
	ch_hraInput chan<- MyWorldView,
	ch_singleElevMode chan<- bool,
	localID string) {

	var myWorldView MyWorldView
	var peersAlive PeersAlive
	var remoteRequestView peerview.RemoteRequestView
	var isSingleElevMode bool

	myWorldView.initMyWorldView(localID)
	remoteRequestView.InitRemoteRequestView()
	fmt.Println(" ")

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
				if lostPeer != localID {
					delete(myWorldView.ElevStates, lostPeer)

					delete(remoteRequestView.RemoteHallRequestViews, lostPeer)
					delete(remoteRequestView.RemoteCabRequestViews, lostPeer)
				}
			}
			//TODO: Må undersøke om denne checken er nok, og undersøke om ch_isSingleElevMode kan
			// tas i bruk av nodeview.
			if len(peersAlive) <= 1 {
				ch_singleElevMode <- true
				isSingleElevMode = true
				ch_hraInput <- copyMyWorldView(myWorldView)
				ch_remoteRequestView <- peerview.CopyRemoteRequestView(remoteRequestView)
			} else {
				ch_singleElevMode <- false
				isSingleElevMode = false
			}
		case nodeView := <-ch_receiveNodeView:
			//fmt.Println("worldview: nodeView")
			fmt.Println("Received from ", nodeView.ID)

			nodeID := nodeView.ID

			// Break out of case if IsPeerAlive returns false
			if !peersAlive.IsPeerAlive(nodeID) && localID != nodeID {
				break
			}

			myWorldView.ElevStates[nodeID] = nodeView.ElevState
			myWorldView.CabRequests[nodeID] = nodeView.CabRequests[nodeID]

			if nodeID != localID {
				remoteRequestView.RemoteHallRequestViews[nodeID] = nodeView.HallRequests
				remoteRequestView.RemoteCabRequestViews[nodeID] = nodeView.CabRequests
			} else {
				myWorldView.HallRequestView = nodeView.HallRequests
			}

			ch_remoteRequestView <- peerview.CopyRemoteRequestView(remoteRequestView)
			if !isSingleElevMode {
				ch_hraInput <- copyMyWorldView(myWorldView)
			}
		}
	}
}
