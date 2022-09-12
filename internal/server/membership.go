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
var membersLock = make(chan struct{}, 1)
var memberShortList map[string]*api.Member
var memberShortListLock = make(chan struct{}, 1)
var memberFailList map[string]*api.Member
var memberFailListLock = make(chan struct{}, 1)
var heartbeatResponses map[string]*heartbeatResponseTracker
var heartbeatResponsesLock = make(chan struct{}, 1)
var clockUpdate = make(chan int64)
var clockLock = make(chan struct{}, 1)
var memberHeartbeatRequest = make(chan *api.Member)
var memberHeartbeatResponse = make(chan *api.Member)
var memberNotResponding = make(chan *api.Member)
var memberHello = make(chan *api.Member)
var memberGoodbye = make(chan *api.Member)
var memberHelloMulticast = make(chan *api.Member)

var NoResponseTime time.Time //time.Date(1970, 1, 1, 0, 0,0, 0, nil)
const MaxShortListSize = 3

type MembershipService func(*MembershipServiceContext)

func init() {
	members = make(map[string]*api.Member)
	memberShortList = make(map[string]*api.Member)
	memberFailList = make(map[string]*api.Member)
	heartbeatResponses = make (map[string]*heartbeatResponseTracker)
	NoResponseTime = time.Unix(0, 0)
}

type MembershipServiceContext struct {
	context.Context
	TracingEnabled    bool
	TracerProvider    *tracesdk.TracerProvider
	SelfAddress       net.Addr
	Self              *api.Member
	Identity          string
	HelloMessage      []byte
	GoodbyeMessage    []byte
	HeartbeatRequest  []byte
	HeartbeatResponse []byte
	ServerPort        string
}

// TODO test this and refine
// heartbeatResponseTracker is a way to track if the members we send a heartbeat too, are responding
type heartbeatResponseTracker struct {
	MissedResponsesCounter int
	LastResponse           time.Time
	LastResponseClock      int64
}

// SetReadDeadlineOnCancel sets the deadline for connections to "now", once the context is finished
// else the sockets stay open, and do not let us close gracefully, see link below:
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
			err := sendMessageToMember(memberToMessage, goodbyeMessage, "leave")
			if err != nil {
				fmt.Printf("Could not send leave message to %v: %v\n", memberToMessage, err)
			}
		}(member)
	}
	wg.Wait()
}

func sendMessageToMember(memberToMessage *api.Member, message []byte, messageType string) error {
	serverAddress := memberToMessage.IP.String() + ":" + memberToMessage.PortSelf
	fmt.Printf("Sending %v message to %v @%v\n", messageType, memberToMessage.MemberName, serverAddress)
	remotePort, err := strconv.Atoi(memberToMessage.PortSelf)
	if err != nil {
		fmt.Printf("Encountered an error when parsing remote port: %s\n", err)
		return nil
	}
	udpServer := net.UDPAddr{IP: net.ParseIP(memberToMessage.IP.String()), Port: remotePort}

	connection, err := net.ListenUDP(api.MembershipNetwork, nil)
	if err != nil {
		fmt.Printf("Encountered an error when creating the local connection: %s\n", err)
		return nil
	}

	defer connection.Close()
	_, err = connection.WriteToUDP(message, &udpServer)
	if err != nil {
		fmt.Printf("Encountered an error when sending the %v message: %s\n", messageType, err)
		return err
	}
	return nil
}

func CloseChannels() {
	close(memberHelloMulticast)
	close(memberHello)
}
