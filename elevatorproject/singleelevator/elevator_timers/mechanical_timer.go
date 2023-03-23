package elevator_timers

import (
	"time"
)

const standardMechanicalWait = 12000 // 12 seconds

var mechanicalCounter int64 = -1

func StartMechanicalTimer() {
	obstructionCounter = time.Now().UnixMilli()
}

func StopMechanicalTimer() {
	obstructionCounter = -1
}

func CheckMechanicalTimer(ch_timer chan int) {
	for {
		if mechanicalCounter == -1 {
			continue
		}
		if time.Now().UnixMilli()-mechanicalCounter > standardMechanicalWait {
			ch_timer <- 1
			mechanicalCounter = -1
		}
		time.Sleep(10 * time.Millisecond)
	}
}
