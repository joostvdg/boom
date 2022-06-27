package api

const MEMBERSHIP_NETWORK = "udp"
const MEMBERSHIP_GROUP_ADDRESS = "230.0.0.0:7791"

const HelloPrefix byte = 0x01
const HelloPrefixSize = 1
const HelloMessageSizeSize = 4
const HelloMemberNameSize = 8
const HelloHostnameSize = 8
const HelloIPSize = 4
const HelloMessageSize = HelloPrefixSize + HelloMessageSizeSize + HelloMemberNameSize + HelloHostnameSize + HelloIPSize

const HelloPort = ":7777"

type Member struct {
	MemberName string
	Hostname   string
	IP         string
}

func (m Member) CreateMemberMessage() []byte {
	cursor := 0
	message := make([]byte, HelloMessageSize)
	message[cursor] = HelloPrefix
	cursor += HelloPrefixSize

	message = addHeaderToMessage(message, cursor, cursor+HelloMemberNameSize, []byte(m.MemberName))
	cursor += HelloMemberNameSize

	message = addHeaderToMessage(message, cursor, cursor+HelloHostnameSize, []byte(m.Hostname))
	cursor += HelloHostnameSize

	message = addHeaderToMessage(message, cursor, cursor+HelloIPSize, []byte(m.IP))
	cursor += HelloIPSize

	return message
}

func addHeaderToMessage(message []byte, start int, end int, data []byte) []byte {
	cursor := start
	for _, dataByte := range data {
		if cursor <= end {
			message[cursor] = dataByte
			cursor++
		}
	}
	return message
}
