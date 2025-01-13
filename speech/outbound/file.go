package outbound

import (
	"fmt"
	"os"

	"github.com/showntop/llmack/speech"
)

// FileOutBound for test ...
type FileOutBound struct {
	ff *os.File
}

// NewFileOutBound ...
func NewFileOutBound() speech.Outbound {
	f, _ := os.Create("out.wav")
	return &FileOutBound{ff: f}

}

// Write ...
func (o *FileOutBound) Write(msg []byte) error {
	fmt.Println("file outbound write", len(msg))
	_, err := o.ff.Write(msg)
	return err
}

// Reset ...
func (o *FileOutBound) Reset() error {
	return nil
}

// Close ...
func (o *FileOutBound) Close() error {
	return o.ff.Close()
}
