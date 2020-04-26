package objx

import "reflect"

type Option struct {
	Value reflect.Value
	ConvertToMap
}

type ConvertToMap func(interface{}) map[string]interface{}
