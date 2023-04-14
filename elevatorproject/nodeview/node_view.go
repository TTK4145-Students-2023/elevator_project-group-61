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
	RS_Unknown   RequestState = 0
	RS_NoOrder   RequestState = 1
	RS_Pending   RequestState = 2
	RS_Confirmed RequestState = 3
	RS_Completed RequestState = 4
)

type MyNodeView struct {
	ID           string
	IsAvailable  bool
	ElevState    singleelevator.ElevState
	HallRequests [config.NumFloors][2]RequestState // n number of floors
	CabRequests  map[string][config.NumFloors]RequestState
}

type RemoteRequestView struct {
	RemoteHallRequestViews map[string][config.NumFloors][2]RequestState
	RemoteCabRequestViews  map[string]map[string][config.NumFloors]RequestState
}


// New initMyNodeView function that initializes the MyNodeView struct and all requests to RS_Unknown
func (myNodeView *MyNodeView) InitMyNodeView(localID string) {
	myNodeView.ID = localID
	myNodeView.IsAvailable = true
	myNodeView.ElevState.InitElevState()
	myNodeView.HallRequests = [config.NumFloors][2]RequestState{}
	myNodeView.CabRequests = make(map[string][config.NumFloors]RequestState)
	myNodeView.CabRequests[localID] = [config.NumFloors]RequestState{}
}
// function that takes a [][2]RequestState as input and return [][2]bool

func convertHallRequests(hallrequests [config.NumFloors][2]RequestState, isSingleElevMode bool) [config.NumFloors][2]bool {
	requests := [config.NumFloors][2]bool{}
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

func convertCabRequests(cabrequests [config.NumFloors]RequestState, isSingleElevMode bool) [config.NumFloors]bool {
	requests := [config.NumFloors]bool{}
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

func updateMyHallRequestView(myHallRequestView [config.NumFloors][2]RequestState, remoteHallRequestView map[string][config.NumFloors][2]RequestState) [config.NumFloors][2]RequestState {
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
func updateMyCabRequestView(myCabRequestView [config.NumFloors]RequestState, remoteCabRequestViews map[string][config.NumFloors]RequestState) [config.NumFloors]RequestState {
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
	for _, cabRequests := range myNodeView.CabRequests {
		for floor := 0; floor < config.NumFloors; floor++ {
			if cabRequests[floor] == RS_NoOrder || cabRequests[floor] == RS_Confirmed {
				cabRequests[floor] = RS_Unknown
				// TODO: Må sjekke om det blir oppdater her
			}
		}
	}

}

func (remoteRequestView *RemoteRequestView) InitRemoteRequestView() {
	remoteRequestView.RemoteHallRequestViews = make(map[string][config.NumFloors][2]RequestState)
	remoteRequestView.RemoteCabRequestViews = make(map[string]map[string][config.NumFloors]RequestState)
}

func NodeView(ch_sendMyNodeView chan<- MyNodeView,
	ch_newRequest <-chan elevio.ButtonEvent,
	ch_completedRequest <-chan elevio.ButtonEvent,
	ch_elevState <-chan singleelevator.ElevState,
	ch_hallLamps chan<- [config.NumFloors][2]bool,
	ch_cabLamps chan<- [config.NumFloors]bool,
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

			for remoteID, _ := range remoteRequestView.RemoteCabRequestViews {
				if _, ok := myNodeView.CabRequests[remoteID]; !ok {
					myNodeView.CabRequests[remoteID] = [config.NumFloors]RequestState{}
				}
			}
			fmt.Println("hi")
			if numRemoteNodes > 0 {
				if isSingleElevMode {
					isSingleElevMode = false
					myNodeView.ChangeNoOrderAndConfirmedToUnknown()
				}
				// Run update my cab request view on every node in myNodeView.CabRequests


				for id, myCabRequestView := range myNodeView.CabRequests {
					specificPeerRemoteCabRequestViews := make(map[string][config.NumFloors]RequestState)
					for remoteID, remoteCabRequestViews := range remoteRequestView.RemoteCabRequestViews {
						if remoteCabRequestView, ok := remoteCabRequestViews[id]; ok {
							specificPeerRemoteCabRequestViews[remoteID] = remoteCabRequestView
						}
					}
					myNodeView.CabRequests[id] = updateMyCabRequestView(myCabRequestView, specificPeerRemoteCabRequestViews)
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
