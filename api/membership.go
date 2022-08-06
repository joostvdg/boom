package api

import (
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"time"
)

const MembershipNetwork = "udp4"
const MembershipGroupAddress = "230.0.0.0:7791"

const GoodbyePrefix byte = 0x02
const GoodbyePrefixSize = 1
const HelloPrefix byte = 0x01
const HelloPrefixSize = 1
const HelloPort = "7777"

type MessageField struct {
	Name            string
	Size            int
	MemberField     string
	MemberFieldType string
}

type MessageType struct {
	Prefix        byte
	PrefixSize    int
	MessageFields []MessageField
}

type Member struct {
	MemberName string
	Hostname   string
	IP         *IP4Address
	PortSelf   string
	IPSelf     *IP4Address
	LastSeen   time.Time
}

var MemberNameField MessageField
var HostnameField MessageField
var IPField MessageField
var PortField MessageField

var HelloMessage MessageType
var GoodbyeMessage MessageType

func init() {
	MemberNameField = MessageField{
		Name:            "MemberName",
		Size:            12,
		MemberField:     "MemberName",
		MemberFieldType: "string",
	}

	HostnameField = MessageField{
		Name:            "Hostname",
		Size:            12,
		MemberField:     "Hostname",
		MemberFieldType: "string",
	}
	IPField = MessageField{
		Name:            "IP",
		Size:            4,
		MemberField:     "IP",
		MemberFieldType: "ip4address",
	}
	PortField = MessageField{
		Name:            "Port",
		Size:            6,
		MemberField:     "PortSelf",
		MemberFieldType: "string",
	}

	HelloMessage = MessageType{
		Prefix:        HelloPrefix,
		PrefixSize:    HelloPrefixSize,
		MessageFields: []MessageField{MemberNameField, HostnameField, IPField, PortField},
	}
	GoodbyeMessage = MessageType{
		Prefix:        GoodbyePrefix,
		PrefixSize:    GoodbyePrefixSize,
		MessageFields: []MessageField{MemberNameField, HostnameField, IPField, PortField},
	}
}

func (mt MessageType) HeaderSize() int {
	headerSize := mt.PrefixSize
	for _, field := range mt.MessageFields {
		headerSize += field.Size
	}
	return headerSize
}

func (mt MessageType) CreateMemberMessage(m *Member) []byte {
	message := make([]byte, mt.HeaderSize(), mt.HeaderSize())
	cursor := 0
	message[cursor] = mt.Prefix
	cursor += mt.PrefixSize
	for _, field := range mt.MessageFields {
		var fieldValue []byte
		switch field.MemberFieldType {
		case "string":
			fieldValue = []byte(reflect.ValueOf(*m).FieldByName(field.MemberField).String())
		case "ip4address":
			fieldValue = m.IPSelf.ToByteArray() // special case
		default:
			panic("Do not understand MessageFieldType")
		}

		message = appendHeaderToMessage(message, cursor, cursor+field.Size, fieldValue)
		cursor += field.Size
	}
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

func ReadMemberMessage(rawMessage []byte, messageOriginAddress *net.UDPAddr) (*Member, MessageType, error) {
	bytesProcessed := 0
	var messageType MessageType
	messagePrefix, bytesProcessed := getHeader(rawMessage, bytesProcessed, HelloPrefixSize)

	if len(messagePrefix) == HelloPrefixSize {
		prefix := messagePrefix[0]
		switch prefix {
		case HelloPrefix:
			messageType = HelloMessage
		case GoodbyePrefix:
			messageType = GoodbyeMessage
		default:
			return nil, messageType, errors.New("unknown message type")
		}
	} else {
		return nil, messageType, errors.New("unreadable message")
	}

	// TODO replace with reading from messageType struct

	memberName, bytesProcessed := getHeader(rawMessage, bytesProcessed, MemberNameField.Size)
	hostname, bytesProcessed := getHeader(rawMessage, bytesProcessed, HostnameField.Size)
	selfKnownIP, bytesProcessed := getHeader(rawMessage, bytesProcessed, IPField.Size)
	port, bytesProcessed := getHeader(rawMessage, bytesProcessed, PortField.Size)

	selfKnownAddress := &IP4Address{
		A: selfKnownIP[0],
		B: selfKnownIP[1],
		C: selfKnownIP[2],
		D: selfKnownIP[3],
	}

	memberName = removeEmptyBytes(memberName)
	hostname = removeEmptyBytes(hostname)
	port = removeEmptyBytes(port)

	originAddress, err := NewIP4Address(messageOriginAddress.String())
	if err != nil {
		fmt.Printf("Ran into an error: %s\n", err)
	}

	member := &Member{
		MemberName: string(memberName),
		Hostname:   string(hostname),
		IPSelf:     selfKnownAddress,
		IP:         &originAddress,
		PortSelf:   string(port),
	}
	return member, messageType, nil
}

func removeEmptyBytes(bytesRead []byte) []byte {
	bytesToReturn := make([]byte, 0)
	for _, byteRead := range bytesRead {
		if byteRead != byte(0) {
			bytesToReturn = append(bytesToReturn, byteRead)
		}
	}
	return bytesToReturn
}

func getHeader(rawMessage []byte, start int, length int) ([]byte, int) {
	end := start + length
	return rawMessage[start:end], end
}

func ConstructGoodbyeMessage(name string, localAddress string, port string) []byte {
	ip, _ := NewIP4Address(localAddress)
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	member := &Member{
		MemberName: name,
		Hostname:   hostname,
		IPSelf:     &ip,
		PortSelf:   port,
	}
	return GoodbyeMessage.CreateMemberMessage(member)
}

func ConstructHelloMessage(name string, localAddress string, port string) []byte {
	ip, _ := NewIP4Address(localAddress)
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	member := &Member{
		MemberName: name,
		Hostname:   hostname,
		IPSelf:     &ip,
		PortSelf:   port,
	}
	return HelloMessage.CreateMemberMessage(member)
}
