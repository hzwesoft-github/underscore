package lang

import (
	"fmt"
	"testing"
)

/*
Validate struct by field's tag. the tags currently supported as below:

string: required, ip4addr, ip6addr, ip4net, ip6net, netmask, macaddr, domain, date, time, datetime

int, float: nonzero, positive, negative, gte0, lte0

uint: nonzero, positive

pointer: required and actual type's tag
*/
type ValidateTestSingle struct {
	Str_Required   string  `v:"required"`
	Str_Ip4        string  `v:"ip4addr"`
	Str_Ip6        string  `v:"ip6addr"`
	Str_Ip4net     string  `v:"ip4net"`
	Str_Ip6net     string  `v:"ip6net"`
	Str_Netmask    string  `v:"netmask"`
	Str_Mac        string  `v:"macaddr"`
	Str_Domain     string  `v:"domain"`
	Str_Date       string  `v:"date"`
	Str_Time       string  `v:"time"`
	Str_DateTime   string  `v:"datetime"`
	Int_Nonzero    int     `v:"nonzero"`
	Int_Positive   int8    `v:"positive"`
	Int_Negative   int16   `v:"negative"`
	Int_Gte0       int32   `v:"gte0"`
	Int_Lte0       int64   `v:"lte0"`
	Float_Nonzero  float32 `v:"nonzero"`
	Float_Positive float32 `v:"positive"`
	Float_Negative float32 `v:"negative"`
	Float_Gte0     float64 `v:"gte0"`
	Float_Lte0     float64 `v:"lte0"`
	Pointer_Ip4    *string `v:"ip4addr"`
}

func TestNetmask(t *testing.T) {
	s := &ValidateTestSingle{}
	output(Validate(s))

	s.Str_Required = "astring"
	s.Str_Ip4 = "114.114.114.1111"
	output(Validate(s))

	s.Str_Ip4 = "240e:390:e05:8080:75a1:d7d3:9d6e:b23a"
	output(Validate(s))

	s.Str_Ip4 = "114.114.114.114"
	s.Str_Ip6 = "114.114.114.114"
	output(Validate(s))

	s.Str_Ip6 = "240e:390:e05:8080:75a1:d7d3:9d6e:b23a"
	s.Str_Ip4net = "192.168.100.0/33"
	output(Validate(s))

	s.Str_Ip4net = "192.168.100.0/24"
	s.Str_Ip6net = "192.168.100.0/24"
	output(Validate(s))

	s.Str_Ip6net = "2001:db8::/32"
	s.Str_Netmask = "255.255.255.1"
	output(Validate(s))

	s.Str_Netmask = "255.255.255.0"
	s.Str_Mac = "11:22:33:44:55:YY"
	output(Validate(s))

	s.Str_Mac = "11:22:33:44:55:66"
	s.Str_Domain = "1"
	output(Validate(s))

	s.Str_Mac = "11:22:33:44:55:66"
	s.Str_Domain = "~!!!.baidu.com"
	output(Validate(s))
}

func output(err error) {
	fmt.Println(err)
}
