package main

import (
	"flag"
	"fmt"
	"github.com/joostvdg/boom/api"
	"net"
	"time"
)

func main() {
	helloPortOverride := flag.String("helloPort", api.HelloPort, fmt.Sprintf("Port number for listening for Hello messages, default %s", api.HelloPort))
	flag.Parse()

	startHelloServer(*helloPortOverride)
}

var members map[string]*api.Member
var memberHello = make(chan *api.Member)
var membersLock = make(chan struct{}, 1)

func startHelloServer(port string) {
	s, err := net.ResolveUDPAddr("udp4", ":" + port)
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
	members = make(map[string]*api.Member)
	go handleMember()

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
			memberHello <-member
		}
	}
}

func handleMember() {
	for {
		select {
			case member:= <-memberHello:
				member.LastSeen = time.Now()
				membersLock <-struct{}{} //acquire token
				if members[member.Identifier()] == nil {
					fmt.Printf("Received Hello from new Member: %+v\n", member)
				} else {
					lastSeenInfo := members[member.Identifier()]
					durationSinceLastSeen := member.LastSeen.Sub(lastSeenInfo.LastSeen)
					fmt.Printf("Received Hello from known Member: %s, first message since: %v\n", member.Identifier(), durationSinceLastSeen)
				}
				members[member.Identifier()] = member
				<-membersLock //release token
		}
	}
}


