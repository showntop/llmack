package asr

import (
	"strings"
)

// MockASR MockASR
type MockASR struct {
	current  int
	contents []string
	calback  func(string, bool) error
}

// NewMockASR 新建一个MockASR
func NewMockASR(c func(string, bool) error) *MockASR {
	t := &MockASR{calback: c}
	// t.contents = []string{"你好，", "请问怎么", "处理广告", "不起量的", "问题。"}
	t.contents = []string{"你是谁。"}
	return t
}

// Input 实现语音识别，转录为文字
func (t *MockASR) Input(content []byte) error {
	if len(content) == 0 {
		return nil
	}
	if t.current >= len(t.contents) {
		return nil
	}

	ssss := t.contents[t.current]
	final := false
	if t.current == len(t.contents)-1 {
		ssss = strings.Join(t.contents, "")
		final = true
		// t.current = 0
	}

	t.calback(ssss, final)
	t.current++
	return nil
}

// Recognize 实现语音识别，转录为文字
func (t *MockASR) Recognize(content []byte) (string, error) {
	return "", nil
}
