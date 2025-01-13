package deepdoc

import (
	"bytes"

	"baliance.com/gooxml/document"
)

// DocxDocument ...
type DocxDocument struct {
}

// Docx ...
func Docx() *DocxDocument {
	return &DocxDocument{}
}

// Extract ...
func (d *DocxDocument) Extract(filename string, binary []byte) ([]string, error) {
	doc, err := document.Read(bytes.NewReader(binary), int64(len(binary)))
	if err != nil {
		return nil, err
	}

	sections := []string{}
	for _, para := range doc.Paragraphs() {
		//run为每个段落相同格式的文字组成的片段
		for _, run := range para.Runs() {
			sections = append(sections, run.Text())
		}
	}
	return sections, nil
}
