package nodeview

import (
	"elevatorproject/config"
	"elevatorproject/singleelevator"
	"elevatorproject/singleelevator/elevio"
	"fmt"
	"time"
)

type RequestState int

const (
	RS_Unknown   RequestState = -1
	RS_NoOrder   RequestState = 0
	RS_Pending   RequestState = 1
	RS_Confirmed RequestState = 2
	RS_Completed RequestState = 3
)

type MyNodeView struct {
	ID           string
	IsAvailable  bool
	ElevState    singleelevator.ElevState
	HallRequests [][2]RequestState // n number of floors
	CabRequests  map[string][]RequestState
}

type RemoteRequestView struct {
	RemoteHallRequestViews map[string][][2]RequestState
	RemoteCabRequestViews  map[string]map[string][]RequestState
}

func CopyRemoteRequestView(remoteRequestView RemoteRequestView) RemoteRequestView {
	var copy RemoteRequestView
	copy.RemoteHallRequestViews = make(map[string][][2]RequestState, config.NumElevators)
	copy.RemoteCabRequestViews = make(map[string]map[string][]RequestState, config.NumElevators)
	for id, hallRequests := range remoteRequestView.RemoteHallRequestViews {
		copy.RemoteHallRequestViews[id] = make([][2]RequestState, config.NumFloors)
		for floor := 0; floor < config.NumFloors; floor++ {
			copy.RemoteHallRequestViews[id][floor] = hallRequests[floor]
		}
	}
	for id, cabRequests := range remoteRequestView.RemoteCabRequestViews {
		copy.RemoteCabRequestViews[id] = make(map[string][]RequestState, config.NumElevators)
		for id2, cabRequests2 := range cabRequests {
			copy.RemoteCabRequestViews[id][id2] = make([]RequestState, config.NumFloors)
			for floor := 0; floor < config.NumFloors; floor++ {
				copy.RemoteCabRequestViews[id][id2][floor] = cabRequests2[floor]
			}
		}
	}
	return copy
}

func (myNodeView *MyNodeView) InitMyNodeView(localID string) {
	myNodeView.ID = localID
	myNodeView.HallRequests = make([][2]RequestState, config.NumFloors)
	myNodeView.CabRequests = make(map[string][]RequestState)
	// Run InitElevState
	myNodeView.ElevState.InitElevState();
}

// function that takes a [][2]RequestState as input and return [][2]bool

func convertHallRequests(hallrequests [][2]RequestState, isSingleElevMode bool) [][2]bool {
	requests := make([][2]bool, len(hallrequests))
	for row := 0; row < len(hallrequests); row++ {
		for col := 0; col < len(hallrequests[row]); col++ {
			if hallrequests[row][col] == RS_Confirmed {
				requests[row][col] = true
			} else if (hallrequests[row][col] == RS_Pending) && isSingleElevMode {
				requests[row][col] = true
			} else {
				requests[row][col] = false
			}
		}
	}
	return requests
} 

func convertCabRequests(cabrequests []RequestState, isSingleElevMode bool) []bool {
	requests := make([]bool, len(cabrequests))
	for floor := 0; floor < len(cabrequests); floor++ {
		if cabrequests[floor] == RS_Confirmed {
			requests[floor] = true
		} else if (cabrequests[floor] == RS_Pending) && isSingleElevMode {
			requests[floor] = true
		} else {
			requests[floor] = false
		}
	}
	return requests
}

func updateSingleRequest(myRequest RequestState, remoteRequest map[string]RequestState) RequestState {
	updatedRequest := myRequest
	switch myRequest {
	case RS_Unknown:
		max_count := int(myRequest)
		for _, request := range remoteRequest {
			if (int(request) > max_count) && request != RS_Completed {
				max_count = int(request)
			}
		}
		updatedRequest = RequestState(max_count)
	case RS_NoOrder:
		// Go to RS_Pending if any other node has RS_Pending
		for _, request := range remoteRequest {
			if request == RS_Pending {
				updatedRequest = RS_Pending
				break
			}
		}
	case RS_Pending:
		pendingCount := 0
		for _, request := range remoteRequest {
			if request == RS_Confirmed {
				updatedRequest = RS_Confirmed
				break
			} else if request == RS_Pending {
				pendingCount++
			}
		}
		if pendingCount == len(remoteRequest) {
			updatedRequest = RS_Confirmed
		}
	case RS_Confirmed:
		for _, request := range remoteRequest {
			if request == RS_Completed {
				updatedRequest = RS_NoOrder
				break
			}
		}
	case RS_Completed:
		// Go to RS_NoOrder if all other nodes have anything else than RS_Confirmed
		noOrderCount := 0
		for _, request := range remoteRequest {
			if request != RS_Confirmed {
				noOrderCount++
			}
		}
		if noOrderCount == len(remoteRequest) {
			updatedRequest = RS_NoOrder
		}
	}
	return updatedRequest
}

func updateMyHallRequestView(myHallRequestView [][2]RequestState, remoteHallRequestView map[string][][2]RequestState) [][2]RequestState {
	for row := 0; row < len(myHallRequestView); row++ {
		for col := 0; col < len(myHallRequestView[row]); col++ {
			hallRequest := myHallRequestView[row][col]
			remoteHallRequest := make(map[string]RequestState)
			for id, hallRequestView := range remoteHallRequestView {
				remoteHallRequest[id] = hallRequestView[row][col]
			}
			myHallRequestView[row][col] = updateSingleRequest(hallRequest, remoteHallRequest)
		}
	}
	return myHallRequestView
}

