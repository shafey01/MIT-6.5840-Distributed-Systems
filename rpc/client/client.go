package rpc

import (
	"log"
	"net/rpc"
)

func CallServer(host string, port string, sdr *SendDataRPC, result *server) *server {
	client, err := rpc.DialHTTP("tcp", host+port)
	if err != nil {
		log.Fatal("Dialing Error: ", err)
	}
	err = client.Call("RPCServer.MutateData", &sdr, &result)
	if err != nil {
		log.Fatal("Call MutateData Error: ", err)
	}

	return result
}
