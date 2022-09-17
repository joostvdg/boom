package server

import (
	"fmt"
	"time"
)

func HandleMember(serviceContext *MembershipServiceContext) {
	ctx := serviceContext.Context
	myIdentity := serviceContext.Identity
	for {
		select {
		case <-ctx.Done(): // Activated when ctx.Done() closes
			fmt.Println("Closing HandleMember")
			return
		case member := <-memberHello:
			// ignore myself
			if member.Identifier() == myIdentity {
				return
			}

			member.LastSeen = time.Now()
			if members[member.Identifier()] == nil {
				fmt.Printf("Received Hello from new Member: %+v\n", member)
			} else {
				lastSeenInfo := members[member.Identifier()]
				durationSinceLastSeen := member.LastSeen.Sub(lastSeenInfo.LastSeen)
				fmt.Printf("Received Hello from known Member: %s, first message since: %v\n", member.Identifier(), durationSinceLastSeen)
			}
			membersLock <- struct{}{} //acquire token
			members[member.Identifier()] = member
			<-membersLock //release token
			fmt.Printf("Updated member %s's last seen", member.MemberName)
		case member := <-memberGoodbye:
			// ignore myself or any member we didn't know anyway
			if member.Identifier() == myIdentity || members[member.Identifier()] == nil {
				return
			}
			fmt.Printf("Received Goodbye from known Member: %s (%v), removing from Membership\n", member.Identifier(), member.IP.String())
			membersLock <- struct{}{} //acquire token
			delete(members, member.Identifier())
			<-membersLock //release token
			if memberShortList[member.Identifier()] != nil {
				memberShortListLock <- struct{}{} //acquire token
				delete(memberShortList, member.Identifier())
				<-memberShortListLock //release token
			}
			fmt.Printf("Member %s removed\ngo", member.MemberName)
		case member := <-memberHeartbeatRequest:
			// ignore myself
			if member.Identifier() == myIdentity {
				continue
			}
			// when we get a request, answer it with a response
			fmt.Printf("Received heartbeat request from member %v\n", member)
			clockUpdate <- 1

			// if we have not filled our shortlist yet, we can probably fill it with those that are talking to us
			// TODO: this might be counter productive, and perhaps we should reset this list overtime?
			memberShortListLock <- struct{}{}
			if len(memberShortList) < MaxShortListSize && memberShortList[member.Identifier()] == nil {
				memberShortList[member.Identifier()] = member
			}
			<-memberShortListLock

			err := sendMessageToMember(member, serviceContext.HeartbeatResponse, "heartbeatResponse")
			if err != nil {
				fmt.Printf("Could not send heartbeat response to %v: %v", member, err)
			}
		case member := <-memberHeartbeatResponse:
			// ignore myself
			if member.Identifier() == myIdentity {
				continue
			}
			fmt.Printf("Received heartbeat response from member %v\n", member)
			go HandleHeartbeatResponseTrackingUpdate(member)
		case member := <-memberNotResponding:
			if member.Identifier() == myIdentity {
				continue
			fmt.Printf("We heard member %v is no longer alive, lets scrap him \n", member)
			HandleMemberNotResponding(member, serviceContext.HeartbeatRequest)}
			// TODO: when we've tried to reach a Member for X tries, and we do not get a response, let the others know
			// TODO: we should probably verify this ourselves, before scrapping the poor sod
			// TODO: limit how many times we send a failure propagation
		case member := <-memberHelloMulticast:
			// ignore myself
			if member.Identifier() == myIdentity {
				continue
			}
			member.LastSeen = time.Now()
			membersLock <- struct{}{} //acquire token
			members[member.Identifier()] = member
			<-membersLock //release token
			fmt.Printf("Received Multicast from Member: %s @%v(%v:%v / %v)\n", member.MemberName, member.Hostname, member.IP, member.PortSelf, member.IPSelf)
		}
	}
}


func CleanupMembers(serviceContext *MembershipServiceContext) {
	ctx := serviceContext.Context
	clock := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done(): // Activated when ctx.Done() closes
			fmt.Println("Closing CleanupMembers")
			return
		case <-clock.C:
			for _, member := range members {
				durationSinceLastSeen := time.Now().Sub(member.LastSeen)
				if durationSinceLastSeen > (time.Second * 40) {
					fmt.Printf("Removing member %v because they did not check in recently\n", member)
					membersLock <- struct{}{} //acquire token
					delete(members, member.Identifier())
					<-membersLock //release tokenc
				}
			}

			// TODO: we should also cleanup failing members that have not responded to our heartbeat request
			for _, member := range memberFailList {
				tracker := heartbeatResponses[member.Identifier()]
				if tracker.LastResponse.After(time.Unix(0,0)) {
					// TODO: OMG, it is resurrected from the Dead, what to do?
					fmt.Printf("We heard from our long lost brother: %v\n", member)
				} else {
					memberFailListLock <- struct{}{}
					delete(memberFailList, member.Identifier())
					<-memberShortListLock

					heartbeatResponsesLock <- struct{}{}
					delete(heartbeatResponses, member.Identifier())
					<- heartbeatResponsesLock
				}
			}
		}
	}
}
