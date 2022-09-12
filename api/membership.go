package api

import (
	"encoding/binary"
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

const HeartbeatRequestPrefix byte = 0x10
const HeartbeatRequestPrefixSize = 1
const HeartbeatResponsePrefix byte = 0x11
const HeartbeatResponsePrefixSize = 1

const MemberFailureDetectedPrefix byte = 0x20
const MemberFailureDetectedPrefixSize = 1

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
	Clock      int64
}

var MemberNameField MessageField
var HostnameField MessageField
var IPField MessageField
var PortField MessageField
var ClockField MessageField

var HelloMessage MessageType
var GoodbyeMessage MessageType
var HeartbeatRequestMessage MessageType
var HeartbeatResponseMessage MessageType
var MemberFailureDetected MessageType

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
	ClockField = MessageField{
		Name:            "Clock",
		Size:            8,
		MemberField:     "Clock",
		MemberFieldType: "int",
	}

	HelloMessage = MessageType{
		Prefix:        HelloPrefix,
		PrefixSize:    HelloPrefixSize,
		MessageFields: []MessageField{MemberNameField, HostnameField, IPField, PortField, ClockField},
	}
	GoodbyeMessage = MessageType{
		Prefix:        GoodbyePrefix,
		PrefixSize:    GoodbyePrefixSize,
		MessageFields: []MessageField{MemberNameField, HostnameField, IPField, PortField, ClockField},
	}
	HeartbeatRequestMessage = MessageType{
		Prefix:        HeartbeatRequestPrefix,
		PrefixSize:    HeartbeatRequestPrefixSize,
		MessageFields: []MessageField{MemberNameField, HostnameField, IPField, PortField, ClockField},
	}
	HeartbeatResponseMessage = MessageType{
		Prefix:        HeartbeatResponsePrefix,
		PrefixSize:    HeartbeatResponsePrefixSize,
		MessageFields: []MessageField{MemberNameField, HostnameField, IPField, PortField, ClockField},
	}
	MemberFailureDetected = MessageType{
		Prefix:        MemberFailureDetectedPrefix,
		PrefixSize:    MemberFailureDetectedPrefixSize,
		MessageFields: []MessageField{MemberNameField, HostnameField, IPField, PortField, ClockField},
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
		fieldValue := make([]byte, field.Size)
		switch field.MemberFieldType {
		case "int":
			intValue := reflect.ValueOf(*m).FieldByName(field.MemberField).Int()
			binary.LittleEndian.PutUint64(fieldValue, uint64(intValue))
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
		case HeartbeatResponsePrefix:
			messageType = HeartbeatResponseMessage
		case HeartbeatRequestPrefix:
			messageType = HeartbeatRequestMessage
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
	clock, bytesProcessed := getHeader(rawMessage, bytesProcessed, ClockField.Size)

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
		Clock:      int64(binary.LittleEndian.Uint64(clock)),
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

func constructMemberForMessage(name string, localAddress string, port string) (*Member, error) {
	ip, _ := NewIP4Address(localAddress)
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &Member{
		MemberName: name,
		Hostname:   hostname,
		IPSelf:     &ip,
		PortSelf:   port,
		Clock: 		0,
	}, nil
}

func ConstructMemberFailureDetectedMessage(member *Member) []byte {
	return MemberFailureDetected.CreateMemberMessage(member)
}

func ConstructHeartbeatRequestMessage(name string, localAddress string, port string) []byte {
	member, err := constructMemberForMessage(name, localAddress, port)
	if err != nil {
		fmt.Printf("Ran into an error trying to create a Member for heartbeat request message: %v\n", err)
		panic("Cannot instantiate HeartbeatRequestMessage")
	}
	return HeartbeatRequestMessage.CreateMemberMessage(member)
}

func ConstructHeartbeatResponseMessage(name string, localAddress string, port string) []byte {
	member, err := constructMemberForMessage(name, localAddress, port)
	if err != nil {
		fmt.Printf("Ran into an error trying to create a Member for heartbeat response message: %v\n", err)
		panic("Cannot instantiate HeartbeatResponseMessage")
	}
	return HeartbeatResponseMessage.CreateMemberMessage(member)
}

func ConstructGoodbyeMessage(name string, localAddress string, port string) []byte {
	member, err := constructMemberForMessage(name, localAddress, port)
	if err != nil {
		fmt.Printf("Ran into an error trying to create a Member for goodbye message: %v\n", err)
		panic("Cannot instantiate GoodbyeMessage")
	}
	return GoodbyeMessage.CreateMemberMessage(member)
}

func ConstructHelloMessage(name string, localAddress string, port string) []byte {
	member, err := constructMemberForMessage(name, localAddress, port)
	if err != nil {
		fmt.Printf("Ran into an error trying to create a Member for hello message: %v\n", err)
		panic("Cannot instantiate HelloMessage")
	}
	return HelloMessage.CreateMemberMessage(member)
}
