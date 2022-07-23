package main

import (
	"flag"
	"fmt"
	"github.com/joostvdg/boom/api"
	"github.com/joostvdg/boom/internal/server"
	"net"
	"os"
)


func main() {
	helloPortOverride := flag.String("helloPort", api.HelloPort, fmt.Sprintf("Port number for listening for Hello messages, default %s", api.HelloPort))
	helloName := flag.String("helloName", "MySelf", "Name of this Boom server")
	flag.Parse()

	myAddress := determineAddress()
	myMessage := api.ConstructMembershipMessage(*helloName, myAddress.String())
	myself := createMyself(*helloName, myAddress)
	myIdentity := myself.Identifier()

	go server.HandleMember(myIdentity)
	go server.ListenForMulticast()
	go server.MulticastExistence(myMessage)
	server.StartHelloServer(*helloPortOverride)
}

func createMyself(name string, address net.Addr) api.Member {
	hostname, error := os.Hostname()
	if error != nil {
		panic(error)
	}
	ip, _ := api.NewIP4Address(address.String())

	member := api.Member{
		MemberName: name,
		Hostname:   hostname,
		IP:         &ip,
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



