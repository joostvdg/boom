package server

import (
	"fmt"
	"github.com/joostvdg/boom/api"
	"log"
	"net"
	"time"
)

func StartMembershipServer(serviceContext *MembershipServiceContext) {
	port := serviceContext.ServerPort
	ctx := serviceContext.Context
	listenAddress := serviceContext.Self.IPSelf.String()
	s, err := net.ResolveUDPAddr(api.MembershipNetwork, listenAddress+":"+port)
	if err != nil {
		fmt.Println(err)
		return
	}

	connection, err := net.ListenUDP(api.MembershipNetwork, s)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer connection.Close()
	defer fmt.Println("Exiting StartMembershipServer")
	SetReadDeadlineOnCancel(ctx, connection)

	buffer := make([]byte, 1024)
	members = make(map[string]*api.Member)

	fmt.Printf("Listening on port %s for Hello & Goodbye messages...\n", port)
	for {
		numberOfBytes, address, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Encountered an error reading from UDP connection: %s\n", err)
			return
		}
		fmt.Printf("Received message %s from %s\n", string(buffer[0:numberOfBytes]), address.String())
		member, messageType, err := api.ReadMemberMessage(buffer[0:numberOfBytes], address)
		if err != nil {
			fmt.Printf("Encountered an error determining message type: %s\n", err)
		} else {
			switch messageType.Prefix {
			case api.HelloPrefix:
				memberHello <- member
			case api.GoodbyePrefix:
				memberGoodbye <- member
			default:
				fmt.Println("Ran into an error, unknown message type")
			}

		}

	}
}

// https://github.com/dmichael/go-multicast
func ListenForMulticast(serviceContext *MembershipServiceContext) {
	ctx := serviceContext.Context
	addr, err := net.ResolveUDPAddr(api.MembershipNetwork, api.MembershipGroupAddress)
	if err != nil {
		log.Fatal(err)
	}

	// Open up a connection
	connection, err := net.ListenMulticastUDP(api.MembershipNetwork, nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	defer fmt.Println("Exiting ListenForMulticast")
	SetReadDeadlineOnCancel(ctx, connection)

	buffer := make([]byte, 1024)
	for {
		numberOfBytes, originAddress, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Received an error: %s\n", err)
			return
		}

		// TODO: should we treat this type of message differently?
		member, _, err := api.ReadMemberMessage(buffer[0:numberOfBytes], originAddress)
		if err != nil {
			fmt.Printf("Ran into an error: %s\n", err)
		} else {
			memberHelloMulticast <- member
		}
	}
}

func MulticastExistence(serviceContext *MembershipServiceContext) {
	ctx := serviceContext.Context
	message := serviceContext.HelloMessage
	clock := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-clock.C:
			serverAddress := api.MembershipGroupAddress
			udpServer, err := net.ResolveUDPAddr(api.MembershipNetwork, serverAddress)
			connection, err := net.ListenUDP(api.MembershipNetwork, nil)
			if err != nil {
				fmt.Printf("Received an error: %s\n", err)
				return
			}
			defer connection.Close()

			_, err = connection.WriteToUDP(message, udpServer)
			if err != nil {
				fmt.Printf("Received an error: %s", err)
				return
			}
		case <-ctx.Done(): // Activated when ctx.Done() closes
			fmt.Println("Closing MulticastExistence")
			return
		}
	}
}
