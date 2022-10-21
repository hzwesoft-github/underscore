package lang

import (
	"errors"
	"reflect"
)

func Validate(obj any) error {
	if obj == nil {
		return nil
	}

	return doValidate(reflect.ValueOf(obj))
}

func doValidate(value reflect.Value) error {
	if value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
		return doValidate(value.Elem())
	}

	if value.Kind() != reflect.Struct {
		return errors.New("ng: value must be struct")
	}

	typ := value.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if vtag, ok := field.Tag.Lookup("v"); ok {
			if IsBlank(vtag) {
				continue
			}

			fieldValue := value.Field(i)

			// TODO
			switch fieldValue.Kind() {
			case reflect.Slice:
				if err := validateSlice(fieldValue); err != nil {
					return err
				}

			case reflect.Pointer, reflect.Interface:
			case reflect.String:
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			case reflect.Float32, reflect.Float64:
			case reflect.Bool:
			default:
			}

		}
	}

	return nil
}

func validateSlice(value reflect.Value) error {
	// TODO
	return nil
}

func validateString(value reflect.Value) error {
	// TODO
	return nil
}
