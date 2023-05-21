package shootoutstarter

import (
	"time"
)

type ShootoutStarter interface {
	ReceiveShootoutTime(time.Time)
	WaitForShootout()
}
