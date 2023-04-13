//NOTES FOR THIS MODULE

//Takes uses two channels
//ch_hraoutput of type [][2]bool
//ch_hraInput of type SystemAwareness struct

//TODO: In network module, implement I'm available channel, which truns off broadcasting for the unavailable node.

package requestassigner

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
	transformedHRACabRequests := make(map[string][]bool)
	systemHallRequests := myWorldView.HallRequestView
	systemCabRequests := myWorldView.CabRequests
	for i, floor := range systemHallRequests {
		for j, requestState := range floor {
			if requestState == nodeview.RS_Confirmed {
				transfromedHRAHallRequests[i][j] = true
			} else {
				transfromedHRAHallRequests[i][j] = false
			}
		}
	}
	for id, requestStates := range systemCabRequests {
		for j, requestState := range requestStates {
			if requestState == nodeview.RS_Confirmed {
				transformedHRACabRequests[id][j] = true
			} else {
				transformedHRACabRequests[id][j] = false
			}
		}
	}
		
	transfromedHRAStates := make(map[string]HRAElevState)
	systemElevState := myWorldView.ElevStates
	for id, elevState := range systemElevState {
		if elevState.IsAvailable {
			newHRAElevState := HRAElevState{
				Behaviour:   elevState.Behaviour,
				Floor:       elevState.Floor,
				Direction:   elevState.Direction,
				CabRequests: transformedHRACabRequests[id],
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

func diffHallRequests(oldHallRequests [][2]bool, newHallRequests [][2]bool) bool {
	for i := 0; i < config.NumFloors; i++ {
		for j := 0; j < 2; j++ {
			if oldHallRequests[i][j] != newHallRequests[i][j] {
				return true
			}
		}
	}
	return false
}

func diffCabRequests(oldCabRequests []bool, newCabRequests []bool) bool {
	for i := 0; i < config.NumFloors; i++ {
		if oldCabRequests[i] != newCabRequests[i] {
			return true
		}
	}
	return false
}

func AssignRequests(ch_hraInput <-chan worldview.MyWorldView, ch_hallRequest chan<- [][2]bool, ch_cabRequests chan<- []bool, localID string) {
	oldHallRequests := make([][2]bool, config.NumFloors)
	for i := 0; i < config.NumFloors; i++ {
		for j := 0; j < 2; j++ {
			oldHallRequests[i][j] = true
		}
	}
	oldCabRequests := make([]bool, config.NumFloors)
	for i := 0; i < config.NumFloors; i++ {
		oldCabRequests[i] = true
	}

	for {
		select {
		case myWorldView := <-ch_hraInput:
			//fmt.Println("HRA has gotten input")
			if !myWorldView.ElevStates[localID].IsAvailable {
				continue
			}
			hraInput := transformToHRAInput(myWorldView)

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
			filename := "hra_assigner_input" + localID
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

			hraOutput := new(map[string][][2]bool)
			err = json.Unmarshal(ret, &hraOutput)
			if err != nil {
				fmt.Println("json.Unmarshal error: ", err)
				return
			}
			hallRequests := (*hraOutput)[localID] //TODO: Get the local ID from somewhere
			cabRequests := hraInput.States[localID].CabRequests

			// printer output
			
			// fmt.Printf("output: \n")
    		// for k, v := range *output {
        	// 	fmt.Printf("%6v :  %+v\n", k, v)
    		// }
			if diffHallRequests(oldHallRequests, hallRequests) {
				ch_hallRequest <- hallRequests
				copy(oldHallRequests, hallRequests)
			}
			if diffCabRequests(oldCabRequests, cabRequests) {
				ch_cabRequests <- cabRequests
				copy(oldCabRequests, cabRequests)
			}
		default:
			time.Sleep(50*time.Millisecond)
		}
	}
}
