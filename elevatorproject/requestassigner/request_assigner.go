//NOTES FOR THIS MODULE

//Takes uses two channels
//ch_hraoutput of type [][2]bool
//ch_hraInput of type SystemAwareness struct

//TODO: In network module, implement I'm available channel, which truns off broadcasting for the unavailable node.

package requestassigner

import (
	"elevatorproject/config"
	"elevatorproject/nodeview"
	"elevatorproject/worldview"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"runtime"
	"time"
)

type HRAElevState struct {
	Behaviour   string                 `json:"behaviour"`
	Floor       int                    `json:"floor"`
	Direction   string                 `json:"direction"`
	CabRequests [config.NumFloors]bool `json:"cabRequests"` //Dont need this for ElevState
}

// The HallRequests are all the requests in the system, but from this nodes point of view.
type HRAInput struct {
	HallRequests [config.NumFloors][2]bool `json:"hallRequests"`
	States       map[string]HRAElevState   `json:"states"`
}

// TODO: get localID from somewhere
func transformToHRAInput(myWorldView worldview.MyWorldView, localID string) HRAInput {
	//make array of bools for each floor, and then make a map of those arrays
	transfromedHRAHallRequests := [config.NumFloors][2]bool{}
	transformedHRACabRequests := make(map[string][config.NumFloors]bool, config.NumElevators)
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
		transformedRequestStates := [config.NumFloors]bool{}
		for floor, requestState := range requestStates {
			if requestState == nodeview.RS_Confirmed {
				transformedRequestStates[floor] = true
			} else {
				transformedRequestStates[floor] = false
			}
		}
		transformedHRACabRequests[id] = transformedRequestStates
	}

	transfromedHRAStates := make(map[string]HRAElevState)
	systemElevState := myWorldView.ElevStates
	for id, elevState := range systemElevState {
		if elevState.IsAvailable || id == localID {
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

func diffHallRequests(oldHallRequests [config.NumFloors][2]bool, newHallRequests [config.NumFloors][2]bool) bool {
	for i := 0; i < config.NumFloors; i++ {
		for j := 0; j < 2; j++ {
			if oldHallRequests[i][j] != newHallRequests[i][j] {
				return true
			}
		}
	}
	return false
}

func diffCabRequests(oldCabRequests [config.NumFloors]bool, newCabRequests [config.NumFloors]bool) bool {
	for i := 0; i < config.NumFloors; i++ {
		if oldCabRequests[i] != newCabRequests[i] {
			return true
		}
	}
	return false
}

func AssignRequests(ch_hraInput <-chan worldview.MyWorldView, ch_hallRequest chan<- [config.NumFloors][2]bool, ch_cabRequests chan<- [config.NumFloors]bool, localID string) {
	oldHallRequests := [config.NumFloors][2]bool{}
	oldCabRequests := [config.NumFloors]bool{}

	for {
		select {
		case myWorldView := <-ch_hraInput:
			//fmt.Println("HRA has gotten input")
			//if !myWorldView.ElevStates[localID].IsAvailable && {
			//continue
			//}
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

			hraOutput := new(map[string][config.NumFloors][2]bool)
			err = json.Unmarshal(ret, &hraOutput)
			if err != nil {
				fmt.Println("json.Unmarshal error: ", err)
				return
			}
			hallRequests := (*hraOutput)[localID]
			cabRequests := hraInput.States[localID].CabRequests

			// printer output

			// fmt.Printf("output: \n")
			// for k, v := range *output {
			// 	fmt.Printf("%6v :  %+v\n", k, v)
			// }
			if diffHallRequests(oldHallRequests, hallRequests) {
				ch_hallRequest <- hallRequests
				oldHallRequests = hallRequests
			}
			if diffCabRequests(oldCabRequests, cabRequests) {
				ch_cabRequests <- cabRequests
				oldCabRequests = cabRequests
			}
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}
