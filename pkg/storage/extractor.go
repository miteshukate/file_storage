package storage

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
)

// TextExtractor defines the interface for extracting text from various file types
type TextExtractor interface {
	Extract(reader io.Reader) (string, error)
}

// ExtractTextFromFile extracts text content from a file based on its MIME type
func ExtractTextFromFile(reader io.Reader, mimeType string) (string, error) {
	// Read all content into buffer for reusability
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Determine file type and extract accordingly
	switch {
	case strings.HasPrefix(mimeType, "text/plain"):
		return extractPlainText(buf)
	case strings.HasPrefix(mimeType, "application/pdf"):
		return extractPDFText(buf)
	case strings.HasPrefix(mimeType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"):
		return extractExcelText(buf)
	case strings.HasPrefix(mimeType, "application/vnd.ms-excel"):
		return extractExcelText(buf)
	default:
		// For unsupported types, return filename extraction hint
		return "", nil
	}
}

// extractPowerPointText extracts text from PowerPoint files
func extractPowerPointText(buf *bytes.Buffer) (string, error) {
	// extract text from PowerPoint files using a suitable library (e.g., github.com/unidoc/unioffice)
	return buf.String(), nil
}

// extractPlainText extracts text from plain text files
func extractPlainText(buf *bytes.Buffer) (string, error) {
	return buf.String(), nil
}

// extractPDFText extracts text from PDF files
func extractPDFText(buf *bytes.Buffer) (string, error) {
	// Convert buffer to bytes
	data := buf.Bytes()

	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to read PDF: %w", err)
	}

	var text strings.Builder
	pageCount := reader.NumPage()

	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		page := reader.Page(pageNum)

		content, err := page.GetPlainText(nil)
		if err != nil {
			// Log error but continue with other pages
			fmt.Printf("failed to extract text from PDF page %d: %v\n", pageNum, err)
			continue
		}

		if content != "" {
			text.WriteString(content)
			text.WriteString("\n")
		}
	}

	result := strings.TrimSpace(text.String())
	if result == "" {
		return "", fmt.Errorf("no text extracted from PDF")
	}

	return result, nil
}

// extractExcelText extracts text from Excel files
func extractExcelText(buf *bytes.Buffer) (string, error) {
	file, err := excelize.OpenReader(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read Excel file: %w", err)
	}
	defer file.Close()

	var text strings.Builder

	// Iterate through all sheets
	sheets := file.GetSheetList()
	for _, sheetName := range sheets {
		rows, err := file.GetRows(sheetName)
		if err != nil {
			fmt.Printf("failed to get rows from sheet %s: %v\n", sheetName, err)
			continue
		}

		// Add sheet name to text
		text.WriteString(fmt.Sprintf("Sheet: %s\n", sheetName))

		// Iterate through rows
		for _, row := range rows {
			for i, cell := range row {
				text.WriteString(cell)
				if i < len(row)-1 {
					text.WriteString(" | ")
				}
			}
			text.WriteString("\n")
		}

		text.WriteString("\n")
	}

	result := strings.TrimSpace(text.String())
	if result == "" {
		return "", fmt.Errorf("no text extracted from Excel file")
	}

	return result, nil
}

// ExtractSummary creates a concise summary of the extracted text
func ExtractSummary(fullText string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = 500
	}

	if len(fullText) <= maxLength {
		return fullText
	}

	return fullText[:maxLength] + "..."
}
