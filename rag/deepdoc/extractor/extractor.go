package extractor

import (
	"regexp"
)

var reDocx = regexp.MustCompile(`\.docx$`)
var reXlsx = regexp.MustCompile(`\.xlsx$`)
var rePptx = regexp.MustCompile(`\.pptx$`)
var rePdf = regexp.MustCompile(`\.pdf$`)
var reTxtx = regexp.MustCompile(`\.(txt|py|js|java|c|cpp|h|php|go|ts|sh|cs|kt)$`)
var reDoc = regexp.MustCompile(`\.doc$`)
var reCsv = regexp.MustCompile(`\.csv$`)
var reJson = regexp.MustCompile(`\.json$`)
var reMdx = regexp.MustCompile(`\.md$`)

var extractors = map[string]Extractor{}

func init() {
	extractors["doc"] = NewDocumentExtractor()
	extractors["faq"] = NewFaqExtractor()
	extractors["unknown"] = NewUnknownExtractor()
}

// Extractor ...
type Extractor interface {
	Extract(*Meta, []byte) ([]string, error)
}

// Meta ...
type Meta struct {
	Path     string
	Filename string
}

// Extract ...
func Extract(meta *Meta, content []byte, typ string) ([]string, error) {
	return extractors[typ].Extract(meta, content)
}
