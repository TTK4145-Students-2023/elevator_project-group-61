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
	RemoteCabRequests map[string][]bool
}

type RemoteRequestView struct {
	RemoteHallRequestViews map[string][][2]RequestState
	RemoteCabRequests      map[string][]bool
}

func (myNodeView *MyNodeView) InitMyNodeView() {
	myNodeView.ID = config.LocalID
	myNodeView.HallRequests = make([][2]RequestState, config.NumFloors)
	myNodeView.RemoteCabRequests = make(map[string][]bool)
	myNodeView.ElevState = singleelevator.ElevState{
		Behaviour: "moving",
		Floor : 1,
		Direction: "up",
		CabRequests: make([]bool, config.NumFloors),
		IsAvailable: true,
	}
}

// function that takes a [][2]RequestState as input and return [][2]bool
func convertHallRequestStateToBool(hallRequests [][2]RequestState, singleElevatorMode bool) [][2]bool {
	hallRequestsBool := make([][2]bool, len(hallRequests))
	for row := 0; row < len(hallRequests); row++ {
		for col := 0; col < len(hallRequests[row]); col++ {
			if hallRequests[row][col] == RS_Confirmed {
				hallRequestsBool[row][col] = true
			} else if (hallRequests[row][col] == RS_Pending) && singleElevatorMode {
				hallRequestsBool[row][col] = true
			} else {
				hallRequestsBool[row][col] = false
			}
		}
	}
	return hallRequestsBool
}


func updateMyHallRequestView(myHallRequestView [][2]RequestState, remoteHallRequestView map[string][][2]RequestState) [][2]RequestState {
	for row := 0; row < len(myHallRequestView); row++ {
		for col := 0; col < len(myHallRequestView[row]); col++ {
			hall_order := myHallRequestView[row][col]

			switch hall_order {
			case RS_Unknown:
				max_count := int(hall_order)
				for _, nodeView := range remoteHallRequestView {
					if (int(nodeView[row][col]) > max_count) && nodeView[row][col] != RS_Completed {
						max_count = int(nodeView[row][col])
					}
				}
				myHallRequestView[row][col] = RequestState(max_count)
			case RS_NoOrder:
				// Go to RS_Pending if any other node has RS_Pending
				for _, nodeView := range remoteHallRequestView {
					if nodeView[row][col] == RS_Pending {
						myHallRequestView[row][col] = RS_Pending
						break
					}
				}
			case RS_Pending:
				pendingCount := 0
				for _, nodeView := range remoteHallRequestView {
					if nodeView[row][col] == RS_Confirmed {
						myHallRequestView[row][col] = RS_Confirmed
						break
					} else if nodeView[row][col] == RS_Pending {
						pendingCount++
					}
				}
				if pendingCount == len(remoteHallRequestView) {
					myHallRequestView[row][col] = RS_Confirmed
				}
			case RS_Confirmed:
				for _, nodeView := range remoteHallRequestView {
					// TODO: Check if or nodeView[row][col] == RS_Confirmed is needed
					if nodeView[row][col] == RS_Completed {
						myHallRequestView[row][col] = RS_NoOrder
						break
					}
				}
			case RS_Completed:
				// Go to RS_NoOrder if all other nodes have anything else than RS_Confirmed
				noOrderCount := 0
				for _, nodeView := range remoteHallRequestView {
					if nodeView[row][col] != RS_Confirmed {
						noOrderCount++
					}
				}
				if noOrderCount == len(remoteHallRequestView) {
					myHallRequestView[row][col] = RS_NoOrder
				}
			}
		}
	}
	return myHallRequestView
}

func (myNodeView *MyNodeView) ChangeNoOrderAndConfirmedToUnknown() {
	for row := 0; row < len(myNodeView.HallRequests); row++ {
		for col := 0; col < len(myNodeView.HallRequests[row]); col++ {
			if myNodeView.HallRequests[row][col] == RS_NoOrder || myNodeView.HallRequests[row][col] == RS_Confirmed {
				myNodeView.HallRequests[row][col] = RS_Unknown
			}
		}
	}
}

func (remoteRequestView *RemoteRequestView) InitRemoteRequestView() {
	remoteRequestView.RemoteHallRequestViews = make(map[string][][2]RequestState)
	remoteRequestView.RemoteCabRequests = make(map[string][]bool)
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
	ch_newHallRequest <-chan elevio.ButtonEvent,
	ch_completedHallRequest <-chan elevio.ButtonEvent,
	ch_elevState <-chan singleelevator.ElevState,
	ch_hallRequests chan<- [][2]bool,
	ch_remoteRequestView <-chan RemoteRequestView) {

	var myNodeView MyNodeView
	var isSingleElevMode = true

	myNodeView.InitMyNodeView()

	for {
		select {
		case remoteRequestView := <-ch_remoteRequestView:
			fmt.Println("nodeview: remoteRequestView")

			numRemoteNodes := len(ch_remoteRequestView)

			if numRemoteNodes > 0 {
				if isSingleElevMode {
					isSingleElevMode = false
					myNodeView.ChangeNoOrderAndConfirmedToUnknown()
				}
				myNodeView.HallRequests = updateMyHallRequestView(myNodeView.HallRequests, remoteRequestView.RemoteHallRequestViews)
				myNodeView.RemoteCabRequests = remoteRequestView.RemoteCabRequests
			} else {
				isSingleElevMode = true
			}

			ch_hallRequests <- convertHallRequestStateToBool(myNodeView.HallRequests, isSingleElevMode)

		case newHallRequest := <-ch_newHallRequest:
			fmt.Println("nodeview: newHallRequest")
			myNodeView.HallRequests[newHallRequest.Floor][int(newHallRequest.Button)] = RS_Pending

		case completedHallRequest := <-ch_completedHallRequest:
			fmt.Println("nodeview: completedHallRequest")
			nextRS := RS_Completed

			if isSingleElevMode {
				nextRS = RS_NoOrder
			}

			myNodeView.HallRequests[completedHallRequest.Floor][int(completedHallRequest.Button)] = nextRS

		case elevState := <-ch_elevState:
			fmt.Println("nodeview: elevState")
			myNodeView.ElevState = elevState

		case <-time.After(50 * time.Millisecond):
			fmt.Println("nodeview: broadcaster myNodeView")
			ch_sendMyNodeView <- myNodeView

		//default:
			//time.Sleep(100*time.Millisecond)
		}
		time.Sleep(50*time.Millisecond)
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
