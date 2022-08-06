package api

import (
	"reflect"
	"testing"
)

type testData struct {
	testName string
	name     string
	host     string
	ip       string
	port     string
	data     []byte
}

func TestMember_CreateMemberMessage(t *testing.T) {
	testOne := testData{
		testName: "TestIfAllFieldsWork",
		name:     "Test",
		host:     "Boreas",
		ip:       "0.0.0.0",
		port:     "37160",
	}
	testOne.data = CreateBasicTestData(testOne.name, testOne.host, testOne.ip, testOne.port)

	testTwo := testData{
		testName: "TestIfAllFieldsWorkAlt",
		name:     "MySelf",
		host:     "localhost",
		ip:       "127.0.0.1",
		port:     "37160",
	}
	testTwo.data = CreateBasicTestData(testTwo.name, testTwo.host, testTwo.ip, testTwo.port)

	testHostnameTooLongGetsTruncated := testData{
		testName: "testHostnameTooLongGetsTruncated",
		name:     "MySelf",
		host:     "localhost",
		ip:       "127.0.0.1",
		port:     "37160",
	}
	truncatedHostname := truncate(testHostnameTooLongGetsTruncated.host, HostnameField.Size)
	testHostnameTooLongGetsTruncated.data = CreateBasicTestData(testHostnameTooLongGetsTruncated.name, truncatedHostname, testHostnameTooLongGetsTruncated.ip, testHostnameTooLongGetsTruncated.port)

	testMaxIP := testData{
		testName: "TestMaxIPShouldWork",
		name:     "MySelf",
		host:     "localhost",
		ip:       "255.255.255.254",
		port:     "37160",
	}
	testMaxIP.data = CreateBasicTestData(testMaxIP.name, testMaxIP.host, testMaxIP.ip, testMaxIP.port)

	type fields struct {
		MemberName string
		Hostname   string
		IP         string
		IPSelf     string
		Port       string
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.

		{
			fields: fields{
				MemberName: testOne.name,
				Hostname:   testOne.host,
				IPSelf:     testOne.ip,
				Port:       testOne.port,
			},
			name: testOne.testName,
			want: testOne.data,
		},
		{
			fields: fields{
				MemberName: testTwo.name,
				Hostname:   testTwo.host,
				IPSelf:     testTwo.ip,
				Port:       testTwo.port,
			},
			name: testTwo.testName,
			want: testTwo.data,
		},
		{
			fields: fields{
				MemberName: testHostnameTooLongGetsTruncated.name,
				Hostname:   testHostnameTooLongGetsTruncated.host,
				IPSelf:     testHostnameTooLongGetsTruncated.ip,
				Port:       testHostnameTooLongGetsTruncated.port,
			},
			name: testHostnameTooLongGetsTruncated.testName,
			want: testHostnameTooLongGetsTruncated.data,
		},
		{
			fields: fields{
				MemberName: testMaxIP.name,
				Hostname:   testMaxIP.host,
				IPSelf:     testMaxIP.ip,
				Port:       testMaxIP.port,
			},
			name: testMaxIP.testName,
			want: testMaxIP.data,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip4Address, _ := NewIP4Address(tt.fields.IPSelf)
			m := Member{
				MemberName: tt.fields.MemberName,
				Hostname:   tt.fields.Hostname,
				IP:         &ip4Address,
				IPSelf:     &ip4Address,
				PortSelf:   tt.fields.Port,
			}
			if got := HelloMessage.CreateMemberMessage(&m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateMemberMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func truncate(valueToTruncate string, size int) string {
	bytes := []byte(valueToTruncate)
	newBytes := bytes
	if len(bytes) >= size {
		newBytes = bytes[0:size]
	}
	return string(newBytes)
}

func CreateBasicTestData(memberName string, hostname string, ipAddress string, port string) []byte {
	memberNameBytes := []byte(memberName)
	hostnameBytes := []byte(hostname)
	ip4Address, _ := NewIP4Address(ipAddress)
	ipAddressBytes := ip4Address.ToByteArray()
	portBytes := []byte(port)

	var basicTestWant = make([]byte, 0, HelloMessage.HeaderSize())
	basicTestWant = append(basicTestWant, HelloPrefix)
	basicTestWant = append(basicTestWant, memberNameBytes...)
	memberNamePadding := createPadding(MemberNameField.Size, memberNameBytes)
	if len(memberNamePadding) > 0 {
		basicTestWant = append(basicTestWant, memberNamePadding...)
	}

	basicTestWant = append(basicTestWant, hostnameBytes...)
	hostnamePadding := createPadding(HostnameField.Size, hostnameBytes)
	if len(hostnamePadding) > 0 {
		basicTestWant = append(basicTestWant, hostnamePadding...)
	}

	basicTestWant = append(basicTestWant, ipAddressBytes...)
	ipAddressPadding := createPadding(IPField.Size, ipAddressBytes)
	if len(ipAddressPadding) > 0 {
		basicTestWant = append(basicTestWant, ipAddressPadding...)
	}

	basicTestWant = append(basicTestWant, portBytes...)
	portPadding := createPadding(PortField.Size, portBytes)
	if len(portPadding) > 0 {
		basicTestWant = append(basicTestWant, portPadding...)
	}

	return basicTestWant
}

func createPadding(size int, bytes []byte) []byte {
	var padding []byte
	paddingLength := size - len(bytes)
	if paddingLength > 0 {
		padding = make([]byte, paddingLength, paddingLength)
	}
	return padding
}

func TestReadMemberMessage(t *testing.T) {
	type args struct {
		rawMessage []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Member
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := ReadMemberMessage(tt.args.rawMessage, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMemberMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadMemberMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addHeaderToMessage(t *testing.T) {
	type args struct {
		message []byte
		start   int
		end     int
		data    []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appendHeaderToMessage(tt.args.message, tt.args.start, tt.args.end, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("appendHeaderToMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHeader(t *testing.T) {
	type args struct {
		rawMessage []byte
		start      int
		length     int
	}
	tests := []struct {
		name  string
		args  args
		want  []byte
		want1 int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getHeader(tt.args.rawMessage, tt.args.start, tt.args.length)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getHeader() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getHeader() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
