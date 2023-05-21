package shotqueue

type ShotQueue interface {
	// DequeueShot returns a channel that will send an empty struct on each shot
	DequeueShot() <-chan struct{}

	// QueueShot queues a shot
	QueueShot()
}
