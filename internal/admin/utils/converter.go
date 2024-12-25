package utils

import "fmt"

func ConvertMapStringToMapInterface(m map[string]string) (res map[string]interface{}) {
	res = make(map[string]interface{})
	for key, value := range m {
		fmt.Println(key, value)
		res[key] = value
	}
	return res
}
