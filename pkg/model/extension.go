package model

import (
	"math/rand"
	"reflect"

	"github.com/jinzhu/gorm"
)

//SelectFields is function that projects fields from domain struct to sql column names
func SelectFields(object interface{}) []string {
	value := reflect.ValueOf(object)
	elem := reflect.Indirect(value)
	typeOfT := elem.Type()
	result := make([]string, 0)
	for i := 0; i < elem.NumField(); i++ {
		val, ok := typeOfT.Field(i).Tag.Lookup("field")
		if ok && val == "ignore" {
			continue
		}
		if ok && val == "extend" {
			extend := SelectFields(elem.Field(i).Interface())
			result = append(result, extend...)
			continue
		}
		result = append(result, gorm.ToColumnName(typeOfT.Field(i).Name))
	}

	return result
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//RandStringRunes creates random string with specified length
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
