package server

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"github.com/google/uuid"
)

type server struct {
	serverID   string
	taskID     string
	workerName string
	termID     int
	state      string
	assignAt   time.Time
}

func NewServer() server {
	return server{}
}

func (s *server) MutateData(data *SendDataRPC) *server {
	server := server{}
	server.serverID = uuid.New().String()
	server.taskID = data.taskID
	server.assignAt = time.Now()
	server.workerName = data.workerName
	server.termID = data.termID
	server.state = "SEND"
	return &server
}

func (s *server) Call() {
	RPCServer := new(server)
	rpc.Register(RPCServer)
	rpc.HandleHTTP()
	port := ":3212"

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Couldn't listen to port", err)
	}
	http.Serve(lis, nil)
}
