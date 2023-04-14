package peerview

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

type MyPeerView struct {
	ID           string
	ElevState    singleelevator.ElevState
	HallRequests [config.NumFloors][2]RequestState // n number of floors
	CabRequests  map[string][config.NumFloors]RequestState
}

type RemoteRequestView struct {
	RemoteHallRequestViews map[string][config.NumFloors][2]RequestState
	RemoteCabRequestViews  map[string]map[string][config.NumFloors]RequestState
}

// make function that returns a deep copy of myNodeView

func copyMyPeerView(myPeerView MyPeerView) MyPeerView {
	copyPeerView := MyPeerView{}
	copyPeerView.ID = myPeerView.ID
	copyPeerView.ElevState = myPeerView.ElevState
	//copyNodeView.ElevState = singleelevator.CopyElevState(myNodeView.ElevState)
	copyPeerView.HallRequests = myPeerView.HallRequests
	copyPeerView.CabRequests = make(map[string][config.NumFloors]RequestState)
	for key, value := range myPeerView.CabRequests {
		copyPeerView.CabRequests[key] = value
	}
	return copyPeerView
}

func CopyRemoteRequestView(remoteRequestView RemoteRequestView) RemoteRequestView {
	copyRemoteRequestView := RemoteRequestView{}
	copyRemoteRequestView.RemoteHallRequestViews = make(map[string][config.NumFloors][2]RequestState)
	for key, value := range remoteRequestView.RemoteHallRequestViews {
		copyRemoteRequestView.RemoteHallRequestViews[key] = value
	}
	copyRemoteRequestView.RemoteCabRequestViews = make(map[string]map[string][config.NumFloors]RequestState)
	for key, value := range remoteRequestView.RemoteCabRequestViews {
		copyRemoteRequestView.RemoteCabRequestViews[key] = make(map[string][config.NumFloors]RequestState)
		for key2, value2 := range value {
			copyRemoteRequestView.RemoteCabRequestViews[key][key2] = value2
		}
	}
	return copyRemoteRequestView
}

