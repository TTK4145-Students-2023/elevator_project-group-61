package worldview

import (
	"elevatorproject/config"
	"elevatorproject/network/peers"
	"elevatorproject/nodeview"
	"elevatorproject/singleelevator"
	"fmt"
	"strings"
)

type PeersAlive []string

type MyWorldView struct {
	ElevStates     map[string]singleelevator.ElevState
	HallRequestView   [][2]nodeview.RequestState
	CabRequests  map[string][]nodeview.RequestState
	NodesAvailable map[string]bool
}


// Make deap copy of MyWorldView

func copyMyWorldView(worldView MyWorldView) MyWorldView {
	var copy MyWorldView
	copy.ElevStates = make(map[string]singleelevator.ElevState, config.NumElevators)
	copy.HallRequestView = make([][2]nodeview.RequestState, config.NumFloors)
	copy.CabRequests = make(map[string][]nodeview.RequestState, config.NumElevators)
	copy.NodesAvailable = make(map[string]bool, config.NumElevators)
	for id, elevState := range worldView.ElevStates {
		copy.ElevStates[id] = elevState
	}
	for id, cabRequests := range worldView.CabRequests {
		copy.CabRequests[id] = cabRequests
	}
	for id, isAvailable := range worldView.NodesAvailable {
		copy.NodesAvailable[id] = isAvailable
	}
	for floor := 0; floor < config.NumElevators; floor++ {
		copy.HallRequestView[floor] = worldView.HallRequestView[floor]
	}
	return copy
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
	myWorldView.HallRequestView = make([][2]nodeview.RequestState, config.NumFloors)
	myWorldView.NodesAvailable = make(map[string]bool, config.NumElevators)
	myWorldView.NodesAvailable[config.LocalID] = true
	//myWorldView.NodesAvailable[config.SecondElev] = true
	myWorldView.ElevStates[config.LocalID] = singleelevator.ElevState{
		Behaviour: "moving",
		Floor : 1,
		Direction: "up",
		CabRequests: make([]bool, config.NumFloors),
		IsAvailable: true,
	}
	// init cab requests
	myWorldView.CabRequests = make(map[string][]nodeview.RequestState, config.NumElevators)
	myWorldView.CabRequests[config.LocalID] = make([]nodeview.RequestState, config.NumFloors)
	
}

func WorldView(ch_receiveNodeView <-chan nodeview.MyNodeView,
	ch_receivePeerUpdate <-chan peers.PeerUpdate,
	ch_setTransmitEnable chan <- bool,
	ch_CabRequests chan <- []bool,
	ch_remoteRequestView chan <- nodeview.RemoteRequestView,
	ch_hraInput chan<- MyWorldView,
	ch_singleElevMode chan <- bool) {


	var myWorldView MyWorldView
	var peersAlive PeersAlive
	var remoteRequestView nodeview.RemoteRequestView
	var isSingleElevMode bool

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
	
					delete(remoteRequestView.RemoteHallRequestViews, lostPeer)
					delete(remoteRequestView.RemoteCabRequestViews, lostPeer)
				}
			}
			//TODO: Må undersøke om denne checken er nok
			if len(peersAlive) <= 1 {
				ch_singleElevMode <- true
				isSingleElevMode = true
				ch_remoteRequestView <- remoteRequestView
			} else {
				ch_singleElevMode <- false
				isSingleElevMode = false
			}
		case nodeView := <-ch_receiveNodeView:
			//fmt.Println("worldview: nodeView")
			fmt.Println("Received from ", nodeView.ID)

			nodeID := nodeView.ID
			// Break out of case if IsPeerAlive returns false
			if !peersAlive.IsPeerAlive(nodeID) && config.LocalID != nodeID {
				break
			}
			myWorldView.NodesAvailable[nodeID] = nodeView.ElevState.IsAvailable
			
			myWorldView.ElevStates[nodeID] = nodeView.ElevState
			myWorldView.CabRequests[nodeID] = nodeView.CabRequests[nodeID]

			if nodeID != config.LocalID {
				remoteRequestView.RemoteHallRequestViews[nodeID] = nodeView.HallRequests
				remoteRequestView.RemoteCabRequestViews[nodeID] = nodeView.CabRequests
			} else {
				myWorldView.HallRequestView = nodeView.HallRequests
			}

			ch_remoteRequestView <- nodeview.CopyRemoveRequestView(remoteRequestView)
			if !isSingleElevMode {
				ch_hraInput <- copyMyWorldView(myWorldView)
			}
		}
	}
}

func PrintMyWorldView(view MyWorldView) {
	fmt.Printf("Elevators:\n")
	for id, state := range view.ElevStates {
		fmt.Printf("  %s\n", id)
		fmt.Printf("    - Behaviour: %s\n", state.Behaviour)
		fmt.Printf("    - Floor: %d\n", state.Floor)
		fmt.Printf("    - Direction: %s\n", state.Direction)
		fmt.Printf("    - CabRequests: %s\n", boolSliceToString(state.CabRequests))
		fmt.Printf("    - IsAvailable: %t\n", state.IsAvailable)
	}
	fmt.Printf("\n")

	fmt.Printf("Hall Requests:\n")
	for i, req := range view.HallRequestView {
		fmt.Printf("  Floor %d:\n", i+1)
		for j, state := range req {
			fmt.Printf("    %s Hall: %s\n", hallDirectionToString(j), nodeview.RequestStateToString(state))
		}
	}
	fmt.Printf("\n")

	fmt.Printf("Node Availability:\n")
	for id, available := range view.NodesAvailable {
		fmt.Printf("  %s: %t\n", id, available)
	}
}

func boolSliceToString(arr []bool) string {
	return strings.Join(strings.Split(fmt.Sprintf("%v", arr), " "), ", ")
}

func hallDirectionToString(dir int) string {
	if dir == 0 {
		return "Up"
	} else {
		return "Down"
	}
}

