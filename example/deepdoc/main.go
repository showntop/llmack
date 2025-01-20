package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/showntop/llmack/rag/deepdoc/chunk"
	"github.com/showntop/llmack/rag/deepdoc/extractor"

	pstrings "github.com/showntop/llmack/pkg/strings"
)

func main() {

	file, err := os.OpenFile(
		"example/deepdoc/企业级 SaaS 行业增长白皮书.pdf",
		os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	content, _ := io.ReadAll(file)
	sections, err := extractor.Extract(
		&extractor.Meta{Filename: "企业级 SaaS 行业增长白皮书.pdf"}, content, "doc")
	if err != nil {
		panic(err)
	}

	chunker := chunk.NewCharacterChunker(500, 50, "")
	chunks, _ := chunker.Chunk(string(strings.Join(sections, "\n")))
	fmt.Println(len(chunks))
	for i := 0; i < len(chunks); i++ {
		fmt.Println("--------------------")
		fmt.Printf(pstrings.TrimSpecial(chunks[i]))
		fmt.Println(len([]rune(pstrings.TrimSpecial(chunks[i]))))
	}
}
