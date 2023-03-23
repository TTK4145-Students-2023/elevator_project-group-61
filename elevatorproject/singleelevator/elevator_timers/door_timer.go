package elevator_timers

import (
	"time"
)

const standardDoorWait = 3000 // 3 seconds
var doorCounter int64 = -1

func StartDoorTimer() {
	doorCounter = time.Now().UnixMilli()
}

func CheckDoorTimer(ch_timer chan int) {
	for {
		if doorCounter == -1 {
			continue
		}
		if time.Now().UnixMilli()-doorCounter > standardDoorWait {
			ch_timer <- 1
			doorCounter = -1
		}
		time.Sleep(10 * time.Millisecond)
	}
}

