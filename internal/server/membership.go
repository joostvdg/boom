package server

import (
	"context"
	"fmt"
	"github.com/joostvdg/boom/api"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"net"
	"strconv"
	"sync"
	"time"
)

var members map[string]*api.Member
var memberHello = make(chan *api.Member)
var memberGoodbye = make(chan *api.Member)
var memberHelloMulticast = make(chan *api.Member)
var membersLock = make(chan struct{}, 1)

type MembershipService func(*MembershipServiceContext)

type MembershipServiceContext struct {
	context.Context
	TracerProvider *tracesdk.TracerProvider
	SelfAddress    net.Addr
	Self           api.Member
	Identity       string
	HelloMessage   []byte
	GoodbyeMessage []byte
	ServerPort     string
}

// https://github.com/golang/go/issues/20280#issuecomment-655588450
func SetReadDeadlineOnCancel(ctx context.Context, connection *net.UDPConn) {
	go func() {
		<-ctx.Done()
		connection.SetReadDeadline(time.Now())
	}()
}

func NotifyMembersOfLeaving(goodbyeMessage []byte) {
	fmt.Printf("Notifying Members Of Leaving...\n")
	var wg sync.WaitGroup
	for _, member := range members {
		wg.Add(1)
		go func(memberToMessage *api.Member) {
			defer wg.Done()
			serverAddress := memberToMessage.IP.String() + ":" + memberToMessage.PortSelf
			fmt.Printf("Sending goodbye message to %v @%v\n", memberToMessage.MemberName, serverAddress)
			remotePort, err := strconv.Atoi(memberToMessage.PortSelf)
			if err != nil {
				fmt.Printf("Encountered an error when parsing remote port: %s\n", err)
				return
			}
			udpServer := net.UDPAddr{IP: net.ParseIP(memberToMessage.IP.String()), Port: remotePort}

			connection, err := net.ListenUDP(api.MembershipNetwork, nil)
			if err != nil {
				fmt.Printf("Encountered an error when creating the local connection: %s\n", err)
				return
			}

			defer connection.Close()
			_, err = connection.WriteToUDP(goodbyeMessage, &udpServer)
			if err != nil {
				fmt.Printf("Encountered an error when sending the goodbye message: %s\n", err)
				return
			}
		}(member)
	}
	wg.Wait()
}

func CloseChannels() {
	close(memberHelloMulticast)
	close(memberHello)
}
