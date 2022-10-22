package lang

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
)

var (
	domainPattern *regexp2.Regexp
	urlPattern    *regexp2.Regexp
)

func init() {
	domainPattern, _ = regexp2.Compile(`^(?=^.{3,255}$)[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$`, regexp2.None)
	urlPattern, _ = regexp2.Compile(`((http|ftp|https)://)(([a-zA-Z0-9\._-]+\.[a-zA-Z]{2,6})|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}))(:[0-9]{1,4})*(/[a-zA-Z0-9\&%_\./-~-]*)?`, regexp2.None)
}

/*
Validate struct by field's tag. the tags currently supported as below:

string: required, digit, ip4addr, ip6addr, ip4net, ip6net, netmask, macaddr, domain, url, port, date, datetime, time

int: nonzero, positive, negative, gte0, lte0, port

uint: nonzero, positive, port

float: nonzero, positive, negative, gte0, lte0

pointer: required and actual type's tag
*/
func Validate(obj any) error {
	if obj == nil {
		return nil
	}

	return validate(reflect.ValueOf(obj))
}

func validate(value reflect.Value) error {
	if value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
		return validate(value.Elem())
	}

	if value.Kind() != reflect.Struct {
		return errors.New("ng: value must be struct")
	}

	return validateStruct(value)
}

