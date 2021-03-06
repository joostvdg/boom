package api

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const MEMBERSHIP_NETWORK = "udp"
const MEMBERSHIP_GROUP_ADDRESS = "230.0.0.0:7791"

const HelloPrefix byte = 0x01
const HelloPrefixSize = 1
const HelloMemberNameSize = 12
const HelloHostnameSize = 12
const HelloIPSize = 4
const HelloMessageHeaderSize = HelloPrefixSize + HelloMemberNameSize + HelloHostnameSize + HelloIPSize
const HelloPort = "7777"



type Member struct {
	MemberName string
	Hostname   string
	IP         *IP4Address
	LastSeen   time.Time
}

func (m Member) CreateMemberMessage() []byte {
	message := make([]byte, HelloMessageHeaderSize, HelloMessageHeaderSize)
	cursor := 0
	message[cursor] = HelloPrefix
	cursor += HelloPrefixSize

	message = appendHeaderToMessage(message, cursor, cursor+HelloMemberNameSize, []byte(m.MemberName))
	cursor += HelloMemberNameSize

	message = appendHeaderToMessage(message, cursor, cursor+HelloHostnameSize, []byte(m.Hostname))
	cursor += HelloHostnameSize

	message = appendHeaderToMessage(message, cursor, cursor+HelloIPSize, m.IP.ToByteArray())

	return message
}

func (m *Member) Identifier() string {
	return m.MemberName + "@" + m.Hostname
}

func appendHeaderToMessage(message []byte, start int, end int, data []byte) []byte {
	cursor := start
	for _, dataByte := range data {
		if cursor < end {
			message[cursor] = dataByte
			cursor++
		}
	}
	return message
}

func ReadMemberMessage(rawMessage []byte) (*Member, error) {
	bytesProcessed := 0
	messagePrefix, bytesProcessed := getHeader(rawMessage, bytesProcessed, HelloPrefixSize)

	if len(messagePrefix) == HelloPrefixSize && messagePrefix[0] == HelloPrefix {
		// fmt.Println("Received a Member Hello message!")
		// message type is Hello
	} else {
		return nil, errors.New("unknown message")
	}

	memberName, bytesProcessed := getHeader(rawMessage, bytesProcessed, HelloMemberNameSize)
	hostname, bytesProcessed := getHeader(rawMessage, bytesProcessed, HelloHostnameSize)
	ip, bytesProcessed := getHeader(rawMessage, bytesProcessed, HelloIPSize)
	ip4Address := &IP4Address{
		A: ip[0],
		B: ip[1],
		C: ip[2],
		D: ip[3],
	}

	memberName = removeEmptyBytes(memberName)
	hostname = removeEmptyBytes(hostname)

	member := &Member{
		MemberName: string(memberName),
		Hostname:   string(hostname),
		IP:         ip4Address,
	}
	return member, nil
}

func removeEmptyBytes(bytesRead []byte) []byte {
	bytesToReturn := make([]byte, 0)
	for _, byteRead := range bytesRead {
		if byteRead != byte(0) {
			bytesToReturn = append(bytesToReturn,byteRead)
		}
	}
	return bytesToReturn
}

func getHeader(rawMessage []byte, start int, length int) ([]byte, int) {
	end := start + length
	return rawMessage[start:end], end
}

func ConstructMembershipMessage(name string, localAddress string) []byte {
	ip, _ := NewIP4Address(localAddress)
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	member := &Member{
		MemberName: name,
		Hostname:   hostname,
		IP:         &ip,
	}
	return member.CreateMemberMessage()
}