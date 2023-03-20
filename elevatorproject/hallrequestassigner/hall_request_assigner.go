//NOTES FOR THIS MODULE

//Takes uses two channels
//ch_hraoutput of type [][2]bool
//ch_hraInput of type SystemAwareness struct

//TODO: In network module, implement I'm available channel, which truns off broadcasting for the unavailable node.

package hallrequestassigner

import (
	"elevatorproject/config"
	"elevatorproject/worldview"
	"elevatorproject/nodeview"
	"encoding/json"
	"fmt"
	"time"
	"os/exec"
	"runtime"
	"io/ioutil"
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
func transformToHRAInput(myWorldView worldview.MyWorldView) HRAInput {
	transfromedHRAHallRequests := make([][2]bool, config.NumFloors)
	systemHallRequests := myWorldView.HallRequestView
	for i, floor := range systemHallRequests {
		for j, requestState := range floor {
			if requestState == nodeview.RS_Confirmed {
				transfromedHRAHallRequests[i][j] = true
			} else if requestState == nodeview.RS_Pending && len(myWorldView.ElevStates) == 1 {
				transfromedHRAHallRequests[i][j] = true
			} else {
				transfromedHRAHallRequests[i][j] = false
			}
		}
	}

	transfromedHRAStates := make(map[string]HRAElevState)
	systemElevState := myWorldView.ElevStates
	//systemCabRequests := systemAwareness.SystemCabRequests
	systemNodesAvailable := myWorldView.NodesAvailable
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

func diffHRARequests(oldHRARequests [][2]bool, newHRARequests [][2]bool) bool {
	for i := 0; i < config.NumFloors; i++ {
		for j := 0; j < 2; j++ {
			if oldHRARequests[i][j] != newHRARequests[i][j] {
				return true
			}
		}
	}
	return false
}

func AssignHallRequests(ch_hraInput <-chan worldview.MyWorldView, ch_hraoutput chan<- [][2]bool) {
	oldHRARequests := make([][2]bool, config.NumFloors)
	for {
		select {
		case systemAwareness := <-ch_hraInput:
			fmt.Println("HRA has gotten input")
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

			// Print out json to file
			filename := "hra_assigner_input" + config.LocalID
			err = ioutil.WriteFile(filename, jsonBytes, 0644)
			if err != nil {
				panic(err)
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

			// printer output
			
			// fmt.Printf("output: \n")
    		// for k, v := range *output {
        	// 	fmt.Printf("%6v :  %+v\n", k, v)
    		// }
			if diffHRARequests(oldHRARequests, hraOutput) {
				ch_hraoutput <- hraOutput
				oldHRARequests = hraOutput
			}
		default:
			time.Sleep(50*time.Millisecond)
		}
	}
}