// New initMyNodeView function that initializes the MyNodeView struct and all requests to RS_Unknown
func (myPeerView *MyPeerView) InitMyPeerView(localID string) {
	myPeerView.ID = localID
	myPeerView.ElevState.InitElevState()
	myPeerView.HallRequests = [config.NumFloors][2]RequestState{}
	myPeerView.CabRequests = make(map[string][config.NumFloors]RequestState)
	myPeerView.CabRequests[localID] = [config.NumFloors]RequestState{}
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
	for floor := 0; floor < len(myHallRequestView); floor++ {
		for btn := 0; btn < len(myHallRequestView[floor]); btn++ {
			hallRequest := myHallRequestView[floor][btn]
			remoteHallRequest := make(map[string]RequestState)
			for id, hallRequestView := range remoteHallRequestView {
				remoteHallRequest[id] = hallRequestView[floor][btn]
			}
			myHallRequestView[floor][btn] = updateSingleRequest(hallRequest, remoteHallRequest)
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

func (myNodeView *MyPeerView) ChangeNoOrderAndConfirmedToUnknown() {
	for floor := 0; floor < len(myNodeView.HallRequests); floor++ {
		for btn := 0; btn < len(myNodeView.HallRequests[floor]); btn++ {
			if myNodeView.HallRequests[floor][btn] == RS_NoOrder || myNodeView.HallRequests[floor][btn] == RS_Confirmed {
				myNodeView.HallRequests[floor][btn] = RS_Unknown
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

func PeerView(ch_sendMyPeerView chan<- MyPeerView,
	ch_newRequest <-chan elevio.ButtonEvent,
	ch_completedRequest <-chan elevio.ButtonEvent,
	ch_elevState <-chan singleelevator.ElevState,
	ch_hallLamps chan<- [config.NumFloors][2]bool,
	ch_cabLamps chan<- [config.NumFloors]bool,
	ch_remoteRequestView <-chan RemoteRequestView,
	localID string) {

	var myPeerView MyPeerView
	var isSingleElevMode = true

	myPeerView.InitMyPeerView(localID)

	for {
		select {
		case remoteRequestView := <-ch_remoteRequestView:
			numRemoteNodes := len(remoteRequestView.RemoteHallRequestViews)
			fmt.Println("Is available", myPeerView.ElevState.IsAvailable)

			
			for remoteID := range remoteRequestView.RemoteCabRequestViews {
				if _, ok := myPeerView.CabRequests[remoteID]; !ok {
					myPeerView.CabRequests[remoteID] = [config.NumFloors]RequestState{}
				}
			}
			
			//fmt.Println("hi")
			if numRemoteNodes > 0 {
				if isSingleElevMode {
					isSingleElevMode = false
					myPeerView.ChangeNoOrderAndConfirmedToUnknown()
				}
				// Run update my cab request view on every node in myNodeView.CabRequests
				for id, myCabRequestView := range myPeerView.CabRequests {
					specificPeerRemoteCabRequestViews := make(map[string][config.NumFloors]RequestState)
					for remoteID, remoteCabRequestViews := range remoteRequestView.RemoteCabRequestViews {
						if remoteCabRequestView, ok := remoteCabRequestViews[id]; ok {
							specificPeerRemoteCabRequestViews[remoteID] = remoteCabRequestView
						}
					}
					myPeerView.CabRequests[id] = updateMyCabRequestView(myCabRequestView, specificPeerRemoteCabRequestViews)
				}
				myPeerView.HallRequests = updateMyHallRequestView(myPeerView.HallRequests, remoteRequestView.RemoteHallRequestViews)

			} else {
				isSingleElevMode = true
			}

			ch_hallLamps <- convertHallRequests(myPeerView.HallRequests, isSingleElevMode)
			ch_cabLamps <- convertCabRequests(myPeerView.CabRequests[localID], isSingleElevMode)


		case newRequest := <-ch_newRequest:
			//fmt.Println("nodeview: newHallRequest")
			// if newHallRequest is cabrequest, then set myNodeView.CabRequests
			if newRequest.Button == elevio.BT_Cab {
				// set all cab calls for localID to pending
				cabs := myPeerView.CabRequests[localID]
				cabs[newRequest.Floor] = RS_Pending
				myPeerView.CabRequests[localID] = cabs
			} else {
				myPeerView.HallRequests[newRequest.Floor][int(newRequest.Button)] = RS_Pending
			}
			if isSingleElevMode {
				ch_hallLamps <- convertHallRequests(myPeerView.HallRequests, isSingleElevMode)
				ch_cabLamps <- convertCabRequests(myPeerView.CabRequests[localID], isSingleElevMode)
			}

		case completedRequest := <-ch_completedRequest:
			//fmt.Println("nodeview: completedHallRequest")
			nextRS := RS_Completed

			if isSingleElevMode {
				nextRS = RS_NoOrder
			}

			if completedRequest.Button == elevio.BT_Cab {
				cabs := myPeerView.CabRequests[localID]
				cabs[completedRequest.Floor] = nextRS
				myPeerView.CabRequests[localID] = cabs
			} else {
				myPeerView.HallRequests[completedRequest.Floor][int(completedRequest.Button)] = nextRS
			}

			if isSingleElevMode {
				ch_hallLamps <- convertHallRequests(myPeerView.HallRequests, isSingleElevMode)
				ch_cabLamps <- convertCabRequests(myPeerView.CabRequests[localID], isSingleElevMode)
			}

		case elevState := <-ch_elevState:
			myPeerView.ElevState = elevState

		case <-time.After(25 * time.Millisecond):
			fmt.Println("nodeview: broadcaster myNodeView")
			ch_sendMyPeerView <- copyMyPeerView(myPeerView)
		}
	}

}
