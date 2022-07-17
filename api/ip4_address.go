package api

import (
	"fmt"
	"strconv"
	"strings"
)

type IP4Address  struct {
	A byte
	B byte
	C byte
	D byte
}

func (ip *IP4Address) ToByteArray() []byte {
	bytes := make([]byte, 4, 4)
	bytes[0] = ip.A
	bytes[1] = ip.B
	bytes[2] = ip.C
	bytes[3] = ip.D
	return bytes
}

func NewIP4Address(originalAddress string) (IP4Address, error)  {
	address := strings.Split(originalAddress, ":")[0]
	addressParts := strings.Split(address, ".")
	if len(addressParts) < 4 {
		return IP4Address{}, fmt.Errorf("Could not create IP4Address from input %s", address)
	}

	a, err := strconv.Atoi(addressParts[0])
	if err != nil {
		return IP4Address{}, err
	}
	b,err := strconv.Atoi(addressParts[1])
	if err != nil {
		return IP4Address{}, err
	}
	c,err := strconv.Atoi(addressParts[2])
	if err != nil {
		return IP4Address{}, err
	}
	d,err := strconv.Atoi(addressParts[3])
	if err != nil {
		return IP4Address{}, err
	}

	return IP4Address{
		A: byte(a),
		B: byte(b),
		C: byte(c),
		D: byte(d),
	}, nil
}

func (ip *IP4Address) String() string {
	return fmt.Sprintf("%v.%v.%v.%v", ip.A, ip.B, ip.C, ip.D)
}