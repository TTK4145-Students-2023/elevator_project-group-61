package requestassigner

import (
	"elevatorproject/config"
	"elevatorproject/worldview"
	"elevatorproject/peerview"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

type HRAElevState struct {
	Behaviour   string                 `json:"behaviour"`
	Floor       int                    `json:"floor"`
	Direction   string                 `json:"direction"`
	CabRequests [config.NumFloors]bool `json:"cabRequests"` 
}

type HRAInput struct {
	HallRequests [config.NumFloors][2]bool `json:"hallRequests"`
	States       map[string]HRAElevState   `json:"states"`
}

func transformToHRAInput(myWorldView worldview.MyWorldView, localID string) HRAInput {
	transfromedHRAHallRequests := [config.NumFloors][2]bool{}
	transformedHRACabRequests := make(map[string][config.NumFloors]bool, config.NumElevators)
	allHallRequests := myWorldView.HallRequestView
	allCabRequests := myWorldView.CabRequests

	for floor, requestStates := range allHallRequests {
		for button, requestState := range requestStates {
			if requestState == peerview.RS_Confirmed {
				transfromedHRAHallRequests[floor][button] = true
			} else {
				transfromedHRAHallRequests[floor][button] = false
			}
		}
	}
	for id, requestStates := range allCabRequests {
		transformedRequestStates := [config.NumFloors]bool{}
		for floor, requestState := range requestStates {
			if requestState == peerview.RS_Confirmed {
				transformedRequestStates[floor] = true
			} else {
				transformedRequestStates[floor] = false
			}
		}
		transformedHRACabRequests[id] = transformedRequestStates
	}

	hraStates := make(map[string]HRAElevState)
	allElevStates := myWorldView.ElevStates
	for id, elevState := range allElevStates {
		if elevState.IsAvailable || id == localID {
			newHRAElevState := HRAElevState{
				Behaviour:   elevState.Behaviour,
				Floor:       elevState.Floor,
				Direction:   elevState.Direction,
				CabRequests: transformedHRACabRequests[id],
			}
			hraStates[id] = newHRAElevState
		}
	}

	hraInput := HRAInput{
		HallRequests: transfromedHRAHallRequests,
		States:       hraStates,
	}
	return hraInput
}

func hallRequestsChanged(oldHallRequests [config.NumFloors][2]bool, newHallRequests [config.NumFloors][2]bool) bool {
	for i := 0; i < config.NumFloors; i++ {
		for j := 0; j < 2; j++ {
			if oldHallRequests[i][j] != newHallRequests[i][j] {
				return true
			}
		}
	}
	return false
}

func cabRequestsChanged(oldCabRequests [config.NumFloors]bool, newCabRequests [config.NumFloors]bool) bool {
	for i := 0; i < config.NumFloors; i++ {
		if oldCabRequests[i] != newCabRequests[i] {
			return true
		}
	}
	return false
}

func AssignRequests(ch_myWorldView <-chan worldview.MyWorldView, ch_hallRequest chan<- [config.NumFloors][2]bool, ch_cabRequests chan<- [config.NumFloors]bool, localID string) {
	oldHallRequests := [config.NumFloors][2]bool{}
	oldCabRequests := [config.NumFloors]bool{}

	for {
		select {
		case myWorldView := <-ch_myWorldView:
			
			hraInput := transformToHRAInput(myWorldView, localID)

			hraExecutable := ""
			switch runtime.GOOS {
			case "linux":
				hraExecutable = "hall_request_assigner"
			case "windows":
				hraExecutable = "hall_request_assigner.exe"
			default:
				panic("OS not supported")
			}

			jsonBytes, err := json.Marshal(hraInput)
			if err != nil {
				fmt.Println("json.Marshal error: ", err)
				return
			}

			ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
			if err != nil {
				fmt.Println("exec.Command error: ", err)
				fmt.Println(string(ret))
				return
			}

			hraOutput := new(map[string][config.NumFloors][2]bool)
			err = json.Unmarshal(ret, &hraOutput)
			if err != nil {
				fmt.Println("json.Unmarshal error: ", err)
				return
			}

			hallRequests := (*hraOutput)[localID]
			cabRequests := hraInput.States[localID].CabRequests

			if hallRequestsChanged(oldHallRequests, hallRequests) {
				ch_hallRequest <- hallRequests
				oldHallRequests = hallRequests
			}
			if cabRequestsChanged(oldCabRequests, cabRequests) {
				ch_cabRequests <- cabRequests
				oldCabRequests = cabRequests
			}
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}
