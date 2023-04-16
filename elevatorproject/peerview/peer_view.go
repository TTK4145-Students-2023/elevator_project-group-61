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
	HallRequests [config.NumFloors][2]RequestState
	CabRequests  map[string][config.NumFloors]RequestState
}

type RemoteRequestViews struct {
	RemoteHallRequestViews map[string][config.NumFloors][2]RequestState
	RemoteCabRequestViews  map[string]map[string][config.NumFloors]RequestState
}

func (myPeerView *MyPeerView) initMyPeerView(localID string) {
	myPeerView.ID = localID
	myPeerView.ElevState.InitElevState()
	myPeerView.HallRequests = [config.NumFloors][2]RequestState{}
	myPeerView.CabRequests = make(map[string][config.NumFloors]RequestState)
	myPeerView.CabRequests[localID] = [config.NumFloors]RequestState{}
}

func (remoteRequestViews *RemoteRequestViews) InitRemoteRequestViews() {
	remoteRequestViews.RemoteHallRequestViews = make(map[string][config.NumFloors][2]RequestState)
	remoteRequestViews.RemoteCabRequestViews = make(map[string]map[string][config.NumFloors]RequestState)
}

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
		// Go to maximum state of all other peers except RS_Completed
		max_count := int(myRequest)
		for _, request := range remoteRequest {
			if (int(request) > max_count) && request != RS_Completed {
				max_count = int(request)
			}
		}
		updatedRequest = RequestState(max_count)
	case RS_NoOrder:
		// Go to RS_Pending if any other peer has RS_Pending
		for _, request := range remoteRequest {
			if request == RS_Pending {
				updatedRequest = RS_Pending
				break
			}
		}
	case RS_Pending:
		// Go to RS_Confirmed if all peers alive have RS_Pending, or any other peer has RS_Confirmed
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
		// Go to RS_NoOrder if any other peer has RS_Completed
		for _, request := range remoteRequest {
			if request == RS_Completed {
				updatedRequest = RS_NoOrder
				break
			}
		}
	case RS_Completed:
		// Go to RS_NoOrder if all other peers have anything else than RS_Confirmed
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

func updateMyCabRequestView(myCabRequestView [config.NumFloors]RequestState, remoteCabRequestViews map[string][config.NumFloors]RequestState) [config.NumFloors]RequestState {
	for i := 0; i < len(myCabRequestView); i++ {
		cabOrder := myCabRequestView[i]
		remoteCabRequest := make(map[string]RequestState)
		for id, cabRequestView := range remoteCabRequestViews {
			remoteCabRequest[id] = cabRequestView[i]
		}
		myCabRequestView[i] = updateSingleRequest(cabOrder, remoteCabRequest)
	}
	return myCabRequestView
}

func getAllRemoteCabRequestViewsForSpecificPeer(remoteCabRequestViews map[string]map[string][config.NumFloors]RequestState, id string) map[string][config.NumFloors]RequestState {
	remoteCabRequestViewsSpecificPeer := make(map[string][config.NumFloors]RequestState)
	for remoteID, remoteCabRequestViews := range remoteCabRequestViews {
		if remoteCabRequestView, ok := remoteCabRequestViews[id]; ok {
			remoteCabRequestViewsSpecificPeer[remoteID] = remoteCabRequestView
		}
	}
	return remoteCabRequestViewsSpecificPeer
}


func changeNoOrderAndConfirmedToUnknown(myPeerView MyPeerView) MyPeerView {
	changedMyPeerView := myPeerView
	for floor := 0; floor < len(changedMyPeerView.HallRequests); floor++ {
		for btn := 0; btn < len(changedMyPeerView.HallRequests[floor]); btn++ {
			if changedMyPeerView.HallRequests[floor][btn] == RS_NoOrder || changedMyPeerView.HallRequests[floor][btn] == RS_Confirmed {
				changedMyPeerView.HallRequests[floor][btn] = RS_Unknown
			}
		}
	}
	for id, cabRequests := range myPeerView.CabRequests {
		newCabRequests := cabRequests
		for floor := 0; floor < config.NumFloors; floor++ {
			if cabRequests[floor] == RS_NoOrder || cabRequests[floor] == RS_Confirmed {
				newCabRequests[floor] = RS_Unknown
			}
			changedMyPeerView.CabRequests[id] = newCabRequests
		}
	}
	return changedMyPeerView
}

