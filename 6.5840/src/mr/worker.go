package mr

import (
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"

	"github.com/google/uuid"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

var workerID string

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string,
) {
	workerID = uuid.New().String()
}

// run mapper func

// run reducer func

// get task func
func getTask() (*TaskAssign, error) {
	taskAssign := TaskAssign{}
	taskRequest := TaskRequest{}
	taskRequest.WorkerID = workerID
	if call("coordinator.AssignTask", &taskRequest, &taskAssign) {
		return &taskAssign, nil
	} else {
		return nil, errors.New("can not call the coordinator")
	}
}

// task done func
func taskDone(filenames []string, taskId int, taskType string) {
	taskDoneNotif := TaskDoneNotif{}
	taskDoneNotif.FilesNames = filenames
	taskDoneNotif.WorkerID = workerID
	taskDoneNotif.TaskID = taskId
	taskDoneNotif.Type = taskType
	taskDone := TaskDone{}
	if call("coordinator.TaskDone", &taskDoneNotif, &taskDone) {
		if !taskDone.Done {
			log.Fatal("coordinator didn't response with Done")
		} else {
			log.Fatal("couldn't notifiy coordinator ")
		}
	}
}

// read file func
// task assigned func
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
