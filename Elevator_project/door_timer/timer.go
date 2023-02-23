package door_timer

import (
	"time"
)

const StandardDoorWait = 3000

var TimeCounter int64 = -1

func StartTimer() {
	TimeCounter = time.Now().UnixMilli()
}

func CheckTimer(ch_timer chan int) {
	for {
		if TimeCounter == -1 {
			continue
		}
		if time.Now().UnixMilli()-TimeCounter > StandardDoorWait {
			ch_timer <- 1
			TimeCounter = -1
		}
		time.Sleep(10 * time.Millisecond)
	}
}
