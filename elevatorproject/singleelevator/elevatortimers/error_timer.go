package elevatortimers

import "time"

const standardErrorWait = 12000 // 12 seconds

var errorCounter int64 = -1

func StartErrorTimer() {
	errorCounter = time.Now().UnixMilli()
}

func StopErrorTimer() {
	errorCounter = -1
}

func GetErrorCounter() int64 {
	return errorCounter
}

func CheckErrorTimer(ch_timer chan int) {
	for {
		if errorCounter == -1 {
			continue
		}
		if time.Now().UnixMilli()-errorCounter > standardErrorWait {
			ch_timer <- 1
			errorCounter = -1
		}
		time.Sleep(10 * time.Millisecond)
	}
}
