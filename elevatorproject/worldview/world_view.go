package worldview

import (
	"elevatorproject/config"
	"elevatorproject/network/peers"
	"elevatorproject/peerview"
	"elevatorproject/singleelevator"
	"fmt"
)

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

func isPeerAlive(peerID string, peersAlive []string) bool {
	for _, peer := range peersAlive {
		if peer == peerID {
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

	myWorldView.CabRequests = make(map[string][config.NumFloors]peerview.RequestState, config.NumElevators)
	myWorldView.CabRequests[localID] = [config.NumFloors]peerview.RequestState{}
}

func hasMyWorldViewChanged(myWorldView MyWorldView, prevMyWorldView MyWorldView) bool {
	for id, elevstate := range myWorldView.ElevStates {
		if elevstate != prevMyWorldView.ElevStates[id] {
			return true
		}
	}

	for floor := 0; floor < config.NumFloors; floor++ {
		for button := 0; button < 2; button++ {
			if myWorldView.HallRequestView[floor][button] != prevMyWorldView.HallRequestView[floor][button] {
				return true
			}
		}
	}

	for id, requestStates := range myWorldView.CabRequests {
		for floor := 0; floor < config.NumFloors; floor++ {
			if requestStates[floor] != prevMyWorldView.CabRequests[id][floor] {
				return true
			}
		}
	}
	return false
}

func WorldView(ch_receive <-chan peerview.MyPeerView,
	ch_peerUpdate <-chan peers.PeerUpdate,
	ch_remoteRequestView chan<- peerview.RemoteRequestViews,
	ch_myWorldView chan<- MyWorldView,
	ch_singleElevMode chan<- bool,
	localID string) {

	var myWorldView MyWorldView
	var prevMyWorldView MyWorldView
	var peersAlive []string
	var remoteRequestView peerview.RemoteRequestViews
	var prevRemoteRequestView peerview.RemoteRequestViews
	var isSingleElevMode bool

	myWorldView.initMyWorldView(localID)
	prevMyWorldView.initMyWorldView(localID)
	remoteRequestView.InitRemoteRequestViews()
	prevRemoteRequestView.InitRemoteRequestViews()

	for {
		select {
		case peerUpdate := <-ch_peerUpdate:
			fmt.Println("worldview: peerUpdate")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)

			peersAlive = peerUpdate.Peers
			peersLost := peerUpdate.Lost

			for _, lostPeer := range peersLost {
				if lostPeer != localID {
					delete(myWorldView.ElevStates, lostPeer)
					delete(myWorldView.CabRequests, lostPeer)

					delete(remoteRequestView.RemoteHallRequestViews, lostPeer)
					delete(remoteRequestView.RemoteCabRequestViews, lostPeer)
				}
			}
			
			if len(peersAlive) <= 1 {
				isSingleElevMode = true
				ch_singleElevMode <- true
				ch_myWorldView <- copyMyWorldView(myWorldView)
				ch_remoteRequestView <- peerview.CopyRemoteRequestViews(remoteRequestView)

			} else {
				isSingleElevMode = false
				ch_singleElevMode <- isSingleElevMode
			}
		case peerView := <-ch_receive:
			peerID := peerView.ID

			if !isPeerAlive(peerID, peersAlive) && localID != peerID {
				break
			}

			myWorldView.ElevStates[peerID] = peerView.ElevState
			myWorldView.CabRequests[peerID] = peerView.CabRequests[peerID]

			if peerID != localID {
				remoteRequestView.RemoteHallRequestViews[peerID] = peerView.HallRequests
				remoteRequestView.RemoteCabRequestViews[peerID] = peerView.CabRequests

			} else {
				myWorldView.HallRequestView = peerView.HallRequests
			}

			if peerview.HasRemoteRequestViewsChanged(remoteRequestView, prevRemoteRequestView) {
				ch_remoteRequestView <- peerview.CopyRemoteRequestViews(remoteRequestView)
				prevRemoteRequestView = peerview.CopyRemoteRequestViews(remoteRequestView)
			}

			if hasMyWorldViewChanged(myWorldView, prevMyWorldView) {
				ch_myWorldView <- copyMyWorldView(myWorldView)
				prevMyWorldView = copyMyWorldView(myWorldView)	
			}
		}
	}
}
