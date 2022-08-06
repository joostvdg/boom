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
			fmt.Printf("Member %s removed\ngo", member.MemberName)
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
		}

	}
}
