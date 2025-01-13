package chunk

import (
	"regexp"
	"strings"
)

// Chunker is the standard interface for splitting texts.
type Chunker interface {
	Chunk(string) ([]string, error)
}

// Base is the base struct for all chunkers.
type Base struct {
	chunkSize   int
	overlapSize int

	len func(s string) int
}

// newBase returns a new Base chunker.
func newBase(chunkSize, overlapSize int) *Base {
	if chunkSize < overlapSize {
		panic("chunk size must be greater than overlap size")
	}
	return &Base{
		chunkSize:   chunkSize,
		overlapSize: overlapSize,
		len:         lenRune,
	}
}

// lenRune ...
func lenRune(s string) int {
	return len([]rune(s))
}

// mergeSplits combine smaller pieces into medium size
func (b *Base) mergeSplits(splits []string, separator string) []string {
	separatorLen := b.len(separator)

	var docs []string
	var combinedSplit []string
	total := 0
	for _, s := range splits {
		sLen := b.len(s)
		// 检查添加片段后总长度是否会超过预设的 chunkSize
		if total+sLen+separatorLen > b.chunkSize {
			if total > b.chunkSize {
				// 如果当前文档总长度超过了 chunkSize，记录警告信息
				println("Warning: Created a chunk of size", total, "which is longer than the specified", b.chunkSize)
			}
			if len(combinedSplit) > 0 {
				// 将当前文档中的片段用 separator 连接起来，形成一个新的中等大小文本块
				doc := strings.Join(combinedSplit, separator)
				docs = append(docs, doc)

				// 如果当前块的总长度超过了 overlapSize 或者加上新片段后总长度仍然超过 chunkSize 且 total 大于 0
				// 则从当前文档的开头移除片段，直到满足条件
				for (total > b.overlapSize || (total+sLen+separatorLen > b.chunkSize && total > 0)) && len(combinedSplit) > 0 {
					removesLen := b.len(combinedSplit[0]) + separatorLen
					if len(combinedSplit) > 1 {
						removesLen = b.len(combinedSplit[0])
					}
					total -= removesLen
					combinedSplit = combinedSplit[1:]
				}
			}

		}
		// 将当前片段 d 添加到 combinedSplit 列表中
		combinedSplit = append(combinedSplit, s)
		// 更新 total 的值
		if len(combinedSplit) > 1 {
			total += sLen + separatorLen
		} else {
			total += sLen
		}
	}
	// 将 combinedSplit 中的片段连接成一个中等大小的文本块
	doc := strings.Join(combinedSplit, separator)
	docs = append(docs, doc)
	return docs
}

func splitTextWithRegex(text string, separator string, reserveSeparator bool) []string {
	// if separator == "" {
	// 	return []string{text}
	// }
	// if reserveSeparator

	regex := regexp.MustCompile(separator)
	splits := regex.Split(text, -1)

	return splits
}

// Chunk 将字符串按照指定的字符数切分成多个分片
func Chunk(content string, chunkSize int) []string {
	var chunks []string
	total := len([]rune(content))
	num := total / chunkSize
	if total%chunkSize != 0 {
		num++
	}

	for i := 0; i < num; i++ {
		from := i * chunkSize
		to := i*chunkSize + chunkSize
		if to > total {
			to = total
		}
		chunks = append(chunks, string([]rune(content)[from:to]))
	}

	return chunks
}
