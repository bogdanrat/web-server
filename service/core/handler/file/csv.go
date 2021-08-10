package file

import (
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"reflect"
	"time"
)

const (
	csvTag        = "csv"
	excelTag      = "excel"
	dateFormat    = "2006-01-02"
	csvFileName   = "Files.csv"
	excelFileName = "Files.xlsx"
)

// GetFilesAsCSVRecords returns the files as CSV writeable records
func GetFilesAsCSVRecords(files []*models.GetFilesResponse, headers []string) [][]string {
	records := [][]string{
		headers,
	}

	// iterate invoices one by one
	for _, file := range files {
		fileStruct := reflect.ValueOf(file).Elem()
		// make sure we are dealing with a struct
		if fileStruct.Kind() == reflect.Struct {
			values := make([]string, 0)
			// iterate struct fields
			for i := 0; i < fileStruct.NumField(); i++ {
				value := ""

				field := fileStruct.Field(i)
				fieldValue := field.Interface()
				fieldType := field.Type()
				fieldKind := field.Kind()

				// parse time if field is of type time.Time
				if fieldType.AssignableTo(reflect.TypeOf(&time.Time{})) {
					value = fieldValue.(*time.Time).Format(dateFormat)
				}

				if fieldKind == reflect.Uint64 {
					value = fmt.Sprintf("%d", fieldValue)
				}

				// take value as it it in case of strings
				if fieldKind == reflect.String {
					value = fieldValue.(string)
				}

				// if value variable is not empty, then we have a field of interest and we save it
				if value != "" {
					values = append(values, value)
				}
			}

			// append one row of values
			records = append(records, values)
		}
	}

	return records
}
