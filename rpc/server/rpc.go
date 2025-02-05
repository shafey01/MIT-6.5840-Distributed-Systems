package server

import "time"

type SendDataRPC struct {
	taskID     string
	workerName string
	termID     int
	state      string
	assignAt   time.Time
}