func validateStruct(value reflect.Value) error {
	typ := value.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := value.Field(i)

		if field.Type.Kind() == reflect.Struct {
			if err := validateStruct(fieldValue); err != nil {
				return err
			}
			continue
		}

		if vtag, ok := field.Tag.Lookup("v"); ok {
			if IsBlank(vtag) || vtag == "-" {
				continue
			}

			if fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Array {
				if err := validateSlice(field, fieldValue, vtag); err != nil {
					return err
				}
			} else {
				if err := validateValue(field, fieldValue, vtag); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateSlice(field reflect.StructField, value reflect.Value, vtag string) error {
	if strings.Contains(vtag, "required") {
		if value.IsNil() || value.Len() == 0 {
			return fmt.Errorf("ng: %s is required", field.Name)
		}
	}

	for i := 0; i < value.Len(); i++ {
		if err := validateValue(field, value.Index(i), vtag); err != nil {
			return err
		}
	}

	return nil
}

func validateValue(field reflect.StructField, value reflect.Value, vtag string) error {
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
		if strings.Contains(vtag, "required") {
			if value.IsNil() {
				return fmt.Errorf("ng: %s is required", field.Name)
			}
		}

		if err := validateValue(field, value.Elem(), vtag); err != nil {
			return err
		}

	case reflect.String:
		str := value.String()
		if strings.Contains(vtag, "required") {
			if IsBlank(str) {
				return fmt.Errorf("ng: %s is required", field.Name)
			}
		}

		if IsBlank(str) {
			return nil
		}

		if strings.Contains(vtag, "digit") {
			if !isDigit(str) {
				return fmt.Errorf("ng: value of %s(%s) is not digit", field.Name, str)
			}
		}
		if strings.Contains(vtag, "ip4addr") {
			if err := validateIP(str, 4, field); err != nil {
				return err
			}
		}
		if strings.Contains(vtag, "ip6addr") {
			if err := validateIP(str, 6, field); err != nil {
				return err
			}
		}
		if strings.Contains(vtag, "ip4net") {
			if err := validateIPNet(str, 4, field); err != nil {
				return err
			}
		}
		if strings.Contains(vtag, "ip6net") {
			if err := validateIPNet(str, 6, field); err != nil {
				return err
			}
		}
		if strings.Contains(vtag, "netmask") {
			if err := validateNetmask(str, field); err != nil {
				return err
			}
		}
		if strings.Contains(vtag, "macaddr") {
			if err := validateMac(str, field); err != nil {
				return err
			}
		}
		if strings.Contains(vtag, "domain") {
			if err := validateDomain(str, field); err != nil {
				return err
			}
		}
		if strings.Contains(vtag, "url") {
			if err := validateURL(str, field); err != nil {
				return err
			}
		}
		if strings.Contains(vtag, "port") {
			if err := validatePort(str, field); err != nil {
				return err
			}
		}
		if strings.Contains(vtag, "date") && !strings.Contains(vtag, "datetime") {
			if _, err := time.Parse("2006-01-02", str); err != nil {
				return fmt.Errorf("ng: value of %s(%s) is not a valid date", field.Name, str)
			}
		}
		if strings.Contains(vtag, "datetime") {
			if _, err := time.Parse("2006-01-02 15:04:05", str); err != nil {
				return fmt.Errorf("ng: value of %s(%s) is not a valid datetime", field.Name, str)
			}
		}
		if strings.Contains(vtag, "time") && !strings.Contains(vtag, "datetime") {
			if _, err := time.Parse("15:04:05", str); err != nil {
				return fmt.Errorf("ng: value of %s(%s) is not a valid time", field.Name, str)
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64 := value.Int()
		if strings.Contains(vtag, "nonzero") && i64 == 0 {
			return fmt.Errorf("ng: value of %s(%d) is 0", field.Name, i64)
		}
		if strings.Contains(vtag, "positive") && i64 <= 0 {
			return fmt.Errorf("ng: value of %s(%d) is negative or 0", field.Name, i64)
		}
		if strings.Contains(vtag, "negative") && i64 >= 0 {
			return fmt.Errorf("ng: value of %s(%d) is positive or 0", field.Name, i64)
		}
		if strings.Contains(vtag, "gte0") && i64 < 0 {
			return fmt.Errorf("ng: value of %s(%d) is negative", field.Name, i64)
		}
		if strings.Contains(vtag, "lte0") && i64 > 0 {
			return fmt.Errorf("ng: value of %s(%d) is positive", field.Name, i64)
		}
		if strings.Contains(vtag, "port") && (i64 <= 0 || i64 > 65535) {
			return fmt.Errorf("ng: value of %s(%d) is not a valid port", field.Name, i64)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// nonzero, positive
		u64 := value.Uint()
		if strings.Contains(vtag, "nonzero") && u64 == 0 {
			return fmt.Errorf("ng: value of %s(%d) is 0", field.Name, u64)
		}
		if strings.Contains(vtag, "positive") && u64 == 0 {
			return fmt.Errorf("ng: value of %s(%d) is 0", field.Name, u64)
		}
		if strings.Contains(vtag, "port") && (u64 == 0 || u64 > 65535) {
			return fmt.Errorf("ng: value of %s(%d) is not a valid port", field.Name, u64)
		}
	case reflect.Float32, reflect.Float64:
		// nonzero, positive, negative, gte0, lte0
		f64 := value.Float()
		if strings.Contains(vtag, "nonzero") && f64 == 0 {
			return fmt.Errorf("ng: value of %s(%f) is 0", field.Name, f64)
		}
		if strings.Contains(vtag, "positive") && f64 <= 0 {
			return fmt.Errorf("ng: value of %s(%f) is negative or 0", field.Name, f64)
		}
		if strings.Contains(vtag, "negative") && f64 >= 0 {
			return fmt.Errorf("ng: value of %s(%f) is positive or 0", field.Name, f64)
		}
		if strings.Contains(vtag, "gte0") && f64 < 0 {
			return fmt.Errorf("ng: value of %s(%f) is negative", field.Name, f64)
		}
		if strings.Contains(vtag, "lte0") && f64 > 0 {
			return fmt.Errorf("ng: value of %s(%f) is positive", field.Name, f64)
		}
	default:
		return nil
	}

	return nil
}

func isDigit(str string) bool {
	_, err := strconv.ParseInt(str, 10, 64)
	return err == nil
}

func validateIP(str string, version int, field reflect.StructField) error {
	ip := net.ParseIP(str)
	if ip == nil {
		return fmt.Errorf("ng: value of %s(%s) is not a valid ipv%d address", field.Name, str, version)
	}

	switch version {
	case 4:
		ipv4 := ip.To4()
		if ipv4 == nil {
			return fmt.Errorf("ng: value of %s(%s) is not a valid ipv%d address", field.Name, str, version)
		}
	case 6:
		ipv4 := ip.To4()
		if ipv4 != nil {
			return fmt.Errorf("ng: value of %s(%s) is not a valid ipv%d address", field.Name, str, version)
		}
	default:
		return nil
	}

	return nil
}

func validateIPNet(str string, version int, field reflect.StructField) error {
	ip, _, err := net.ParseCIDR(str)
	if err != nil {
		return fmt.Errorf("ng: value of %s(%s) is not a valid ipv%d network address", field.Name, str, version)
	}

	switch version {
	case 4:
		ipv4 := ip.To4()
		if ipv4 == nil {
			return fmt.Errorf("ng: value of %s(%s) is not a valid ipv%d network address", field.Name, str, version)
		}
	case 6:
		ipv4 := ip.To4()
		if ipv4 != nil {
			return fmt.Errorf("ng: value of %s(%s) is not a valid ipv%d network address", field.Name, str, version)
		}
	default:
		return nil
	}

	return nil
}

func validateNetmask(str string, field reflect.StructField) error {
	ip := net.ParseIP(str)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("ng: value of %s(%s) is not a valid netmask", field.Name, str)
	}

	ip4 := ip.To4()
	bip := fmt.Sprintf("%08b", ip4[0]) + fmt.Sprintf("%08b", ip4[1]) + fmt.Sprintf("%08b", ip4[2]) + fmt.Sprintf("%08b", ip4[3])
	if strings.Contains(bip, "01") {
		return fmt.Errorf("ng: value of %s(%s) is not a valid netmask", field.Name, str)
	}

	return nil
}

func validateMac(str string, field reflect.StructField) error {
	if _, err := net.ParseMAC(str); err != nil {
		return fmt.Errorf("ng: value of %s(%s) is not a valid mac address", field.Name, str)
	}

	if len(str) != 17 {
		return fmt.Errorf("ng: value of %s(%s) is not a valid mac address", field.Name, str)
	}

	return nil
}

func validateDomain(str string, field reflect.StructField) error {
	if matched, err := domainPattern.MatchString(str); err != nil || !matched {
		return fmt.Errorf("ng: value of %s(%s) is not a valid domain", field.Name, str)
	}

	return nil
}

func validateURL(str string, field reflect.StructField) error {
	if matched, err := urlPattern.MatchString(str); err != nil || !matched {
		return fmt.Errorf("ng: value of %s(%s) is not a valid url", field.Name, str)
	}

	return nil
}

func validatePort(str string, field reflect.StructField) error {
	i64, err := strconv.ParseInt(str, 10, 64)
	if err != nil || i64 <= 0 || i64 > 65535 {
		return fmt.Errorf("ng: value of %s(%s) is not a valid port", field.Name, str)
	}

	return nil
}
