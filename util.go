package goswarm

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/chiwooi/go-swarm/types"
	"github.com/openai/openai-go"
)

func debugPrint(debug bool, fmtmsg string, args ...interface{}) {
	if !debug {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(fmtmsg, args...)
	fmt.Printf("\033[97m[\033[90m%s\033[97m]\033[90m %s\033[0m\n", timestamp, message)
}

// Convert the function to a JSON object.
func functionToJSON(ctx Context, f any) (openai.ChatCompletionToolParam, error) {
	typeMap := map[reflect.Type]string{
		reflect.TypeOf(""):     "string",
		reflect.TypeOf(0):      "integer",
		reflect.TypeOf(0.0):    "number",
		reflect.TypeOf(true):   "boolean",
		reflect.TypeOf([]interface{}{}): "array",
		reflect.TypeOf(map[string]interface{}{}): "object",
	}

	funcType := reflect.TypeOf(f)
	if funcType.Kind() != reflect.Func {
		return openai.ChatCompletionToolParam{}, fmt.Errorf("provided value is not a function")
	}

	ctx = getCallFuncDesc(ctx, f)

	parameters := map[string]map[string]string{}
	requireds := []string{}
	for i := 0; i < funcType.NumIn(); i++ {
		param := funcType.In(i)

		switch param.Kind() {
		case reflect.Struct:
			switch param.Name() {
			case "Context":
			default:
				for j := 0; j < param.NumField(); j++ {
					field := param.Field(j)
					fieldType, ok := typeMap[field.Type]
					if !ok {
						fieldType = "string"
					}

					desc := field.Tag.Get("desc")
					required := field.Tag.Get("required")

					parameters[field.Name] = map[string]string{
						"type": fieldType,
						"description": desc,
					}

					if strings.ToLower(required) == "true" {
						requireds = append(requireds, field.Name)
					}
				}
			}
		}
	}

	fnVal := reflect.ValueOf(f)
	fnName := runtime.FuncForPC(fnVal.Pointer()).Name()
	fnName = funcNameNormalization(fnName)

	result := openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String(fnName),
			Description: openai.String(ctx.GetDescription()),
			Parameters:  openai.F(openai.FunctionParameters{
				"type":       "object",
				"properties": parameters,
				"required":   requireds,
			}),
		}),
	}

	return result, nil
}

func hasArgInFunc(f any, name string) bool {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return false
	}

	for i := 0; i < v.Type().NumIn(); i++ {
		if v.Type().In(i).Name() == name {
			return true
		}
	}

	return false
}

// Call the function with the specified arguments.
func callFuncByArgs(ctx Context, f any, args types.ContextVariables) any {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return ""
	}

	var in []reflect.Value

	for i := 0; i < v.Type().NumIn(); i++ {
		param := v.Type().In(i)

		switch param.Kind() {
		case reflect.Struct, reflect.Interface:
			switch param.Name() {
			case "Context":
				in = append(in, reflect.ValueOf(ctx))
				continue
			default:
				defaultValue := reflect.New(param)

				for j := 0; j < param.NumField(); j++ {
					field := param.Field(j)

					// 인자값이 있으면 해당 값으로 설정, 없으면 기본값으로 설정
					if argValue, ok := args[field.Name]; ok {
						setFieldValue(defaultValue.Interface(), field.Name, argValue)
					} else {
						argDefValue := reflect.Zero(field.Type)
						setFieldValue(defaultValue.Interface(), field.Name, argDefValue)
					}
				}

				in = append(in, defaultValue.Elem())
			}
		}
	}

	out := v.Call(in)
	if len(out) == 0 {
		return ""
	}

	return out[0].Interface()
}

// Changes the value of the member variable named fieldName in the specified structure variable to the designated value.
func setFieldValue(obj interface{}, fieldName string, value interface{}) error {
	// Check if obj is a pointer.
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("obj must be a non-nil pointer")
	}

	// Access the actual value pointed to by the pointer.
	v = v.Elem()

	// Find the field by its name.
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("no such field: %s", fieldName)
	}

	// Check if the field is modifiable.
	if !field.CanSet() {
		return fmt.Errorf("cannot set field: %s", fieldName)
	}

	// Check if the field's type is compatible with the value, then set it.
	val := reflect.ValueOf(value)
	if field.Type() != val.Type() {
		return fmt.Errorf("provided value type doesn't match field type")
	}

	// set value to field variable
	field.Set(val)
	return nil
}


func getCallFuncDesc(ctx Context, f any) Context {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return ctx
	}

	var in []reflect.Value
	for i := 0; i < v.Type().NumIn(); i++ {
		argType := v.Type().In(i)
		if argType.Name() == "Context" {
			in = append(in, reflect.ValueOf(ctx))
		} else {
			defaultValue := reflect.Zero(argType)
			in = append(in, defaultValue)
		}
	}

	v.Call(in)

	return ctx
}


// go 함수 "." 문자를 "_"문자로 변환
func funcNameNormalization(name string) string {
	name = strings.TrimPrefix(name, "command-line-arguments.") // remove for global variable prefix
	name = strings.ReplaceAll(name, ".", "_")
	return name
}
