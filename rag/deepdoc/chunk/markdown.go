package chunk

import (
	"fmt"
	"strings"
)

// LineType 用于表示带有元数据的行类型
type LineType struct {
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

// HeaderType 用于表示标题类型
type HeaderType struct {
	Level int    `json:"-"`
	Name  string `json:"-"`
	Data  string `json:"-"`
}

// Document 用于表示文档结构
type Document struct {
	PageContent string            `json:"page_content"`
	Metadata    map[string]string `json:"metadata"`
}

// MarkdownHeaderChunker ...
type MarkdownHeaderChunker struct {
	headersToSplitOn [][2]string
}

// NewMarkdownHeaderChunker ...
func NewMarkdownHeaderChunker(htso [][2]string) *MarkdownHeaderChunker {
	return &MarkdownHeaderChunker{
		headersToSplitOn: htso,
	}
}

// Chunk ...
func (c *MarkdownHeaderChunker) Chunk(content string) ([]string, error) {
	documents := c.chunk(content, true)
	for i := 0; i < len(documents); i++ {
		fmt.Println(documents[i])
	}
	return nil, nil
}

// chunk 函数用于将Markdown文件分割成块
func (c *MarkdownHeaderChunker) chunk(text string, returnEachLine bool) []Document {
	// 按换行符("\n")分割输入文本
	lines := strings.Split(text, "\n")
	// 最终输出
	linesWithMetadata := []LineType{}
	// 当前正在处理的块的内容和元数据
	currentContent := []string{}
	currentMetadata := map[string]string{}
	// 跟踪嵌套的标题结构
	headerStack := []HeaderType{}
	initialMetadata := map[string]string{}

	for _, line := range lines {
		strippedLine := strings.TrimSpace(line)
		// 检查每一行是否以我们打算分割的标题开头（例如，#、##）
		for _, header := range c.headersToSplitOn {
			sep, name := header[0], header[1]
			if strings.HasPrefix(strippedLine, sep) &&
				(len(strippedLine) == len(sep) || strippedLine[len(sep)] == ' ') {
				// 确保我们将标题作为元数据跟踪
				if name != "" {
					// 获取当前标题级别
					currentHeaderLevel := strings.Count(sep, "#")

					// 从堆栈中弹出较低级别或同等级别的标题
					for len(headerStack) > 0 && headerStack[len(headerStack)-1].Level >= currentHeaderLevel {
						// 我们遇到了一个相同或更高级别的新标题
						poppedHeader := headerStack[len(headerStack)-1]
						headerStack = headerStack[:len(headerStack)-1]
						// 清除初始元数据中弹出的标题的元数据
						if _, ok := initialMetadata[poppedHeader.Name]; ok {
							delete(initialMetadata, poppedHeader.Name)
						}
					}

					// 将当前标题推入堆栈
					header := HeaderType{
						Level: currentHeaderLevel,
						Name:  name,
						Data:  strings.TrimSpace(strippedLine[len(sep):]),
					}
					headerStack = append(headerStack, header)
					// 用当前标题更新初始元数据
					initialMetadata[name] = header.Data
				}

				// 只有当 currentContent 不为空时，才将前一行添加到 linesWithMetadata
				if len(currentContent) > 0 {
					linesWithMetadata = append(linesWithMetadata, LineType{
						Content:  strings.Join(currentContent, "\n"),
						Metadata: currentMetadata,
					})
					currentContent = []string{}
				}

				break
			}
		}
		if !strings.HasPrefix(strippedLine, "#") {
			if strippedLine != "" {
				currentContent = append(currentContent, strippedLine)
			} else if len(currentContent) > 0 {
				linesWithMetadata = append(linesWithMetadata, LineType{
					Content:  strings.Join(currentContent, "\n"),
					Metadata: currentMetadata,
				})
				currentContent = []string{}
			}
		}

		currentMetadata = make(map[string]string)
		for k, v := range initialMetadata {
			currentMetadata[k] = v
		}
	}

	if len(currentContent) > 0 {
		linesWithMetadata = append(linesWithMetadata, LineType{
			Content:  strings.Join(currentContent, "\n"),
			Metadata: currentMetadata,
		})
	}

	// linesWithMetadata 中的每一行都有相关的标题元数据
	// 根据共同的元数据将这些行聚合成块
	if !returnEachLine {
		// TODO: 实现将行聚合成块的逻辑
		return c.aggregateLinesToChunks(linesWithMetadata)
	} else {
		// TODO: 实现返回Document结构的逻辑
		documents := []Document{}
		for _, m := range linesWithMetadata {
			documents = append(documents, Document{
				PageContent: m.Content,
				Metadata:    m.Metadata,
			})
		}
		return documents
	}
}

// aggregateLinesToChunks 函数用于将具有共同元数据的行聚合成块
func (c *MarkdownHeaderChunker) aggregateLinesToChunks(lines []LineType) []Document {
	aggregatedChunks := []LineType{}

	for _, line := range lines {
		// if len(aggregatedChunks) > 0 && aggregatedChunks[len(aggregatedChunks)-1].Metadata == line.Metadata {
		if len(aggregatedChunks) > 0 {
			// 如果聚合列表中的最后一行与当前行具有相同的元数据，
			// 则将当前内容追加到最后一行的内容中
			aggregatedChunks[len(aggregatedChunks)-1].Content += "  \n" + line.Content
		} else {
			// 否则，将当前行追加到聚合列表中
			aggregatedChunks = append(aggregatedChunks, line)
		}
	}

	// 将聚合后的块转换为Document结构体列表
	var documents []Document
	for _, chunk := range aggregatedChunks {
		documents = append(documents, Document{
			PageContent: chunk.Content,
			Metadata:    chunk.Metadata,
		})
	}

	return documents
}
