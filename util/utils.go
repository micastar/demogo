package util

import (
	"reflect"
)

func ReverseItem[T any](data T) {
	v := reflect.ValueOf(data)

	n := v.Len()
	for i := 0; i < n/2; i++ {
		// Swap elements using reflection
		iv1 := v.Index(i)
		iv2 := v.Index(n - i - 1)
		temp := reflect.New(iv1.Type().Elem()).Elem()
		temp.Set(iv1.Elem())
		iv1.Elem().Set(iv2.Elem())
		iv2.Elem().Set(temp)
	}
}
