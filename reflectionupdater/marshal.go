package reflectionupdater

import (
	"reflect"
	"strings"
)

type SetRecord map[string]interface{}

func ToUpdateSetRecord(input interface{}) SetRecord {
	rtype := reflect.TypeOf(input)
	rval := reflect.ValueOf(input)

	totalFields := rtype.NumField()
	setRecord := map[string]interface{}{}

	for i := 0; i < totalFields; i++ {
		rfield := rtype.Field(i)
		structTag := rfield.Tag.Get("db")

		// check if field should be ignored
		if structTag != "-" {
			fieldName := rfield.Name
			omitEmpty := false

			// parse the struct tags to set flags created above
			tags := strings.Split(structTag, ",")
			if len(tags) > 0 {
				fieldName = tags[0]
			}
			
			if len(tags) > 1 {
				omitEmpty = tags[1] == "omitempty"
			}

			rvfield := rval.Field(i)
			if rvfield.CanInterface() {
				fieldVal := rvfield.Interface()
				if (rvfield.Kind() != reflect.Ptr || !rvfield.IsNil()) && ((omitEmpty && !rvfield.IsZero()) || !omitEmpty) {
					setRecord[fieldName] = fieldVal
				}
			}
		}
	}

	return setRecord
} 