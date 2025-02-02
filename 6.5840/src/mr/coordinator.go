package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// const values
const (
	IDLE       string = "IDLE"
	INPROGRESS string = "INPROGRESS"
	FAILD      string = "FAILD"
	DONE       string = "DONE"
)

var timeOut = time.Second * 10

// Coordinator struct
type Coordinator struct {
	// Your definitions here.
	files       []string
	mapAssigns  map[int]*TaskInfo
	reduAssigns map[int]*TaskInfo
	numReduce   int
	done        bool
	mu          sync.Mutex
	wg          sync.WaitGroup
}

// task information struct
type TaskInfo struct {
	workerID   int
	filesNames []string
	status     string
	assignedAt time.Time
}

// Your code here -- RPC handlers for the worker to call.

// assign task method
func (c *Coordinator) AssignTask(taskRequest *TaskRequest, taskAssign *TaskAssign) error {
	fmt.Println("Coordinator Started...")
	c.mu.Lock()
	defer c.mu.Unlock()

	// Assign map tasks
	for taskID, taskInfo := range c.mapAssigns {
		if taskInfo.status == IDLE || taskInfo.status == FAILD {
			taskInfo.workerID = taskRequest.WorkerID
			taskInfo.status = INPROGRESS
			taskInfo.assignedAt = time.Now()

			taskAssign.TaskID = taskID
			taskAssign.FilesNames = taskInfo.filesNames
			taskAssign.Type = MAP
			taskAssign.NumRed = c.numReduce
			log.Printf(
				"Map task assigned to workerID: %d , taskID: %d ",
				taskInfo.workerID,
				taskAssign.TaskID,
			)
			return nil
		}
	}

	// Assign reduce tasks
	if c.mapDone() {
		for taskID, taskInfo := range c.reduAssigns {
			if taskInfo.status == IDLE || taskInfo.status == FAILD {
				taskInfo.workerID = taskRequest.WorkerID
				taskInfo.status = INPROGRESS
				taskInfo.assignedAt = time.Now()
				taskAssign.FilesNames = taskInfo.filesNames
				taskAssign.Type = REDUCE
				taskAssign.TaskID = taskID
				taskAssign.NumRed = c.numReduce

				log.Printf(
					"Reduce task assigned to workerID: %d , taskID: %d ",
					taskInfo.workerID,
					taskAssign.TaskID,
				)
				return nil
			}
		}
	}
	return nil
}

// task done method
func (c *Coordinator) TaskDone(taskDoneNotif *TaskDoneNotif, taskDone *TaskDone) error {
	if taskDoneNotif.Type == MAP {
		log.Printf(
			"receivied task done notifi from workerid: %d and taskid: %d \n",
			taskDoneNotif.WorkerID,
			taskDoneNotif.TaskID,
		)
		taskInfo := c.mapAssigns[taskDoneNotif.TaskID]
		c.mu.Lock()
		if taskInfo.status != DONE {
			taskInfo.status = DONE
			c.updateReduce(taskDoneNotif.FilesNames)
		}
		taskDone.Done = true
		c.mu.Unlock()

		log.Printf(
			"receivied map done from workerid: %d and taskid: %d \n",
			taskDoneNotif.WorkerID,
			taskDoneNotif.TaskID,
		)
	} else {
		taskinfo := c.reduAssigns[taskDoneNotif.TaskID]
		c.mu.Lock()
		taskinfo.status = DONE
		c.mu.Unlock()

		log.Printf(
			"receivied reduce done from workerid: %d and taskid: %d \n",
			taskDoneNotif.WorkerID,
			taskDoneNotif.TaskID,
		)
		taskDone.Done = true
	}
	return nil
}

// update reduce tasks when map tasks finished
func (c *Coordinator) updateReduce(mapFilesName []string) {
	for _, filename := range mapFilesName {
		reduceID := getReduceID(filename)
		if r, ok := c.reduAssigns[reduceID]; ok {
			r.filesNames = append(r.filesNames, filename)
		} else {
			c.reduAssigns[reduceID] = createReduceTask(filename, r.workerID)
		}
	}
}

// check inprgress tasks
func (c *Coordinator) checkInProgressTasks() {
	for {

		if c.done {
			break
		}
		now := time.Now()
		c.mu.Lock()
		// check map tasks
		for taskID, taskInfo := range c.mapAssigns {
			if taskInfo.status == INPROGRESS {
				if now.Sub(taskInfo.assignedAt).Seconds() > timeOut.Seconds() {
					taskInfo.status = FAILD
					log.Printf(
						"Map task with ID: %d STUCK! in workerID: %d ",
						taskID,
						taskInfo.workerID,
					)
				}
			}
		}
		// check reduce tasks

		for taskID, taskInfo := range c.reduAssigns {
			if taskInfo.status == INPROGRESS {
				if now.Sub(taskInfo.assignedAt).Seconds() > timeOut.Seconds() {
					taskInfo.status = FAILD
					log.Printf(
						"Reduce task with ID: %d STUCK! in workerID: %d ",
						taskID,
						taskInfo.workerID,
					)
				}
			}
		}

		c.mu.Unlock()
		time.Sleep(time.Second * 10)
	}
	c.mu.Unlock()
}

// map done method
func (c *Coordinator) mapDone() bool {
	for _, m := range c.mapAssigns {
		if m.status != DONE {
			return false
		}
	}
	return true
}

// reduce done method
func (c *Coordinator) reduceDone() bool {
	for _, r := range c.reduAssigns {
		if r.status != DONE {
			return false
		}
	}
	return true
}

// create reduce task
func createReduceTask(filesName string, workerID int) *TaskInfo {
	taskInfo := TaskInfo{}
	taskInfo.filesNames = []string{filesName}
	taskInfo.status = IDLE
	taskInfo.workerID = workerID
	return &taskInfo
}

// get reduce task ID
func getReduceID(filesName string) int {
	regex := regexp.MustCompile("[0-9]+")
	matches := regex.FindAllString(filesName, -1)

	if len(matches) == 2 {
		id, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Fatalf("couldn't parse reduce task id %v \n", err)
		}
		return id
	}

	log.Fatalf("couldn't parse reduce task id  \n")
	return -1
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	// l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	// Your code here.
	c.mu.Lock()
	done := c.mapDone() && c.reduceDone()
	if done {
		c.done = true
		c.mu.Unlock()
		return true
	}
	c.mu.Unlock()
	return false
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}
	c.files = files
	c.numReduce = nReduce

	for i, file := range files {
		taskinfo := TaskInfo{}
		taskinfo.filesNames = []string{file}
		taskinfo.status = IDLE
		c.mapAssigns[i] = &taskinfo
	}

	c.reduAssigns = make(map[int]*TaskInfo)
	c.server()
	c.done = false
	go c.checkInProgressTasks()

	return &c
}
