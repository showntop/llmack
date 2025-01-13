package deepdoc

import (
	"bytes"
	"fmt"
	"log"

	"github.com/showntop/llmack/pkg/strings"

	"github.com/dslipak/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"

	"github.com/showntop/unipdf/extractor"
	"github.com/showntop/unipdf/model"
)

// PdfDocument ...
type PdfDocument struct {
}

// Pdf ...
func Pdf() *PdfDocument {
	return &PdfDocument{}
}

// Extract ...
func (d *PdfDocument) Extract(filename string, binary []byte) ([]string, error) {
	sections := make([]string, 0)

	reader, err := model.NewPdfReader(bytes.NewReader(binary))
	if err != nil {
		return nil, err
	}
	num, _ := reader.GetNumPages()
	for i := 1; i <= num; i++ {
		page, _ := reader.GetPage(i)
		ext, _ := extractor.New(page)
		text, _ := ext.ExtractText()
		sections = append(sections, text)
	}

	return sections, nil
}

// Extract ...
func (d *PdfDocument) Extract3(filename string, binary []byte) ([]string, error) {
	sections := make([]string, 0)
	ctx, err := api.ReadContextFile("./pkg/deepdoc/example/销售认证-销售代表-学习指引.pdf")
	// ctx, err := api.ReadContext(bytes.NewReader(binary), model.NewDefaultConfiguration())
	// pdf2.ExtractContentFile("in.pdf", "outDir", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	page, _ := api.ExtractPage(ctx, 1)
	var p []byte = make([]byte, 100000)
	fmt.Println(page.Read(p))

	fmt.Println(string(p))

	api.ExtractContentFile("./pkg/deepdoc/example/销售认证-销售代表-学习指引.pdf", "./", []string{"1"}, nil)
	// fmt.Println()
	// api.ExtractContent()
	// p, _ := ctx.Pages()
	// _ = p
	// fmt.Println(p.PDFString())
	return sections, nil
}

func (d *PdfDocument) Extract2(filename string, binary []byte) ([]string, error) {
	sections := make([]string, 0)
	reader, err := pdf.NewReader(bytes.NewReader(binary), int64(len(binary)))
	if err != nil {
		return nil, err
	}

	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		for _, txt := range page.Content().Text {
			// fmt.Println(txt.Font, txt.S)
			_ = txt
		}
		// content, _ := reader.Page(i).GetPlainText(nil)
		content, _ := page.GetPlainText(nil)
		sections = append(sections, strings.TrimSpecial(content))
	}
	return sections, nil
}
