package server

import (
	"fmt"
	"github.com/joostvdg/boom/api"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

const name = "boom-server"

// StartMembershipServer starts the server that listens to all kinds of Membership messages
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
	defer fmt.Println("Closing StartMembershipServer")
	SetReadDeadlineOnCancel(ctx, connection)

	buffer := make([]byte, 1024)

	fmt.Printf("Listening on port %s for Hello & Goodbye messages...\n", port)
	for {
		var span trace.Span
		if serviceContext.TracingEnabled {
			_, span = serviceContext.TracerProvider.Tracer(name).Start(ctx, "Membership")
		}
		numberOfBytes, address, err := connection.ReadFromUDP(buffer)
		// TODO find a way to abstract away these steps
		if serviceContext.TracingEnabled {
			span.SetAttributes(attribute.Int("bytes", numberOfBytes))
			span.SetAttributes(attribute.String("origin", address.String()))
		}
		if err != nil {
			fmt.Printf("Encountered an error reading from UDP connection: %s\n", err)
			return
		}
		fmt.Printf("Received message %s from %s\n", string(buffer[0:numberOfBytes]), address.String())
		member, messageType, err := api.ReadMemberMessage(buffer[0:numberOfBytes], address)
		helloMessageType := "unknown"
		if err != nil {
			fmt.Printf("Encountered an error determining message type: %s\n", err)
		} else {
			switch messageType.Prefix {
			case api.HelloPrefix:
				helloMessageType = "hello"
				memberHello <- member
			case api.GoodbyePrefix:
				helloMessageType = "goodbye"
				memberGoodbye <- member
			case api.HeartbeatRequestPrefix:
				helloMessageType = "HeartbeatRequest"
				memberHeartbeatRequest <- member
			case api.HeartbeatResponsePrefix:
				helloMessageType = "HeartbeatResponse"
				memberHeartbeatResponse <- member
			case api.MemberFailureDetectedPrefix:
				helloMessageType = "MemberFailureDetected"
				memberNotResponding <- member
			default:
				fmt.Println("Ran into an error, unknown message type")
			}
		}
		if serviceContext.TracingEnabled {
			span.SetAttributes(attribute.String("type", helloMessageType))
			span.End()
		}
	}
}

// ListenForMulticast listens for BOOM servers annoucning themselves via UDP multicast
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
	defer fmt.Println("Closing ListenForMulticast")
	SetReadDeadlineOnCancel(ctx, connection)

	buffer := make([]byte, 1024)
	for {
		var span trace.Span
		if serviceContext.TracingEnabled {
			_, span = serviceContext.TracerProvider.Tracer(name).Start(ctx, "Membership-Broadcast")
		}
		numberOfBytes, originAddress, err := connection.ReadFromUDP(buffer)
		if serviceContext.TracingEnabled {
			span.SetAttributes(attribute.Int("bytes", numberOfBytes))
			span.SetAttributes(attribute.String("origin", originAddress.String()))
		}
		if err != nil {
			fmt.Printf("Received an error: %s\n", err)
			if serviceContext.TracingEnabled {
				span.End()
			}
			return
		}

		// TODO: should we treat this type of message differently?
		member, _, err := api.ReadMemberMessage(buffer[0:numberOfBytes], originAddress)
		if err != nil {
			fmt.Printf("Ran into an error: %s\n", err)
		} else {
			memberHelloMulticast <- member
		}
		if serviceContext.TracingEnabled {
			span.End()
		}
	}
}

func MulticastExistence(serviceContext *MembershipServiceContext) {
	ctx := serviceContext.Context
	message := serviceContext.HelloMessage
	clock := time.NewTicker(30 * time.Second)
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
			_, err = connection.WriteToUDP(message, udpServer)
			if err != nil {
				fmt.Printf("Received an error: %s", err)
				return
			}
			connection.Close() // not using defer as we're in a loop
		case <-ctx.Done(): // Activated when ctx.Done() closes
			fmt.Println("Closing MulticastExistence")
			return
		}
	}
}

