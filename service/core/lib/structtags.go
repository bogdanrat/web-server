package lib

import "reflect"

func GetStructTagValues(structPtr interface{}, tag string) []string {
	if structPtr == nil {
		return nil
	}

	if isPointer(structPtr) {
		if reflect.ValueOf(structPtr).IsNil() {
			// instantiate a new object
			pType := reflect.TypeOf(structPtr)
			structPtr = reflect.New(pType.Elem()).Interface()
		}
		return getTags(structPtr, tag)
	}

	return nil
}

func getTags(structPtr interface{}, tag string) []string {
	tagValues := make([]string, 0)

	ptrValue := reflect.ValueOf(structPtr)
	ptrType := ptrValue.Elem().Type()

	for i := 0; i < ptrType.NumField(); i++ {
		field := ptrType.Field(i)
		fieldTag := field.Tag.Get(tag)

		// skip fields marked with "-" tag
		if fieldTag != "" && fieldTag != "-" {
			tagValues = append(tagValues, fieldTag)
		}
	}

	return tagValues
}

func isPointer(obj interface{}) bool {
	if reflect.TypeOf(obj) == nil {
		return false
	}
	return reflect.TypeOf(obj).Kind() == reflect.Ptr
}
