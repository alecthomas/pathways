package pathways

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"
)

type RouteAction func(context *Context) *Response

func (r RouteAction) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r(&Context{
		Request:  request,
		Response: writer,
	}).Write()
}

// An action that returns
func ApiNotFound(cx *Context) *Response {
	return cx.ApiError(http.StatusNotFound, "Not Found")
}

func applyHandler(handler http.Handler) RouteAction {
	return func(cx *Context) *Response {
		return ResponseFromContext(cx, func(w http.ResponseWriter) {
			handler.ServeHTTP(w, cx.Request)
		})
	}
}

func applyFunction(f interface{}, requestTemplateType interface{}) RouteAction {
	// TODO: Inspect arguments and return value of f to ensure correct types
	requestType := reflect.TypeOf(requestTemplateType)
	if requestType != nil && requestType.Kind() != reflect.Ptr {
		panic("request structure must be a pointer")
	}

	function := reflect.ValueOf(f)
	if function.Kind() != reflect.Func || !function.IsValid() {
		panic("invalid function")
	}

	return func(cx *Context) *Response {
		defer cx.Request.Body.Close()
		in := []reflect.Value{
			reflect.ValueOf(cx),
		}
		if requestTemplateType != nil {
			v := reflect.New(requestType.Elem())
			ct := cx.InferContentType("application/json")
			err := Serializers.DecodeRequest(cx.Request, ct, v.Interface())
			if err != nil {
				return cx.ApiError(http.StatusBadRequest, err.Error())
			}
			in = append(in, v)
		}
		response := function.Call(in)
		return response[0].Interface().(*Response)
	}
}

func coerce(s string, t reflect.Type) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Int:
		v, err := strconv.ParseInt(s, 10, 64)
		return reflect.ValueOf(int(v)), err
	case reflect.Int8:
		v, err := strconv.ParseInt(s, 10, 8)
		return reflect.ValueOf(int8(v)), err
	case reflect.Uint8:
		v, err := strconv.ParseUint(s, 10, 8)
		return reflect.ValueOf(uint8(v)), err
	case reflect.Int16:
		v, err := strconv.ParseInt(s, 10, 16)
		return reflect.ValueOf(int16(v)), err
	case reflect.Uint16:
		v, err := strconv.ParseUint(s, 10, 16)
		return reflect.ValueOf(uint16(v)), err
	case reflect.Int32:
		v, err := strconv.ParseInt(s, 10, 32)
		return reflect.ValueOf(int32(v)), err
	case reflect.Uint32:
		v, err := strconv.ParseUint(s, 10, 32)
		return reflect.ValueOf(uint32(v)), err
	case reflect.Int64:
		v, err := strconv.ParseInt(s, 10, 64)
		return reflect.ValueOf(int64(v)), err
	case reflect.Uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		return reflect.ValueOf(uint64(v)), err
	case reflect.Float32:
		v, err := strconv.ParseFloat(s, 32)
		return reflect.ValueOf(float32(v)), err
	case reflect.Float64:
		v, err := strconv.ParseFloat(s, 64)
		return reflect.ValueOf(v), err
	case reflect.String:
		return reflect.ValueOf(s), nil
	}
	return reflect.ValueOf(s), errors.New("unsupported argument type " + t.String())
}
