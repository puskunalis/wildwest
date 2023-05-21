package shootoutstarter

import "time"

type DefaultShootoutStarter struct {
	shootoutTime chan time.Time
}

func New() *DefaultShootoutStarter {
	return &DefaultShootoutStarter{
		shootoutTime: make(chan time.Time, 1),
	}
}

func (ss *DefaultShootoutStarter) ReceiveShootoutTime(t time.Time) {
	ss.shootoutTime <- t
}

func (ss *DefaultShootoutStarter) WaitForShootout() {
	time.Sleep(time.Until(<-ss.shootoutTime))
}
