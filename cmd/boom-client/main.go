package main

import (
	"fmt"
	"github.com/joostvdg/boom/api"
	"net"
)

const (
	RED    = "\x1b[31;1m"
	GREEN  = "\x1b[32;1m"
	YELLOW = "\x1b[33;1m"
)

func main() {
	fmt.Printf("Hello World, this is Boom!\n")

	listNetworkInterfaces()
	sendSillyMessage()
	sendAnotherSillyMessage()
}

func listNetworkInterfaces() {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("We found an error")
		return
	}

	fmt.Printf("Found network interfaces:\n:")
	for _, networkInterface := range interfaces {
		if networkInterface.Name != "eth0" {
			continue
		}
		fmt.Printf("%sName:%s, Index=%d, MAC:%d%s\n", YELLOW, networkInterface.Name, networkInterface.Index, networkInterface.HardwareAddr, YELLOW)

		addrs, err := networkInterface.Addrs()
		if err != nil {
			fmt.Println("We found an error")
		}
		for _, addr := range addrs {
			fmt.Printf("Addr: %s\n", addr)
		}

		multiCastAddrs, err := networkInterface.MulticastAddrs()
		if err != nil {
			fmt.Println("We found an error")
		}
		for _, multiCastAddr := range multiCastAddrs {
			fmt.Printf("multiCastAddr: %s\n", multiCastAddr)
		}
	}
}

func sendAnotherSillyMessage() {
	ip, _ := api.NewIP4Address("127.0.0.1")
	member := &api.Member{
		MemberName: "MySelf",
		Hostname:   "localhost",
		IP:         &ip,
	}

	message := api.HelloMessage.CreateMemberMessage(member)
	serverAddress := "127.0.0.1:" + api.HelloPort
	udpServer, err := net.ResolveUDPAddr("udp4", serverAddress)
	connection, err := net.ListenUDP("udp4", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer connection.Close()
	localAddr := connection.LocalAddr().(*net.UDPAddr)
	fmt.Printf("Local Address: %s", localAddr)
	_, err = connection.WriteToUDP(message, udpServer)
	if err != nil {
		fmt.Printf("Received an error: %s", err)
		return
	}
}

func sendSillyMessage() {

	serverAddress := "127.0.0.1:" + api.HelloPort
	udpServer, err := net.ResolveUDPAddr("udp4", serverAddress)
	connection, err := net.ListenUDP("udp4", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer connection.Close()
	localAddr := connection.LocalAddr().(*net.UDPAddr)
	fmt.Printf("Local Address: %s", localAddr)
	message := api.ConstructHelloMessage("myself", connection.LocalAddr().String(), api.HelloPort)

	_, err = connection.WriteToUDP(message, udpServer)
	if err != nil {
		fmt.Printf("Received an error: %s", err)
		return
	}
}