func HasRemoteRequestViewsChanged(remoteRequestViews RemoteRequestViews, prevRemoteRequestViews RemoteRequestViews) bool {
	for id, requestStates := range remoteRequestViews.RemoteHallRequestViews {
		for floor := 0; floor < config.NumFloors; floor++ {
			for button := 0; button < 2; button++ {
				if requestStates[floor][button] != prevRemoteRequestViews.RemoteHallRequestViews[id][floor][button] {
					return true
				}
			}
		}
	}
	
	for id1, value1 := range remoteRequestViews.RemoteCabRequestViews {
		for id2, value2 := range value1 {
			for floor := 0; floor < config.NumFloors; floor++ {
				if value2[floor] != prevRemoteRequestViews.RemoteCabRequestViews[id1][id2][floor] {
					return true
				}
			}
		}
	}
	return false
}

func copyMyPeerView(myPeerView MyPeerView) MyPeerView {
	copyPeerView := MyPeerView{}
	copyPeerView.ID = myPeerView.ID
	copyPeerView.ElevState = myPeerView.ElevState
	copyPeerView.HallRequests = myPeerView.HallRequests
	copyPeerView.CabRequests = make(map[string][config.NumFloors]RequestState)
	for key, value := range myPeerView.CabRequests {
		copyPeerView.CabRequests[key] = value
	}
	return copyPeerView
}

func CopyRemoteRequestViews(remoteRequestViews RemoteRequestViews) RemoteRequestViews {
	copyRemoteRequestViews := RemoteRequestViews{}
	copyRemoteRequestViews.RemoteHallRequestViews = make(map[string][config.NumFloors][2]RequestState)
	for key, value := range remoteRequestViews.RemoteHallRequestViews {
		copyRemoteRequestViews.RemoteHallRequestViews[key] = value
	}
	copyRemoteRequestViews.RemoteCabRequestViews = make(map[string]map[string][config.NumFloors]RequestState)

	for key, value := range remoteRequestViews.RemoteCabRequestViews {
		copyRemoteRequestViews.RemoteCabRequestViews[key] = make(map[string][config.NumFloors]RequestState)
		for key2, value2 := range value {
			copyRemoteRequestViews.RemoteCabRequestViews[key][key2] = value2
		}
	}
	return copyRemoteRequestViews
}


func PeerView(ch_transmit chan<- MyPeerView,
	ch_newRequest <-chan elevio.ButtonEvent,
	ch_completedRequest <-chan elevio.ButtonEvent,
	ch_elevState <-chan singleelevator.ElevState,
	ch_hallLamps chan<- [config.NumFloors][2]bool,
	ch_cabLamps chan<- [config.NumFloors]bool,
	ch_remoteRequestViews <-chan RemoteRequestViews,
	localID string) {

	var myPeerView MyPeerView
	var isSingleElevMode = true

	myPeerView.initMyPeerView(localID)

	for {
		select {
		case remoteRequestViews := <-ch_remoteRequestViews:
			numRemotePeers := len(remoteRequestViews.RemoteHallRequestViews)

			// Initialize my cab request view of a newly connected peer
			for remoteID := range remoteRequestViews.RemoteCabRequestViews {
				if _, ok := myPeerView.CabRequests[remoteID]; !ok {
					myPeerView.CabRequests[remoteID] = [config.NumFloors]RequestState{}
				}
			}

			// Updating my request views when there are remote peers on the network
			if numRemotePeers > 0 {
				if isSingleElevMode {
					isSingleElevMode = false
					myPeerView = changeNoOrderAndConfirmedToUnknown(myPeerView)
				}
				// Update cab requests for all alive peers
				for id, myCabRequestViewSpecificPeer := range myPeerView.CabRequests {
					remoteCabRequestViewsSpecificPeer := getAllRemoteCabRequestViewsForSpecificPeer(remoteRequestViews.RemoteCabRequestViews, id)
					myPeerView.CabRequests[id] = updateMyCabRequestView(myCabRequestViewSpecificPeer, remoteCabRequestViewsSpecificPeer)
				}
				myPeerView.HallRequests = updateMyHallRequestView(myPeerView.HallRequests, remoteRequestViews.RemoteHallRequestViews)

			} else {
				isSingleElevMode = true
			}

			ch_hallLamps <- convertHallRequests(myPeerView.HallRequests, isSingleElevMode)
			ch_cabLamps <- convertCabRequests(myPeerView.CabRequests[localID], isSingleElevMode)

		case newRequest := <-ch_newRequest:
			if newRequest.Button == elevio.BT_Cab {
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

		case <-time.After(50 * time.Millisecond):
			ch_transmit <- copyMyPeerView(myPeerView)
		}
	}

}
