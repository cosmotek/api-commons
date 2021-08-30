package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
)

/*
MarshalCSV converts an array of typed items (e.g. []Vue)
to CSV, returning an io.Reader and error. By default, the CSV
header is generated using the field names of the type in the array.
In order to use a custom column name in the header, use the "csv"
struct tag.

Code Example:
	type Person struct {
		Name  string `csv:"name"`
		Email string `csv:"email"`
		Age   int    `csv:"age"`
		Male  bool   `csv:"male"`
	}

	func main() {
		data := []interface{}{
			Person{
				Name:  "seth",
				Email: "some email",
				Age:   20,
				Male:  true,
			},
			Person{
				Name:  "justin",
				Email: "some email",
				Age:   32,
				Male:  true,
			},
			Person{
				Name:  "mike",
				Email: "some email",
				Age:   35,
				Male:  true,
			},
		}

		output, err := MarshalCSV(data)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(output))
	}
*/
func MarshalCSV(input []interface{}) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}

	typ := reflect.TypeOf(input[0])
	numFields := typ.NumField()

	header := make([]string, numFields)
	for i := 0; i < numFields; i++ {
		name := typ.Field(i).Tag.Get("csv")
		if name == "" {
			header[i] = typ.Field(i).Name
		} else {
			header[i] = name
		}
	}

	records := make([][]string, len(input)+1)
	records[0] = header

	for i, item := range input {
		valOf := reflect.ValueOf(item)
		row := make([]string, numFields)

		for i := 0; i < numFields; i++ {
			row[i] = fmt.Sprintf("%v", valOf.Field(i).Interface())
		}

		records[i+1] = row
	}

	output := bytes.NewBuffer([]byte{})
	err := csv.NewWriter(output).WriteAll(records)
	if err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}
