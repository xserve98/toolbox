package toolbox

import (
	"fmt"
	"reflect"
)

//GetFunction returns function for provided owner and name, or error
func GetFunction(owner interface{}, name string) (interface{}, error) {
	var ownerType = reflect.TypeOf(owner)
	var method, has = ownerType.MethodByName(name)
	if !has {
		return nil, fmt.Errorf("failed to lookup %T.%v\n", owner, name)
	}
	return reflect.ValueOf(owner).MethodByName(method.Name).Interface(), nil
}

//CallFunction calls passed in function with provided parameters,it returns a function result.
func CallFunction(function interface{}, parameters ...interface{}) []interface{} {
	AssertKind(function, reflect.Func, "function")
	var functionParameters = make([]reflect.Value, 0)
	ProcessSlice(parameters, func(item interface{}) bool {
		functionParameters = append(functionParameters, reflect.ValueOf(item))
		return true
	})
	functionValue := reflect.ValueOf(function)
	var resultValues = functionValue.Call(functionParameters)
	var result = make([]interface{}, len(resultValues))
	for i, resultValue := range resultValues {
		result[i] = resultValue.Interface()
	}
	return result
}

//AsCompatibleFunctionParameters takes incompatible function parameters and converts then into provided function signature compatible
func AsCompatibleFunctionParameters(function interface{}, parameters []interface{}) ([]interface{}, error) {
	AssertKind(function, reflect.Func, "function")
	functionValue := reflect.ValueOf(function)
	funcSignature := GetFuncSignature(function)
	actualMethodSignatureLength := len(funcSignature)
	converter := Converter{}
	if actualMethodSignatureLength != len(parameters) {
		return nil, fmt.Errorf("Invalid number of parameters wanted: [%T],  had: %v", function, 0)
	}
	var functionParameters = make([]interface{}, 0)
	for i, parameterValue := range parameters {
		reflectValue := reflect.ValueOf(parameterValue)
		if reflectValue.Kind() == reflect.Slice && funcSignature[i].Kind() != reflectValue.Kind() {
			return nil, fmt.Errorf("Incompatible types expected: %v, but had %v", funcSignature[i].Kind(), reflectValue.Kind())
		} else if !reflectValue.IsValid() {
			if funcSignature[i].Kind() == reflect.Slice {
				parameterValue = reflect.New(funcSignature[i]).Interface()
				reflectValue = reflect.ValueOf(parameterValue)
			}
		}
		if reflectValue.Type() != funcSignature[i] {
			newValuePointer := reflect.New(funcSignature[i])
			err := converter.AssignConverted(newValuePointer.Interface(), parameterValue)
			if err != nil {
				return nil, fmt.Errorf("failed to assign convert %v to %v due to %v", parameterValue, newValuePointer.Interface(), err)
			}
			reflectValue = newValuePointer.Elem()
		}
		if functionValue.Type().IsVariadic() && funcSignature[i].Kind() == reflect.Slice && i+1 == len(funcSignature) {
			ProcessSlice(reflectValue.Interface(), func(item interface{}) bool {
				functionParameters = append(functionParameters, item)
				return true
			})
		} else {
			functionParameters = append(functionParameters, reflectValue.Interface())
		}
	}
	return functionParameters, nil
}

//BuildFunctionParameters builds function parameters provided in the parameterValues.
// Parameters value will be converted if needed to expected by the function signature type. It returns function parameters , or error
func BuildFunctionParameters(function interface{}, parameters []string, parameterValues map[string]interface{}) ([]interface{}, error) {
	var functionParameters = make([]interface{}, 0)
	for _, name := range parameters {
		functionParameters = append(functionParameters, parameterValues[name])
	}
	return AsCompatibleFunctionParameters(function, functionParameters)
}

//GetFuncSignature returns a function signature
func GetFuncSignature(function interface{}) []reflect.Type {
	AssertKind(function, reflect.Func, "function")
	functionValue := reflect.ValueOf(function)
	var result = make([]reflect.Type, 0)
	functionType := functionValue.Type()
	for i := 0; i < functionType.NumIn(); i++ {
		result = append(result, functionType.In(i))
	}
	return result
}
