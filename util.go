package goswarm

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/openai/openai-go"
)

func debugPrint(debug bool, args ...interface{}) {
	if !debug {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprint(args...)
	fmt.Printf("\033[97m[\033[90m%s\033[97m]\033[90m %s\033[0m\n", timestamp, message)
}

func functionToJSON(f interface{}) (openai.ChatCompletionToolParam, error) {
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

	parameters := map[string]map[string]string{}
	required := []string{}
	for i := 0; i < funcType.NumIn(); i++ {
		param := funcType.In(i)

		// Context Variables 은 제외
		if param == reflect.TypeOf(Args{}) {
			continue
		}

		paramType, ok := typeMap[param]
		if !ok {
			paramType = "string"
		}
		paramName := funcType.In(i).Name() // fmt.Sprintf("param%d", i+1) or funcType.In(i).Name() if it’s named
		parameters[paramName] = map[string]string{"type": paramType}
		// TODO: 인자 설명부 처리

		// TODO: 필수 인자만 처리 (기본값 미지정 인자)
		required = append(required, paramName)
	}

	fnVal := reflect.ValueOf(f)
	fnName := runtime.FuncForPC(fnVal.Pointer()).Name()
	fnName = funcNameNormalization(fnName)

	result := openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String(fnName),
			Description: openai.String("Function signature"), // TODO: 함수 설명 부 추출방법이 없어서 임시로 이렇게 처리
			Parameters: openai.F(openai.FunctionParameters{
				"type":       "object",
				"properties": parameters,
				"required":   required,
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

func callFuncByArgs(f any, args Args) any {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return ""
	}

	var in []reflect.Value
	for i := 0; i < v.Type().NumIn(); i++ {
		argType := v.Type().In(i)
		argName := argType.Name()

		if argType == reflect.TypeOf(Args{}) {
			argName = __CTX_VARS_NAME__
		}

		if argValue, ok := args[argName]; ok {
			in = append(in, reflect.ValueOf(argValue))
		} else {
			defaultValue := reflect.Zero(argType)
			in = append(in, defaultValue)
		}
	}

	out := v.Call(in)
	if len(out) == 0 {
		return ""
	}

	return out[0].Interface()
}

// go 함수 "." 문자를 "_"문자로 변환
func funcNameNormalization(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}
