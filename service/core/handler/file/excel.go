package file

import (
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/xuri/excelize/v2"
)

const (
	defaultSheetName = "Sheet1"
	finalSheetName   = "Files"
)

func WriteFilesToExcel(files []*models.GetFilesResponse, headers []string) (*excelize.File, error) {
	excelFile := excelize.NewFile()
	headerStyle, _ := excelFile.NewStyle(`{"font":{"bold": true, "sz": 16, "size": 16}, "alignment": {"horizontal": "center", "indent": 0, "relative_indent": 0, "justify_last_line": false, "reading_order": 0, "shrink_to_fit": false, "vertical": "center", "wrap_text": false}, "protection": {"locked": true}}`)

	for i, header := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s1", col)
		excelFile.SetCellValue(defaultSheetName, cell, header)
		excelFile.SetCellStyle(defaultSheetName, cell, cell, headerStyle)
	}

	for i, file := range files {
		excelFile.SetCellValue(defaultSheetName, fmt.Sprintf("A%d", i+2), file.Key)
		excelFile.SetCellValue(defaultSheetName, fmt.Sprintf("B%d", i+2), file.LastModified)
		excelFile.SetCellValue(defaultSheetName, fmt.Sprintf("C%d", i+2), file.Size)
		excelFile.SetCellValue(defaultSheetName, fmt.Sprintf("D%d", i+2), file.StorageClass)
	}

	excelFile.SetSheetName(defaultSheetName, finalSheetName)

	return excelFile, nil
}