// Make the same function as updateMyHallRequestView but for cab requests
func updateMyCabRequestView(myCabRequestView []RequestState, remoteCabRequestViews map[string][]RequestState) []RequestState {
	for i := 0; i < len(myCabRequestView); i++ {
		cab_order := myCabRequestView[i]
		remoteCabRequest := make(map[string]RequestState)
		for id, cabRequestView := range remoteCabRequestViews {
			remoteCabRequest[id] = cabRequestView[i]
		}
		myCabRequestView[i] = updateSingleRequest(cab_order, remoteCabRequest)
	}
	return myCabRequestView
}

func (myNodeView *MyNodeView) ChangeNoOrderAndConfirmedToUnknown() {
	for row := 0; row < len(myNodeView.HallRequests); row++ {
		for col := 0; col < len(myNodeView.HallRequests[row]); col++ {
			if myNodeView.HallRequests[row][col] == RS_NoOrder || myNodeView.HallRequests[row][col] == RS_Confirmed {
				myNodeView.HallRequests[row][col] = RS_Unknown
			}
		}
	}
	// Do the same for cab requests of all elevators
	for peerID, cabRequests := range myNodeView.CabRequests {
		for request := 0; request < len(cabRequests); request++ {
			if cabRequests[request] == RS_NoOrder || cabRequests[request] == RS_Confirmed {
				myNodeView.CabRequests[peerID][request] = RS_Unknown
			}
		}
	}

}

func (remoteRequestView *RemoteRequestView) InitRemoteRequestView() {
	remoteRequestView.RemoteHallRequestViews = make(map[string][][2]RequestState)
	remoteRequestView.RemoteCabRequestViews = make(map[string]map[string][]RequestState)
}

func NodeView(ch_sendMyNodeView chan<- MyNodeView,
	ch_newRequest <-chan elevio.ButtonEvent,
	ch_completedRequest <-chan elevio.ButtonEvent,
	ch_elevState <-chan singleelevator.ElevState,
	ch_hallLamps chan<- [][2]bool,
	ch_cabLamps chan<- []bool,
	ch_remoteRequestView <-chan RemoteRequestView,
	localID string) {

	var myNodeView MyNodeView
	var isSingleElevMode = true

	myNodeView.InitMyNodeView(localID)

	for {
		select {
		case remoteRequestView := <-ch_remoteRequestView:
			numRemoteNodes := len(remoteRequestView.RemoteHallRequestViews)
			fmt.Println("Is available", myNodeView.ElevState.IsAvailable)
			if numRemoteNodes > 0 {
				if isSingleElevMode {
					isSingleElevMode = false
					myNodeView.ChangeNoOrderAndConfirmedToUnknown()
				}
				// Run update my cab request view on every node in myNodeView.CabRequests
				for id, myCabRequestView := range myNodeView.CabRequests {
					remoteCabRequestView := make(map[string][]RequestState)
					for remoteID, remoteCabRequestView := range remoteRequestView.RemoteCabRequestViews {
						remoteCabRequestView[remoteID] = remoteCabRequestView[id]
					}
					myNodeView.CabRequests[id] = updateMyCabRequestView(myCabRequestView, remoteCabRequestView)
				}
				myNodeView.HallRequests = updateMyHallRequestView(myNodeView.HallRequests, remoteRequestView.RemoteHallRequestViews)

			} else {
				isSingleElevMode = true
			}

			ch_hallLamps <- convertHallRequests(myNodeView.HallRequests, isSingleElevMode)
			ch_cabLamps <- convertCabRequests(myNodeView.CabRequests[localID], isSingleElevMode)


		case newRequest := <-ch_newRequest:
			//fmt.Println("nodeview: newHallRequest")
			// if newHallRequest is cabrequest, then set myNodeView.CabRequests
			if newRequest.Button == elevio.BT_Cab {
				myNodeView.CabRequests[localID][newRequest.Floor] = RS_Pending
			} else {
				myNodeView.HallRequests[newRequest.Floor][int(newRequest.Button)] = RS_Pending
			}
			if isSingleElevMode {
				ch_hallLamps <- convertHallRequests(myNodeView.HallRequests, isSingleElevMode)
				ch_cabLamps <- convertCabRequests(myNodeView.CabRequests[localID], isSingleElevMode)
			}

		case completedHallRequest := <-ch_completedRequest:
			//fmt.Println("nodeview: completedHallRequest")
			nextRS := RS_Completed

			if isSingleElevMode {
				nextRS = RS_NoOrder
			}

			if completedHallRequest.Button == elevio.BT_Cab {
				myNodeView.CabRequests[localID][completedHallRequest.Floor] = nextRS
			} else {
				myNodeView.HallRequests[completedHallRequest.Floor][int(completedHallRequest.Button)] = nextRS
			}

			if isSingleElevMode {
				ch_hallLamps <- convertHallRequests(myNodeView.HallRequests, isSingleElevMode)
				ch_cabLamps <- convertCabRequests(myNodeView.CabRequests[localID], isSingleElevMode)
			}

		case elevState := <-ch_elevState:
			fmt.Println("nodeview: elevState")
			myNodeView.ElevState = elevState

		case <-time.After(100 * time.Millisecond):
			fmt.Println("nodeview: broadcaster myNodeView")
			ch_sendMyNodeView <- myNodeView

			//default:
			//time.Sleep(100*time.Millisecond)
		}
	}

}

func RequestStateToString(state RequestState) string {
	switch state {
	case RS_Unknown:
		return "Unknown"
	case RS_NoOrder:
		return "No Order"
	case RS_Pending:
		return "Pending"
	case RS_Confirmed:
		return "Confirmed"
	case RS_Completed:
		return "Completed"
	default:
		return fmt.Sprintf("%d", state)
	}
}
