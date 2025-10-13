package restutil

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

var ErrMissingQueryValue = errors.New("missing query value")

var (
	typeString = reflect.TypeOf("")
	typeInt    = reflect.TypeOf(0)
	typeFloat  = reflect.TypeOf(0.0)
	typeBool   = reflect.TypeOf(true)
)

type Query[T any] struct {
	values map[string]string
}

func (q Query[T]) Unmarshal() (T, error) {
	var zero T

	typ := reflect.ValueOf(zero)
	newTyp := reflect.New(typ.Type()).Elem()
	numField := newTyp.NumField()

	if len(q.values) < numField {
		return zero, fmt.Errorf("%w: expected %d values but got %d", ErrMissingQueryValue, len(q.values), numField)
	}

	for i := range numField {
		tp := typ.Type().Field(i)
		fieldName := tp.Name
		if nm, ok := tp.Tag.Lookup("query"); ok {
			fieldName = nm
		}

		val, ok := q.values[fieldName]
		if !ok {
			return zero, fmt.Errorf("%w: field %s not found in query", ErrMissingQueryValue, fieldName)
		}

		field := newTyp.Field(i)
		switch field.Type() {
		case typeInt:
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return zero, err
			}
			field.SetInt(n)
		case typeFloat:
			n, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return zero, err
			}
			field.SetFloat(n)
		case typeString:
			field.SetString(val)
		case typeBool:
			n, err := strconv.ParseBool(val)
			if err != nil {
				return zero, err
			}
			field.SetBool(n)
		}
	}

	return newTyp.Interface().(T), nil
}

func ExtractHTTPQuery[T any](req *http.Request) (T, error) {
	q := Query[T]{
		values: map[string]string{},
	}

	query := req.URL.Query()
	for k, v := range query {
		q.values[k] = v[0]
	}

	return q.Unmarshal()
}
