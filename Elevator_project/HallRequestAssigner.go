
type HRAElevState struct {
    Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}

type HRAInput struct {
    HallRequests    [][2]bool                   `json:"hallRequests"`
    States          map[int]HRAElevState     `json:"states"`
}


func transformToHRAInput() HRAInput{
	for id, node_status := range nodeStatuses {
		
	}
}


type HRAInfo struct {
	
}


func transformMap(inputMap map[string]int) map[string]int {
    outputMap := make(map[string]int)

    for key, value := range inputMap {
        // Perform some transformation on the value
        newValue := value * 2

        // Add the transformed key-value pair to the output map
        outputMap[key] = newValue
    }

    return outputMap
}