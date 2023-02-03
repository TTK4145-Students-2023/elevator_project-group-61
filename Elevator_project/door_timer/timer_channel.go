package door_timer

import "time"

// It is assumed that this function is run as a go routine, (because if not then
// it terminates if the function it is called from terminates?)
// Will send 1 on channel if 3 seconds have passed since last call of a timer start.
// I guess this will also mean that the receiver side must be a go routine function.

// Can remember that I have direct access to other variables in same package, for example StandardDoorWait

func DoorTimerChannel(door_ch chan int) {
	start_time := time.Now().UnixMilli()
	for {
		if time.Now().UnixMilli()-start_time > StandardDoorWait {
			door_ch <- 1
		}
	}
}
