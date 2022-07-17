package main

import (
	"fmt"
	"net"

	"github.com/joostvdg/boom/api"
)

func main() {
	startHelloServer()
}

func startHelloServer() {
	s, err := net.ResolveUDPAddr("udp4", api.HelloPort)
	if err != nil {
		fmt.Println(err)
		return
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer connection.Close()
	buffer := make([]byte, 1024)

	fmt.Printf("Listening on port %s for Hello messages...\n", api.HelloPort)
	for {
		numberOfBytes, address, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Received message %s from %s\n", string(buffer[0:numberOfBytes]), address.String())
		member, err := api.ReadMemberMessage(buffer[0:numberOfBytes])
		if err != nil {
			fmt.Printf("Ran into an error: %s\n", err)
		} else {
			fmt.Printf("Received Member Hello: %+v\n", member)
		}
	}
}


