package data

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Dataset struct {
	Headers []string
	Rows    [][]string
}

func (d *Dataset) RowCount() int {
	return len(d.Rows)
}

func (d *Dataset) ColCount() int {
	return len(d.Headers)
}

func PrintDataset(ds *Dataset) {
	fmt.Println(strings.Join(ds.Headers, "\t"))
	for _, row := range ds.Rows {
		fmt.Println(strings.Join(row, "\t"))
	}
}

func DetectFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".csv":
		return "csv"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".xlsx", ".xls":
		return "excel"
	default:
		return "unknown"
	}
}

func ReadFile(filename string) (*Dataset, error) {
	format := DetectFormat(filename)

	switch format {
	case "csv":
		return ReadCSV(filename)
	case "json":
		return ReadJSON(filename)
	case "xml":
		return ReadXML(filename)
	case "excel":
		return ReadExcel(filename)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", format)
	}
}

func ReadCSV(filename string) (*Dataset, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return &Dataset{Headers: []string{}, Rows: [][]string{}}, nil
	}

	return &Dataset{
		Headers: records[0],
		Rows:    records[1:],
	}, nil
}

func ReadJSON(filename string) (*Dataset, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var data []map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(data) == 0 {
		return &Dataset{Headers: []string{}, Rows: [][]string{}}, nil
	}

	headers := getJSONHeaders(data)
	rows := make([][]string, len(data))
	for i, item := range data {
		rows[i] = make([]string, len(headers))
		for j, h := range headers {
			if val, ok := item[h]; ok {
				rows[i][j] = fmt.Sprintf("%v", val)
			}
		}
	}

	return &Dataset{Headers: headers, Rows: rows}, nil
}

func getJSONHeaders(data []map[string]interface{}) []string {
	headerSet := make(map[string]bool)
	for _, item := range data {
		for k := range item {
			headerSet[k] = true
		}
	}
	headers := make([]string, 0, len(headerSet))
	for k := range headerSet {
		headers = append(headers, k)
	}
	return headers
}

type xmlRecord struct {
	XMLName xml.Name   `xml:"record"`
	Fields  []xmlField `xml:",any"`
}

type xmlField struct {
	XMLName xml.Name
	Value   string `xml:",innerxml"`
}

func ReadXML(filename string) (*Dataset, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var result struct {
		Records []xmlRecord `xml:"record"`
	}

	if err := xml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	if len(result.Records) == 0 {
		return &Dataset{Headers: []string{}, Rows: [][]string{}}, nil
	}

	headers := extractXMLHeaders(result.Records)
	rows := make([][]string, len(result.Records))
	for i, rec := range result.Records {
		rows[i] = make([]string, len(headers))
		for _, f := range rec.Fields {
			for j, h := range headers {
				if strings.EqualFold(h, f.XMLName.Local) {
					rows[i][j] = f.Value
					break
				}
			}
		}
	}

	return &Dataset{Headers: headers, Rows: rows}, nil
}

func extractXMLHeaders(records []xmlRecord) []string {
	headerSet := make(map[string]bool)
	for _, rec := range records {
		for _, f := range rec.Fields {
			headerSet[f.XMLName.Local] = true
		}
	}
	headers := make([]string, 0, len(headerSet))
	for k := range headerSet {
		headers = append(headers, k)
	}
	return headers
}

func ReadExcel(filename string) (*Dataset, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return &Dataset{Headers: []string{}, Rows: [][]string{}}, nil
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel sheet: %w", err)
	}

	if len(rows) == 0 {
		return &Dataset{Headers: []string{}, Rows: [][]string{}}, nil
	}

	return &Dataset{
		Headers: rows[0],
		Rows:    rows[1:],
	}, nil
}

func WriteFile(ds *Dataset, filename string, format string) error {
	if format == "" {
		format = DetectFormat(filename)
	}

	switch format {
	case "csv":
		return WriteCSV(ds, filename)
	case "json":
		return WriteJSON(ds, filename)
	case "xml":
		return WriteXML(ds, filename)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func WriteCSV(ds *Dataset, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(ds.Headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, row := range ds.Rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

func WriteJSON(ds *Dataset, filename string) error {
	data := make([]map[string]interface{}, len(ds.Rows))
	for i, row := range ds.Rows {
		record := make(map[string]interface{})
		for j, h := range ds.Headers {
			if j < len(row) {
				record[h] = row[j]
			}
		}
		data[i] = record
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func WriteXML(ds *Dataset, filename string) error {
	type XMLRecord struct {
		XMLName xml.Name          `xml:"record"`
		Fields  map[string]string `xml:"-"`
	}

	type XMLData struct {
		XMLName xml.Name    `xml:"data"`
		Records []XMLRecord `xml:"record"`
	}

	records := make([]XMLRecord, len(ds.Rows))
	for i, row := range ds.Rows {
		fields := make(map[string]string)
		for j, h := range ds.Headers {
			if j < len(row) {
				fields[h] = row[j]
			}
		}
		records[i] = XMLRecord{Fields: fields}
	}

	data := XMLData{Records: records}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	return encoder.Encode(data)
}
