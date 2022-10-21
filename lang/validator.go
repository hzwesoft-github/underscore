package lang

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strings"
	"time"
)

/*
Validate struct by field's tag. the tags currently supported as below:

string: required, ip4addr, ip6addr, ip4net, ip6net, netmask, macaddr, domain, date, datetime, time

int, uint, float: nonzero, positive, negative, gte0, lte0

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
				return fmt.Errorf("ng: %s is required", value.Type().Name())
			}
		}

		if err := validateValue(field, value.Elem(), vtag); err != nil {
			return err
		}

	case reflect.String:
		// required, ip4addr, ip6addr, ip4net, ip6net, netmask, macaddr, domain, date, datetime, time
		str := value.String()
		if strings.Contains(vtag, "required") {
			if IsBlank(str) {
				return fmt.Errorf("ng: %s is required", value.Type().Name())
			}
		}

		if IsBlank(str) {
			return nil
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
		if strings.Contains(vtag, "date") {
			if _, err := time.Parse("2006-01-02", str); err != nil {
				return fmt.Errorf("ng: value of %s(%s) is not a valid date", field.Name, str)
			}
		}
		if strings.Contains(vtag, "datetime") {
			if _, err := time.Parse("2006-01-02 15:04:05", str); err != nil {
				return fmt.Errorf("ng: value of %s(%s) is not a valid datetime", field.Name, str)
			}
		}
		if strings.Contains(vtag, "time") {
			if _, err := time.Parse("15:04:05", str); err != nil {
				return fmt.Errorf("ng: value of %s(%s) is not a valid time", field.Name, str)
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// nonzero, positive, negative, gte0, lte0
		i64 := value.Int()
		if strings.Contains(vtag, "nonzero") && i64 == 0 {
			return fmt.Errorf("ng: value of %s is 0", field.Name)
		}
		if strings.Contains(vtag, "positive") && i64 <= 0 {
			return fmt.Errorf("ng: value of %s is negative or 0", field.Name)
		}
		if strings.Contains(vtag, "negative") && i64 >= 0 {
			return fmt.Errorf("ng: value of %s is positive or 0", field.Name)
		}
		if strings.Contains(vtag, "gte0") && i64 < 0 {
			return fmt.Errorf("ng: value of %s is negative", field.Name)
		}
		if strings.Contains(vtag, "lte0") && i64 > 0 {
			return fmt.Errorf("ng: value of %s is positive", field.Name)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// nonzero, positive, negative, gte0, lte0
		u64 := value.Uint()
		if strings.Contains(vtag, "nonzero") && u64 == 0 {
			return fmt.Errorf("ng: value of %s is 0", field.Name)
		}
		if strings.Contains(vtag, "positive") && u64 == 0 {
			return fmt.Errorf("ng: value of %s is 0", field.Name)
		}
	case reflect.Float32, reflect.Float64:
		// nonzero, positive, negative, gte0, lte0
		f64 := value.Float()
		if strings.Contains(vtag, "nonzero") && f64 == 0 {
			return fmt.Errorf("ng: value of %s is 0", field.Name)
		}
		if strings.Contains(vtag, "positive") && f64 <= 0 {
			return fmt.Errorf("ng: value of %s is negative or 0", field.Name)
		}
		if strings.Contains(vtag, "negative") && f64 >= 0 {
			return fmt.Errorf("ng: value of %s is positive or 0", field.Name)
		}
		if strings.Contains(vtag, "gte0") && f64 < 0 {
			return fmt.Errorf("ng: value of %s is negative", field.Name)
		}
		if strings.Contains(vtag, "lte0") && f64 > 0 {
			return fmt.Errorf("ng: value of %s is positive", field.Name)
		}
	default:
		return nil
	}

	return nil
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
	u := str
	if !strings.HasPrefix(u, "http://") || !strings.HasPrefix(u, "https://") {
		u = "http://" + u
	}
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return fmt.Errorf("ng: value of %s(%s) is not a valid domain", field.Name, str)
	}

	return nil
}
