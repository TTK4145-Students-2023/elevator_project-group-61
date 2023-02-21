package door_timer

import (
	"time"
	"fmt"
)

const StandardDoorWait = 3000

var TimeCounter int64 = -1

func StartTimer() {
	TimeCounter = time.Now().UnixMilli()
}

func CheckTimer(ch_timer chan int) {
	fmt.Println("Timer check started")
	for {
		if TimeCounter == -1 {
			continue
		}
		if time.Now().UnixMilli()-TimeCounter > StandardDoorWait {
			fmt.Println("Door timer expired")
			ch_timer <- 1
			TimeCounter = -1
		}
		time.Sleep(10 * time.Millisecond)
	}
}
