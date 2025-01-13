package deepdoc

import (
	"os"
	"os/exec"

	"baliance.com/gooxml/document"
	"github.com/google/uuid"
)

// DocDocument ...
type DocDocument struct {
}

// Doc ...
func Doc() *DocDocument {
	return &DocDocument{}
}

// Extract ...
func (d *DocDocument) Extract(filename string, binary []byte) ([]string, error) {
	// save tmp file
	f, err := os.CreateTemp("/tmp", uuid.NewString())
	// defer os.Remove(f.Name())
	_, err = f.Write(binary)
	// safe 校验
	cmd := exec.Command("soffice", "--headless", "--convert-to", "docx", f.Name(), "--outdir", "/tmp")
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// doc, err := document.Read(bytes.NewReader(binary), int64(len(binary)))
	doc, err := document.Open(f.Name() + ".docx")
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
