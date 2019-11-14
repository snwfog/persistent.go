package util

import (
	"reflect"
)

func IsNil(itf interface{}) bool {
	return itf == nil || (reflect.ValueOf(itf).Kind() == reflect.Ptr && reflect.ValueOf(itf).IsNil())
}