func HeartbeatCloseMembers(serviceContext *MembershipServiceContext) {
	ctx := serviceContext.Context
	message := serviceContext.HeartbeatRequest
	clock := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-clock.C:
			clockUpdate <- 1
			// TODO verify if this is a good idea, at least at some point we will have populated this map
			// TODO: maybe we should be able to provide a "starter list" as a possible override in the init

			// As long as we do not have our max in the short list, we should add more
			memberShortListLock <- struct{}{}
			if len(memberShortList) < MaxShortListSize && len(members) > 0 {
				sizeCounter := 0
				for _, member := range members {
					if sizeCounter >= MaxShortListSize {
						return
					}
					memberShortList[member.Identifier()] = member
					sizeCounter++
				}
			}
			<-memberShortListLock
			for _, member := range memberShortList {
				go sendHeartbeatRequest(member, message)
			}
		case <-ctx.Done(): // Activated when ctx.Done() closes
			fmt.Println("Closing HeartbeatCloseMembers")
			return
		}
	}
}

func sendHeartbeatRequest(memberToMessage *api.Member, message []byte) {
	serverAddress := memberToMessage.IP.String() + ":" + memberToMessage.PortSelf
	fmt.Printf("Sending heartbeat request message to %v @%v\n",
		memberToMessage.MemberName, serverAddress)
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
	_, err = connection.WriteToUDP(message, &udpServer)
	if err != nil {
		fmt.Printf("Encountered an error when sending the heartbeat request message: %s\n", err)
		return
	}
	HandleHeartbeatResponseTracking(memberToMessage)
}

func HandleClockUpdates(serviceContext *MembershipServiceContext) {
	ctx := serviceContext.Context
	for {
		select {
		case update := <-clockUpdate:
			clockLock <- struct{}{} // acquire token
			serviceContext.Self.Clock += update
			<-clockLock // release token
		case <-ctx.Done(): // Activated when ctx.Done() closes
			fmt.Println("Closing HandleClockUpdates")
			return
		}
	}
}

func HandleHeartbeatResponseTrackingUpdate(memberResponded *api.Member) {
	if heartbeatResponses[memberResponded.Identifier()] == nil {
		fmt.Printf("Received a response from a member we are no longer tracking: %v\n", memberResponded)
		return
	}
	heartbeatResponsesLock <- struct{}{} // acquire token
	memberTracker := heartbeatResponses[memberResponded.Identifier()]
	memberTracker.LastResponseClock = memberResponded.Clock
	memberTracker.LastResponse = time.Now()
	memberTracker.MissedResponsesCounter = 0
	<-heartbeatResponsesLock
}

func HandleHeartbeatResponseTracking(memberToTrack *api.Member) {
	var memberTracker *heartbeatResponseTracker
	if heartbeatResponses[memberToTrack.Identifier()] == nil {
		fmt.Printf("Requesting a response from a new Member: %+v\n", memberToTrack)
		memberTracker = &heartbeatResponseTracker{
			MissedResponsesCounter: 1,
			LastResponse:           NoResponseTime,
			LastResponseClock:      0,
		}
		heartbeatResponsesLock <- struct{}{} // acquire token
		heartbeatResponses[memberToTrack.Identifier()] = memberTracker
		<-heartbeatResponsesLock
	} else {
		memberTracker = heartbeatResponses[memberToTrack.Identifier()]
		if memberTracker.MissedResponsesCounter >= 5 {
			fmt.Printf("We are not able to reach %v for 5 times, initiating failure propagation\n", memberToTrack)
			// TODO: review this
			for _, member := range memberShortList {
				message := api.ConstructMemberFailureDetectedMessage(memberToTrack)
				if member.Identifier() != memberToTrack.Identifier() {
					err := sendMessageToMember(member, message, "failureDetected")
					if err != nil {
						fmt.Printf("Could not send MemberFailureDetected to %v: %v\n", member, err)
					}
				}
			}
		} else {
			heartbeatResponsesLock <- struct{}{} // acquire token
			memberTracker.MissedResponsesCounter++
			<-heartbeatResponsesLock
		}
	}
}


func HandleMemberNotResponding(member *api.Member, message []byte) {
	// TODO: remove from MembersList and MemberShortList
	// TODO: add to - or update - member in MemberFailList
	// TODO: request a heartbeat response
	membersLock <- struct{}{} //acquire token
	delete(members, member.Identifier())
	<-membersLock //release token

	memberShortListLock <- struct{}{}
	delete(memberShortList, member.Identifier())
	<-memberShortListLock

	memberFailListLock <- struct{}{}
	memberFailList[member.Identifier()] = member
	<- memberFailListLock

	go sendHeartbeatRequest(member, message)
}