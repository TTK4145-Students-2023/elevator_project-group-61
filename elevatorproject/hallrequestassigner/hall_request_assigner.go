//NOTES FOR THIS MODULE

//Takes uses two channels
//ch_hraoutput of type [][2]bool
//ch_hraInput of type SystemAwareness struct

//TODO: In network module, implement I'm available channel, which truns off broadcasting for the unavailable node.

package hallrequestassigner

import (
	"elevatorproject/config"
	"elevatorproject/systemview"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

type HRAElevState struct {
	Behaviour   string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"` //Dont need this for ElevState
}

// The HallRequests are all the requests in the system, but from this nodes point of view.
type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

// TODO: get localID from somewhere
func transformToHRAInput(systemAwareness systemview.SystemAwareness) HRAInput {
	transfromedHRAHallRequests := make([][2]bool, len(systemAwareness.SystemHallRequests[config.LocalID]))
	systemHallRequests := systemAwareness.SystemHallRequests[config.LocalID]
	for i, floor := range systemHallRequests {
		for j, requestState := range floor {
			if requestState == systemview.RS_Confirmed {
				transfromedHRAHallRequests[i][j] = true
			} else if requestState == systemview.RS_Pending && len(systemAwareness.SystemElevState) == 1 {
				transfromedHRAHallRequests[i][j] = true
			} else {
				transfromedHRAHallRequests[i][j] = false
			}
		}
	}

	transfromedHRAStates := make(map[string]HRAElevState)
	systemElevState := systemAwareness.SystemElevState
	//systemCabRequests := systemAwareness.SystemCabRequests
	systemNodesAvailable := systemAwareness.SystemNodesAvailable
	for id, elevState := range systemElevState {
		if systemNodesAvailable[id] {
			newHRAElevState := HRAElevState{
				Behaviour:   elevState.Behaviour,
				Floor:       elevState.Floor,
				Direction:   elevState.Direction,
				CabRequests: elevState.CabRequests,
			}
			transfromedHRAStates[id] = newHRAElevState
		}
	}

	transfromedHRAInput := HRAInput{
		HallRequests: transfromedHRAHallRequests,
		States:       transfromedHRAStates,
	}
	return transfromedHRAInput
}

func AssignHallRequests(ch_hraInput <-chan systemview.SystemAwareness, ch_hraoutput chan<- [][2]bool) {
	for {
		select {
		case systemAwareness := <-ch_hraInput:
			fmt.Println("HRA has gotten input\n")
			hraInput := transformToHRAInput(systemAwareness)

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

			output := new(map[string][][2]bool)
			err = json.Unmarshal(ret, &output)
			if err != nil {
				fmt.Println("json.Unmarshal error: ", err)
				return
			}
			hraOutput := (*output)[config.LocalID] //TODO: Get the local ID from somewhere

			ch_hraoutput <- hraOutput
		}
	}
}
