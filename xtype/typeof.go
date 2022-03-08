package xtype

import (
	"fmt"
	"reflect"
)

func Of(input interface{}) string {
	if input == nil {
		return "nil"
	}

	if t := reflect.TypeOf(input); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

func FullOf(input interface{}) string {
	if input == nil {
		return "nil"
	}

	if t := reflect.TypeOf(input); t.Kind() == reflect.Ptr {
		return fmt.Sprintf("*%s/%s", t.Elem().PkgPath(), t.Elem().Name())
	} else {
		return fmt.Sprintf("%s/%s", t.Elem().PkgPath(), t.Elem().Name())
	}
}
