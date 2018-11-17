// REF: https://github.com/wunderlist/ttlcache
// REF: https://github.com/bitly/go-simplejson

package cache

import (
	"errors"
	"log"
	"reflect"
	"sync"
	"time"
)

// Item represents a record in the cache map
type Item struct {
	sync.RWMutex
	data    interface{}
	expires *time.Time
}

func (item *Item) touch(duration time.Duration) {
	item.Lock()
	expiration := time.Now().Add(duration)
	item.expires = &expiration
	item.Unlock()
}

func (item *Item) expired() bool {
	var value bool
	item.RLock()
	if item.expires == nil {
		value = true
	} else {
		value = item.expires.Before(time.Now())
	}
	item.RUnlock()
	return value
}

// Float64 coerces into a float64
func (i *Item) Float64() (float64, error) {
	switch i.data.(type) {
	case float32, float64:
		return reflect.ValueOf(i.data).Float(), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(i.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(i.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Int coerces into an int
func (i *Item) Int() (int, error) {
	switch i.data.(type) {
	case float32, float64:
		return int(reflect.ValueOf(i.data).Float()), nil
	case int, int8, int16, int32, int64:
		return int(reflect.ValueOf(i.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(i.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Int64 coerces into an int64
func (i *Item) Int64() (int64, error) {
	switch i.data.(type) {
	case float32, float64:
		return int64(reflect.ValueOf(i.data).Float()), nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(i.data).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(i.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Uint64 coerces into an uint64
func (i *Item) Uint64() (uint64, error) {
	switch i.data.(type) {
	case float32, float64:
		return uint64(reflect.ValueOf(i.data).Float()), nil
	case int, int8, int16, int32, int64:
		return uint64(reflect.ValueOf(i.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(i.data).Uint(), nil
	}
	return 0, errors.New("invalid value type")
}

// Map type asserts to `map`
func (i *Item) Map() (map[string]interface{}, error) {
	if m, ok := (i.data).(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

// Array type asserts to an `array`
func (i *Item) Array() ([]interface{}, error) {
	if a, ok := (i.data).([]interface{}); ok {
		return a, nil
	}
	return nil, errors.New("type assertion to []interface{} failed")
}

// Bool type asserts to `bool`
func (i *Item) Bool() (bool, error) {
	if s, ok := (i.data).(bool); ok {
		return s, nil
	}
	return false, errors.New("type assertion to bool failed")
}

// String type asserts to `string`
func (i *Item) String() (string, error) {
	if s, ok := (i.data).(string); ok {
		return s, nil
	}
	return "", errors.New("type assertion to string failed")
}

// Bytes type asserts to `[]byte`
func (i *Item) Bytes() ([]byte, error) {
	if s, ok := (i.data).([]byte); ok {
		return []byte(s), nil
	}
	return nil, errors.New("type assertion to []byte failed")
}

// StringArray type asserts to an `array` of `string`
func (i *Item) StringArray() ([]string, error) {
	arr, err := i.Array()
	if err != nil {
		return nil, err
	}
	retArr := make([]string, 0, len(arr))
	for _, a := range arr {
		if a == nil {
			retArr = append(retArr, "")
			continue
		}
		s, ok := a.(string)
		if !ok {
			return nil, err
		}
		retArr = append(retArr, s)
	}
	return retArr, nil
}

// MustArray guarantees the return of a `[]interface{}` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
//		for i, v := range js.Get("results").MustArray() {
//			fmt.Println(i, v)
//		}
func (i *Item) MustArray(args ...[]interface{}) []interface{} {
	var def []interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustArray() received too many arguments %d", len(args))
	}

	a, err := i.Array()
	if err == nil {
		return a
	}

	return def
}

// MustMap guarantees the return of a `map[string]interface{}` (with optional default)
//
// useful when you want to interate over map values in a succinct manner:
//		for k, v := range js.Get("dictionary").MustMap() {
//			fmt.Println(k, v)
//		}
func (i *Item) MustMap(args ...map[string]interface{}) map[string]interface{} {
	var def map[string]interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustMap() received too many arguments %d", len(args))
	}

	a, err := i.Map()
	if err == nil {
		return a
	}

	return def
}

// MustString guarantees the return of a `string` (with optional default)
//
// useful when you explicitly want a `string` in a single value return context:
//     myFunc(js.Get("param1").MustString(), js.Get("optional_param").MustString("my_default"))
func (i *Item) MustString(args ...string) string {
	var def string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustString() received too many arguments %d", len(args))
	}

	s, err := i.String()
	if err == nil {
		return s
	}

	return def
}

// MustInt guarantees the return of an `int` (with optional default)
//
// useful when you explicitly want an `int` in a single value return context:
//     myFunc(js.Get("param1").MustInt(), js.Get("optional_param").MustInt(5150))
func (i *Item) MustInt(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt() received too many arguments %d", len(args))
	}

	iv, err := i.Int()
	if err == nil {
		return iv
	}

	return def
}

// MustFloat64 guarantees the return of a `float64` (with optional default)
//
// useful when you explicitly want a `float64` in a single value return context:
//     myFunc(js.Get("param1").MustFloat64(), js.Get("optional_param").MustFloat64(5.150))
func (i *Item) MustFloat64(args ...float64) float64 {
	var def float64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustFloat64() received too many arguments %d", len(args))
	}

	f, err := i.Float64()
	if err == nil {
		return f
	}

	return def
}

// MustBool guarantees the return of a `bool` (with optional default)
//
// useful when you explicitly want a `bool` in a single value return context:
//     myFunc(js.Get("param1").MustBool(), js.Get("optional_param").MustBool(true))
func (i *Item) MustBool(args ...bool) bool {
	var def bool

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustBool() received too many arguments %d", len(args))
	}

	b, err := i.Bool()
	if err == nil {
		return b
	}

	return def
}

// MustInt64 guarantees the return of an `int64` (with optional default)
//
// useful when you explicitly want an `int64` in a single value return context:
//     myFunc(js.Get("param1").MustInt64(), js.Get("optional_param").MustInt64(5150))
func (i *Item) MustInt64(args ...int64) int64 {
	var def int64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt64() received too many arguments %d", len(args))
	}

	iv, err := i.Int64()
	if err == nil {
		return iv
	}

	return def
}

// MustUInt64 guarantees the return of an `uint64` (with optional default)
//
// useful when you explicitly want an `uint64` in a single value return context:
//     myFunc(js.Get("param1").MustUint64(), js.Get("optional_param").MustUint64(5150))
func (i *Item) MustUint64(args ...uint64) uint64 {
	var def uint64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustUint64() received too many arguments %d", len(args))
	}

	iv, err := i.Uint64()
	if err == nil {
		return iv
	}

	return def
}
