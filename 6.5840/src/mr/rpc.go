package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//
// Add your RPC definitions here.

// const values

const (
	MAP    string = "MAP"
	REDUCE string = "REDUCE"
)

// task request struct
type TaskRequest struct {
	WorkerID int
}

// task assignment struct
type TaskAssign struct {
	TaskID     int
	Type       string
	FilesNames []string
	NumRed     int
}

// task done notification struct
type TaskDoneNotif struct {
	WorkerID   int
	FilesNames []string
	TaskID     int
	Type       string
}

// task done ack struct
type TaskDone struct {
	Done bool
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
