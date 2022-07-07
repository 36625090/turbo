package utils

import (
	"encoding/json"
	"reflect"
)

func StructToMap(obj interface{}) map[string]interface{} {
	if nil == obj{
		return nil
	}

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr{
		return StructToMap(v.Elem().Interface())
	}

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		key :=  t.Field(i)
		value := v.Field(i)
		value = getValue( value )
		switch value.Type().Kind() {
		case reflect.Interface:
			fallthrough
		case reflect.Struct:
			data[key.Name] = StructToMap( value.Interface() )
		default:
			data[key.Name] = value.Interface()
		}
	}
	return data
}


func StructToJsonMap(obj interface{})map[string]interface{}{
	if nil == obj{
		return nil
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	var ret map[string]interface{}
	if err := json.Unmarshal(data, &ret); err != nil {
		return nil
	}
	return ret
}

func getValue(obj reflect.Value) reflect.Value {
	if  !obj.IsValid() || obj.IsZero() {
		return obj
	}
	switch obj.Type().Kind() {
	case reflect.Ptr:
		return getValue(obj.Elem())
	default:
		return obj
	}
}