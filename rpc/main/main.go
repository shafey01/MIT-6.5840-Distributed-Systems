package main

import (
	"fmt"

	s "github.com/shafey01/MIT-6.006-Algorithems/rpc/server"
)

func main() {
	fmt.Println("GO")
	server := s.NewServer()
	server.Call()
}
