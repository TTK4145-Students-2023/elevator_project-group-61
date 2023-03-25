package elevator_timers

import (
	"time"
)

// Hele denne er lagt til

const standardSpamWait = 12000 // 12 seconds

var spamCounter int64 = -1

func StartSpamTimer() {
	spamCounter = time.Now().UnixMilli()
}

func StopSpamTimer() {
	spamCounter = -1
}

func GetSpamCounter() int64 {
	return spamCounter
}

func CheckSpamTimer(ch_timer chan int) {
	for {
		if spamCounter == -1 {
			continue
		}
		if time.Now().UnixMilli()-spamCounter > standardSpamWait {
			ch_timer <- 1
			spamCounter = -1
		}
		time.Sleep(10 * time.Millisecond)
	}
}