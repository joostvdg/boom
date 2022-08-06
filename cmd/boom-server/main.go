package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joostvdg/boom/api"
	"github.com/joostvdg/boom/internal/server"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	helloPortOverride := flag.String("helloPort", api.HelloPort, fmt.Sprintf("PortSelf number for listening for Hello messages, default %s", api.HelloPort))
	helloName := flag.String("helloName", "MySelf", "Name of this Boom server")
	flag.Parse()

	myAddress := determineAddress()
	helloMessage := api.ConstructHelloMessage(*helloName, myAddress.String(), *helloPortOverride)
	goodbyeMessage := api.ConstructGoodbyeMessage(*helloName, myAddress.String(), *helloPortOverride)
	myself := createMyself(*helloName, myAddress)
	myIdentity := myself.Identifier()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	membershipServiceContext := &server.MembershipServiceContext{
		Context:        ctx,
		SelfAddress:    myAddress,
		Self:           myself,
		Identity:       myIdentity,
		HelloMessage:   helloMessage,
		GoodbyeMessage: goodbyeMessage,
		ServerPort:     *helloPortOverride,
	}
	membershipServices := []server.MembershipService{
		server.ListenForMulticast,
		server.MulticastExistence,
		server.StartMembershipServer,
		server.HandleMember,
		server.CleanupMembers,
	}

	var wg sync.WaitGroup
	for _, membershipService := range membershipServices {
		wg.Add(1)
		go func(service server.MembershipService) {
			service(membershipServiceContext)
			defer wg.Done()
		}(membershipService)
	}
	wg.Wait()

	fmt.Printf("Shutting down!\n")
	server.CloseChannels()
	server.NotifyMembersOfLeaving(goodbyeMessage)
}

func createMyself(name string, address net.Addr) api.Member {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	ip, _ := api.NewIP4Address(address.String())

	member := api.Member{
		MemberName: name,
		Hostname:   hostname,
		IPSelf:     &ip,
	}
	return member
}

func determineAddress() net.Addr {
	connection, err := net.ListenUDP("udp4", nil)
	if err != nil {
		fmt.Println(err)
	}

	defer connection.Close()
	address := connection.LocalAddr()
	return address
}
