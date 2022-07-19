package main

import (
	"flag"
	"fmt"
	"github.com/joostvdg/boom/api"
	"log"
	"net"
	"time"
)

func main() {
	helloPortOverride := flag.String("helloPort", api.HelloPort, fmt.Sprintf("Port number for listening for Hello messages, default %s", api.HelloPort))
	helloName := flag.String("helloName", "MySelf", "Name of this Boom server")
	flag.Parse()

	go listenForMulticast()
	go multicastExistance(*helloName)
	startHelloServer(*helloPortOverride)
}

var members map[string]*api.Member
var memberHello = make(chan *api.Member)
var memberHelloMulticast = make(chan *api.Member)
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

// https://github.com/dmichael/go-multicast
func listenForMulticast() {
	addr, err := net.ResolveUDPAddr("udp4", api.MEMBERSHIP_GROUP_ADDRESS)
	if err != nil {
		log.Fatal(err)
	}

	// Open up a connection
	connection, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	buffer := make([]byte, 1024)
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
			memberHelloMulticast <-member
		}
	}
}

func multicastExistance(name string) {
	for {
		serverAddress := api.MEMBERSHIP_GROUP_ADDRESS
		udpServer, err := net.ResolveUDPAddr("udp4", serverAddress)
		connection, err := net.ListenUDP("udp4", nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer connection.Close()
		localAddr := connection.LocalAddr().(*net.UDPAddr)
		fmt.Printf("Local Address: %s", localAddr)
		message := api.ConstructMembershipMessage(name, connection.LocalAddr().String())

		_, err = connection.WriteToUDP(message, udpServer)
		if err != nil {
			fmt.Printf("Received an error: %s", err)
			return
		}
		time.Sleep(10 * time.Second)
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
				<-membersLock //release tokenc
			case member:= <-memberHelloMulticast:
				fmt.Printf("Received Multicast from Member: %s @%v\n", member.MemberName, member.IP)
		}
	}
}


