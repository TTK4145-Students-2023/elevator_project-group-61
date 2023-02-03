package door_timer

import "time"
import "fmt"

const StandardDoorWait = 3000

var TimeCounter int64

func StartTimer() {
	TimeCounter = time.Now().UnixMilli()
}

func CheckTimer() bool {
	DiffTime := time.Now().UnixMilli() - TimeCounter
	fmt.Println("timediff:", DiffTime)
	return DiffTime > StandardDoorWait // Milliseconds, so need to be 3000
}
