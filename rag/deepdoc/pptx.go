package deepdoc

import (
	"bytes"

	"baliance.com/gooxml/presentation"
)

// PptxDocument ...
type PptxDocument struct {
}

// Pptx ...
func Pptx() *PptxDocument {
	return &PptxDocument{}
}

// Extract ...
func (d *PptxDocument) Extract(filename string, binary []byte) ([]string, error) {
	doc, err := presentation.Read(bytes.NewReader(binary), int64(len(binary)))
	if err != nil {
		return nil, err
	}

	sections := make([]string, 0)

	for _, slide := range doc.Slides() {
		//run为每个段落相同格式的文字组成的片段
		for _, choice := range slide.X().CSld.SpTree.Choice {
			if choice.Sp == nil {
				continue
			}
			for _, sp := range choice.Sp {
				for _, p := range sp.TxBody.P {
					for _, run := range p.EG_TextRun {
						if run.R == nil {
							continue
						}
						sections = append(sections, run.R.T)
					}
				}
			}
		}
	}

	return sections, nil
}
