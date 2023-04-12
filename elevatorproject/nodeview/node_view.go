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
	ID                string
	IsAvailable       bool
	ElevState         singleelevator.ElevState
	HallRequests      [][2]RequestState // n number of floors
	MyCabRequests     []RequestState
	RemoteCabRequests map[string][]RequestState
}

type RemoteRequestView struct {
	RemoteHallRequestViews map[string][][2]RequestState
	RemoteCabRequests      map[string][]RequestState
	MyCabRequests          map[string][]RequestState
}

func CopyRemoveRequestView(remoteRequestView RemoteRequestView) RemoteRequestView {
	var copy RemoteRequestView
	copy.RemoteHallRequestViews = make(map[string][][2]RequestState, config.NumElevators)
	copy.RemoteCabRequests = make(map[string][]RequestState, config.NumElevators)
	copy.MyCabRequests = make(map[string][]RequestState, config.NumElevators)
	for id, hallRequestView := range remoteRequestView.RemoteHallRequestViews {
		copy.RemoteHallRequestViews[id] = hallRequestView
	}
	for id, cabRequestView := range remoteRequestView.RemoteCabRequests {
		copy.RemoteCabRequests[id] = cabRequestView
	}
	for id, cabRequestView := range remoteRequestView.MyCabRequests {
		copy.MyCabRequests[id] = cabRequestView
	}
	return copy
}

func (myNodeView *MyNodeView) InitMyNodeView() {
	myNodeView.ID = config.LocalID
	myNodeView.HallRequests = make([][2]RequestState, config.NumFloors)
	myNodeView.RemoteCabRequests = make(map[string][]RequestState)
	myNodeView.ElevState = singleelevator.ElevState{
		Behaviour:   "moving",
		Floor:       1,
		Direction:   "up",
		CabRequests: make([]bool, config.NumFloors),
		IsAvailable: true,
	}

}

// function that takes a [][2]RequestState as input and return [][2]bool

func convertRequestsToBool(hallrequests [][2]RequestState, cabrequests []RequestState, isSingleElevMode bool) [][3]bool {
	requests := make([][3]bool, len(hallrequests))
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
	for row := 0; row < len(cabrequests); row++ {
		if cabrequests[row] == RS_Confirmed {
			requests[row][2] = true
		} else if (cabrequests[row] == RS_Pending) && isSingleElevMode {
			requests[row][2] = true
		} else {
			requests[row][2] = false
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
	for i := 0; i < len(myNodeView.MyCabRequests); i++ {
		if myNodeView.MyCabRequests[i] == RS_NoOrder || myNodeView.MyCabRequests[i] == RS_Confirmed {
			myNodeView.MyCabRequests[i] = RS_Unknown
		}
	}
}

func (remoteRequestView *RemoteRequestView) InitRemoteRequestView() {
	remoteRequestView.RemoteHallRequestViews = make(map[string][][2]RequestState)
	remoteRequestView.RemoteCabRequests = make(map[string][]RequestState)
	remoteRequestView.MyCabRequests = make(map[string][]RequestState)
}

func printNodeAwareness(node MyNodeView) {
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

func NodeView(ch_sendMyNodeView chan<- MyNodeView,
	ch_newRequest <-chan elevio.ButtonEvent,
	ch_completedRequest <-chan elevio.ButtonEvent,
	ch_elevState <-chan singleelevator.ElevState,
	ch_lamps chan<- [][3]bool,
	ch_remoteRequestView <-chan RemoteRequestView) {

	var myNodeView MyNodeView
	var isSingleElevMode = true

	myNodeView.InitMyNodeView()

	for {
		select {
		case remoteRequestView := <-ch_remoteRequestView:
			//fmt.Println("nodeview: remoteRequestView")

			numRemoteNodes := len(remoteRequestView.RemoteHallRequestViews)
			fmt.Println("Is available", myNodeView.ElevState.IsAvailable)
			if numRemoteNodes > 0 {
				if isSingleElevMode {
					isSingleElevMode = false
					myNodeView.ChangeNoOrderAndConfirmedToUnknown()
				}
				myNodeView.HallRequests = updateMyHallRequestView(myNodeView.HallRequests, remoteRequestView.RemoteHallRequestViews)
				myNodeView.MyCabRequests = updateMyCabRequestView(myNodeView.MyCabRequests, remoteRequestView.MyCabRequests)
				myNodeView.RemoteCabRequests = remoteRequestView.RemoteCabRequests
			} else {
				isSingleElevMode = true
			}

			ch_lamps <- convertRequestsToBool(myNodeView.HallRequests, myNodeView.MyCabRequests, isSingleElevMode)

		case newRequest := <-ch_newRequest:
			//fmt.Println("nodeview: newHallRequest")
			// if newHallRequest is cabrequest, then set myNodeView.CabRequests
			if newRequest.Button == elevio.BT_Cab {
				myNodeView.MyCabRequests[newRequest.Floor] = RS_Pending
			} else {
				myNodeView.HallRequests[newRequest.Floor][int(newRequest.Button)] = RS_Pending
			}
			if isSingleElevMode {
				ch_lamps <- convertRequestsToBool(myNodeView.HallRequests, myNodeView.MyCabRequests, isSingleElevMode)
			}

		case completedHallRequest := <-ch_completedRequest:
			//fmt.Println("nodeview: completedHallRequest")
			nextRS := RS_Completed

			if isSingleElevMode {
				nextRS = RS_NoOrder
			}

			if completedHallRequest.Button == elevio.BT_Cab {
				myNodeView.MyCabRequests[completedHallRequest.Floor] = nextRS
			} else {
				myNodeView.HallRequests[completedHallRequest.Floor][int(completedHallRequest.Button)] = nextRS
			}

			if isSingleElevMode {
				ch_lamps <- convertRequestsToBool(myNodeView.HallRequests, myNodeView.MyCabRequests, isSingleElevMode)
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
		//time.Sleep(50*time.Millisecond)
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
