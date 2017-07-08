package evaluator

import (
	"reflect"
)

// Call applies a function to arguments, converting them as necessary.
//
// The conversion follows Liquid (Ruby?) semantics, which are more aggressive than
// Go conversion.
//
// The function should return one or two values; the second value,
// if present, should be an error.
func Call(fn reflect.Value, args []interface{}) (interface{}, error) {
	in := convertArguments(fn, args)
	results := fn.Call(in)
	return convertResults(results)
}

func convertResults(results []reflect.Value) (interface{}, error) {
	if len(results) > 1 && results[1].Interface() != nil {
		switch e := results[1].Interface().(type) {
		case error:
			return nil, e
		default:
			panic(e)
		}
	}
	return results[0].Interface(), nil
}

// Convert args to match the input types of function fn.
func convertArguments(fn reflect.Value, args []interface{}) (results []reflect.Value) {
	rt := fn.Type()
	results = make([]reflect.Value, rt.NumIn())
	for i, arg := range args {
		if i >= rt.NumIn() {
			// ignore extra arguments
			break
		}
		typ := rt.In(i)
		switch {
		case isDefaultFunctionType(typ):
			results[i] = makeConstantFunction(typ, arg)
		case arg == nil:
			results[i] = reflect.Zero(typ)
		default:
			results[i] = reflect.ValueOf(MustConvert(arg, typ))
		}
	}
	// create zeros and default functions for parameters without arguments
	for i := len(args); i < rt.NumIn(); i++ {
		typ := rt.In(i)
		switch {
		case isDefaultFunctionType(typ):
			results[i] = makeIdentityFunction(typ)
		default:
			results[i] = reflect.Zero(typ)
		}
	}
	return
}

func isDefaultFunctionType(typ reflect.Type) bool {
	return typ.Kind() == reflect.Func && typ.NumIn() == 1 && typ.NumOut() == 1
}

func makeConstantFunction(typ reflect.Type, arg interface{}) reflect.Value {
	return reflect.MakeFunc(typ, func(args []reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(MustConvert(arg, typ.Out(0)))}
	})
}

func makeIdentityFunction(typ reflect.Type) reflect.Value {
	return reflect.MakeFunc(typ, func(args []reflect.Value) []reflect.Value {
		return args
	})
}