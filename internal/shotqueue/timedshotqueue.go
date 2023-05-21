package shotqueue

import "time"

type TimedShotQueue struct {
	shotQueue chan struct{}
}

var _ ShotQueue = (*TimedShotQueue)(nil)

func New(frequency time.Duration) *TimedShotQueue {
	tsq := &TimedShotQueue{
		shotQueue: make(chan struct{}, 1),
	}

	go tsq.queueTimedShots(frequency)

	return tsq
}

func (tsq *TimedShotQueue) queueTimedShots(frequency time.Duration) {
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()

	for range ticker.C {
		tsq.QueueShot()
	}
}

func (tsq *TimedShotQueue) QueueShot() {
	tsq.shotQueue <- struct{}{}
}

func (tsq *TimedShotQueue) DequeueShot() <-chan struct{} {
	return tsq.shotQueue
}
