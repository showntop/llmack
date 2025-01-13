package chunk

import (
	"regexp"
)

// CharacterChunker is a chunker that splits text into chunks of characters.
type CharacterChunker struct {
	separator string
	*Base
}

// NewCharacterChunker ...
func NewCharacterChunker(chunkSize int, overlapSize int, seq string) *CharacterChunker {
	return &CharacterChunker{
		separator: seq,
		Base:      newBase(chunkSize, overlapSize),
	}
}

// Chunk splits text into chunks of characters.
func (c *CharacterChunker) Chunk(content string) ([]string, error) {
	splits := splitTextWithRegex(content, c.separator, false)
	// fmt.Println(strings.Join(splits, ",,,,"))
	return c.mergeSplits(splits, c.separator), nil
}

// RecursiveCharacterChunker is a chunker that splits text into chunks of characters by recursively look at characters.
type RecursiveCharacterChunker struct {
	separators []string
	*Base
}

// NewRecursiveCharacterChunker ...
func NewRecursiveCharacterChunker(chunkSize int, overlapSize int, seqs []string) *RecursiveCharacterChunker {
	return &RecursiveCharacterChunker{
		separators: seqs,
		Base:       newBase(chunkSize, overlapSize),
	}
}

// Chunk splits text into chunks of characters.
func (c *RecursiveCharacterChunker) Chunk(content string) ([]string, error) {
	return c.chunk(content, c.separators, c.chunkSize, false), nil
}

// chunk 函数用于将文本分割成块
func (c *RecursiveCharacterChunker) chunk(text string, separators []string, chunkSize int, keepSeparator bool) []string {
	finalChunks := []string{}
	// 获取要使用的适当分隔符
	separator := separators[len(separators)-1]
	newSeparators := []string{}
	for i, s := range separators {
		if s == "" {
			separator = s
			break
		}
		re := regexp.MustCompile(s)
		if re.MatchString(text) {
			separator = s
			newSeparators = separators[i+1:]
			break
		}
	}
	splits := splitTextWithRegex(text, separator, keepSeparator)
	// 现在开始合并片段，递归地分割较长的文本
	var goodSplits []string
	var separatorToUse string
	// if keepSeparator {
	// separatorToUse = ""
	// } else {
	separatorToUse = separator
	// }
	for _, s := range splits {
		if c.len(s) < chunkSize {
			goodSplits = append(goodSplits, s)
		} else {
			if len(goodSplits) > 0 {
				mergedText := c.mergeSplits(goodSplits, separatorToUse)
				finalChunks = append(finalChunks, mergedText...)
				goodSplits = []string{}
			}
			if len(newSeparators) == 0 {
				finalChunks = append(finalChunks, s)
			} else {
				otherInfo := c.chunk(s, newSeparators, chunkSize, keepSeparator)
				finalChunks = append(finalChunks, otherInfo...)
			}
		}
	}
	if len(goodSplits) > 0 {
		mergedText := c.mergeSplits(goodSplits, separatorToUse)
		finalChunks = append(finalChunks, mergedText...)
	}
	return finalChunks
}
