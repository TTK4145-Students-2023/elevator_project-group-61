//NOTES FOR THIS MODULE

//Takes uses two channels
//ch_hraoutput of type [][2]bool
//ch_hraInput of type SystemAwareness struct

//TODO: In network module, implement I'm available channel, which truns off broadcasting for the unavailable node.





type HRAElevState struct {
    Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"` //Dont need this for ElevState
}
 //The HallRequests are all the requests in the system, but from this nodes point of view.
type HRAInput struct {
    HallRequests    [][2]bool                   `json:"hallRequests"`
    States          map[string]HRAElevState     `json:"states"`
}

// TODO: get localID from somewhere
func transformToHRAInput(systemAwareness SystemAwareness, id string) HRAInput {
	transfromedHRAHallRequests := make([][2]bool, len(systemAwareness.SystemHallRequests[id]))
	systemHallRequests := systemAwareness.SystemHallRequests[id]
	for i, floor := range systemHallRequests {
		for j, requestState := range floor {
			if requestState == RS_Confirmed {
				transfromedHRAHallRequests[i][j] = true
			} else if requestState == RS_Pending && len(systemAwareness.SystemElevState) == 1 {
				transfromedHRAHallRequests[i][j] = true
			} else {
				transfromedHRAHallRequests[i][j] = false
			}
		}
	}

	transfromedHRAStates := make(map[string]HRAElevState)
	systemElevState := systemAwareness.SystemElevState
	systemCabRequests := systemAwareness.SystemCabRequests
	for id, elevState := range systemElevState {
		newHRAElevState := HRAElevState{
			Behaviour:   elevState.Behaviour,
			Floor:       elevState.Floor,
			Direction:   elevState.Direction,
			CabRequests: systemCabRequests[id],
		}
		transfromedHRAStates[id] = newHRAElevState
	}

	transfromedHRAInput := HRAInput{
		HallRequests: transfromedHRAHallRequests,
		States:       transfromedHRAStates,
	}
	return transfromedHRAInput
}l goroutines are asleep - deadlock!



func assignHallRequests(ch_hraInput <-chan SystemAwareness, ch_hraoutput chan<- [][2]bool, id string) {
	for {
		select {
		case systemAwareness := <-ch_hraInput:
			hraInput := transformToHRAInput(systemAwareness, id)

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

			ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
			if err != nil {
				fmt.Println("exec.Command error: ", err)
				fmt.Println(string(ret))
				returnl goroutines are asleep - deadlock!

			}

			output := new(map[string][][2]bool)
			err = json.Unmarshal(ret, &output)
			if err != nil {
				fmt.Println("json.Unmarshal error: ", err)
				return
			}
			hraOutput := (*output)[id] //TODO: Get the local ID from somewhere

			ch_hraoutput <- hraOutput
		}
	}
}