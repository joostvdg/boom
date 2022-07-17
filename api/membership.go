package api

import (
	"errors"
	"fmt"
)

const MEMBERSHIP_NETWORK = "udp"
const MEMBERSHIP_GROUP_ADDRESS = "230.0.0.0:7791"

const HelloPrefix byte = 0x01
const HelloPrefixSize = 1
const HelloMemberNameSize = 12
const HelloHostnameSize = 12
const HelloIPSize = 4
const HelloMessageHeaderSize = HelloPrefixSize + HelloMemberNameSize + HelloHostnameSize + HelloIPSize
const HelloPort = ":7777"



type Member struct {
	MemberName string
	Hostname   string
	IP         *IP4Address
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
		fmt.Println("Received a Member Hello message!")
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
	member := &Member{
		MemberName: string(memberName),
		Hostname:   string(hostname),
		IP:         ip4Address,
	}
	return member, nil
}

func getHeader(rawMessage []byte, start int, length int) ([]byte, int) {
	end := start + length
	return rawMessage[start:end], end
}
