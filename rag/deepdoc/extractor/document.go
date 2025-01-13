package extractor

import (
	"strings"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/rag/deepdoc"
)

// DocumentExtractor ...
type DocumentExtractor struct {
}

// NewDocumentExtractor ...
func NewDocumentExtractor() *DocumentExtractor {
	return &DocumentExtractor{}
}

// Extract ...
func (e *DocumentExtractor) Extract(m *Meta, binary []byte) ([]string, error) {
	filename := m.Filename
	var sections []string
	var err error
	if v := reDocx.FindStringIndex(filename); v != nil { // docx
		sections, err = deepdoc.Docx().Extract(filename, binary)
	} else if v := reXlsx.FindStringIndex(filename); v != nil {
		sections, err = deepdoc.Excel().Extract(filename, binary)
	} else if v := rePptx.FindStringIndex(filename); v != nil {
		sections, err = deepdoc.Pptx().Extract(filename, binary)
	} else if v := rePdf.FindStringIndex(filename); v != nil {
		sections, err = deepdoc.Pdf().Extract(filename, binary)
	} else if v := reTxtx.FindStringIndex(filename); v != nil {
		sections, err = deepdoc.Txtx().Extract(filename, binary)
	} else if v := reDoc.FindStringIndex(filename); v != nil {
		sections, err = deepdoc.Doc().Extract(filename, binary)
	} else if v := reCsv.FindStringIndex(filename); v != nil {
		sections, err = deepdoc.Csv().Extract(filename, binary)
	} else if v := reJson.FindStringIndex(filename); v != nil {
		sections, err = deepdoc.Json().Extract(filename, binary)
	} else if v := reMdx.FindStringIndex(filename); v != nil {
		sections, err = deepdoc.Mdx().Extract(filename, binary)
	} else {
		log.WarnContextf(nil, "extract unknow document filename: %s by txt", filename)
		sections, err = deepdoc.Txtx().Extract(filename, binary)
		// return nil, fmt.Errorf("unsupported file format: %s", filename)
	}
	if err != nil {
		// 兜底 解析 按照文本
		log.WarnContextf(nil, "extract document error: %v, filename: %s", err, filename)
		sections, err = deepdoc.Txtx().Extract(filename, binary)
		return nil, err
	}

	return []string{strings.Join(sections, "")}, nil
}
