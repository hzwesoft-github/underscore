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
	Str_URL        string  `v:"url"`
	Str_Port       string  `v:"port"`
	Str_Date       string  `v:"date"`
	Str_Time       string  `v:"time"`
	Str_DateTime   string  `v:"datetime"`
	Int_Nonzero    int     `v:"nonzero"`
	Int_Positive   int8    `v:"positive"`
	Int_Negative   int16   `v:"negative"`
	Int_Gte0       int32   `v:"gte0"`
	Int_Lte0       int64   `v:"lte0"`
	Int_Port       int     `v:"port"`
	Uint_Nonzero   uint    `v:"nonzero"`
	Uint_Positive  uint32  `v:"positive"`
	Uint_Port      uint64  `v:"port"`
	Float_Nonzero  float32 `v:"nonzero"`
	Float_Positive float32 `v:"positive"`
	Float_Negative float32 `v:"negative"`
	Float_Gte0     float64 `v:"gte0"`
	Float_Lte0     float64 `v:"lte0"`
	Pointer_Ip4    *string `v:"ip4addr"`
}

func TestSimpleValidate(t *testing.T) {
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

	s.Str_Domain = "~!!!.baidu.com"
	output(Validate(s))

	s.Str_Domain = ".baidu.com.cn"
	output(Validate(s))

	s.Str_Domain = "baidu.com"
	s.Str_URL = "http://1"
	output(Validate(s))

	s.Str_URL = "www.163.com"
	output(Validate(s))

	s.Str_URL = "https://www.163.com/u/r/l/"
	s.Str_Port = "0"
	output(Validate(s))

	s.Str_Port = "65536"
	output(Validate(s))

	s.Str_Port = "65535"
	s.Str_Date = "2022-13-02"
	output(Validate(s))

	s.Str_Date = "2022-02-01"
	s.Str_Time = "12:00:61"
	output(Validate(s))

	s.Str_Time = "12:00:01"
	s.Str_DateTime = "2022-13-01 12:00:01"
	output(Validate(s))

	s.Str_DateTime = "2022-01-01 12:00:01"
	s.Int_Nonzero = 0
	output(Validate(s))

	s.Int_Nonzero = 1
	s.Int_Positive = -1
	output(Validate(s))

	s.Int_Positive = 1
	s.Int_Negative = 0
	output(Validate(s))

	s.Int_Negative = -1
	s.Int_Gte0 = -1
	output(Validate(s))

	s.Int_Gte0 = 0
	s.Int_Lte0 = 1
	output(Validate(s))

	s.Int_Lte0 = 0
	s.Int_Port = 0
	output(Validate(s))

	s.Int_Port = 65536
	output(Validate(s))

	s.Int_Port = 65535
	s.Uint_Nonzero = 0
	output(Validate(s))

	s.Uint_Nonzero = 1
	s.Uint_Positive = 0
	output(Validate(s))

	s.Uint_Positive = 1
	s.Uint_Port = 65536
	output(Validate(s))

	s.Uint_Port = 65535
	s.Float_Nonzero = 0
	output(Validate(s))

	s.Float_Nonzero = 1
	s.Float_Positive = -1
	output(Validate(s))

	s.Float_Positive = 1
	s.Float_Negative = 0
	output(Validate(s))

	s.Float_Negative = -1
	s.Float_Gte0 = -1
	output(Validate(s))

	s.Float_Gte0 = 0
	s.Float_Lte0 = 1
	output(Validate(s))

	s.Float_Lte0 = -1
	str := "114.114.111"
	s.Pointer_Ip4 = &str
	output(Validate(s))

	str = "114.114.114.114"
	output(Validate(s))
}

func output(err error) {
	fmt.Println(err)
}

type ValidateTestSlice struct {
	StrSlice []string `v:"required,ip4addr"`
}

func TestSliceValidate(t *testing.T) {
	s := ValidateTestSlice{}
	output(Validate(s))

	s.StrSlice = make([]string, 0)
	output(Validate(s))

	s.StrSlice = append(s.StrSlice, "114.114.114")
	output(Validate(s))

	s.StrSlice = []string{}
	s.StrSlice = append(s.StrSlice, "114.114.114.114", "114.114.114.")
	output(Validate(s))

	s.StrSlice = []string{}
	s.StrSlice = append(s.StrSlice, "114.114.114.114", "114.114.114.115")
	output(Validate(s))
}

type ValidateTestPointer struct {
	StrPointer *string `v:"required,ip4addr"`
}

func TestPointerValidate(t *testing.T) {
	s := &ValidateTestPointer{}
	output(Validate(s))

	str := "114.114.111"
	s.StrPointer = &str
	output(Validate(s))

	str = "114.114.114.114"
	output(Validate(s))
}

type ValidateCustom struct {
	StrPointer *string `v-custom:"required,ip4addr"`
}

func TestPointerCustom(t *testing.T) {
	s := &ValidateCustom{}
	output(ValidateTag(s, "v-custom"))

	str := "114.114.112"
	s.StrPointer = &str
	output(ValidateTag(s, "v-custom"))

	str = "114.114.114.114"
	output(ValidateTag(s, "v-custom"))
}
