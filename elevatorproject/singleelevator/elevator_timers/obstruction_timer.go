package elevator_timers

import (
	"time"
)

const standardObstructionWait = 12000 // 12 seconds

var obstructionCounter int64 = -1

func StartObstructionTimer() {
	obstructionCounter = time.Now().UnixMilli()
}

func StopObstructionTimer() {
	obstructionCounter = -1
}

func GetObstructionCounter() int64 {
	return obstructionCounter
}

func CheckObstructionTimer(ch_timer chan int) {
	for {
		if obstructionCounter == -1 {
			continue
		}
		if time.Now().UnixMilli()-obstructionCounter > standardObstructionWait {
			ch_timer <- 1
			obstructionCounter = -1
		}
		time.Sleep(10 * time.Millisecond)
	}
}
