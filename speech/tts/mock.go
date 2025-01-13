package tts

import (
	"fmt"
	"os"

	"github.com/showntop/llmack/speech"
)

// MockTTS ...
type MockTTS struct {
	q       chan []byte
	calback func([]byte) error
}

// NewMockTTS ...
func NewMockTTS(c func([]byte) error) speech.StreamTTS {
	t := &MockTTS{calback: c, q: make(chan []byte, 10)}
	go func() {
		audio, err := os.Open("./tts.wav")
		if err != nil {
			panic(err)
		}
		defer audio.Close()
		x := 0
		for {
			data := make([]byte, 8044)
			n, err := audio.Read(data)
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				fmt.Printf("read file error: %v\n", err)
				break
			}
			// fmt.Println("out seq init ,", base64.StdEncoding.EncodeToString(data))

			if n <= 0 {
				break
			}
			t.q <- data //[:n]

			// time.Sleep(500 * time.Millisecond)
			x++
			//模拟真实场景，200ms产生200ms数据
			//time.Sleep(200 * time.Millisecond)
		}
	}()
	return t
}

// Synthesize ...
func (c *MockTTS) Synthesize(_ string) (chan []byte, error) {
	return nil, nil
}

// Input ...
func (c *MockTTS) Input(_ string) error {
	for i := 0; i < 10; i++ {
		x := <-c.q
		c.calback(x)
	}
	return nil
}

// Complete ...
func (c *MockTTS) Complete() error {
	return nil
}

// Start ...
func (c *MockTTS) Start() error {
	return nil
}

// Prepare ...
func (c *MockTTS) Prepare() error {
	return nil
}

// Terminate ...
func (c *MockTTS) Terminate() error {
	return nil
}

// StreamResult ...
func (c *MockTTS) StreamResult() chan []byte {
	return nil
}

// Close ...
func (c *MockTTS) Close() error {
	return nil
}

// package tts

// import (
// 	"fmt"
// 	"os"

// 	"github.com/showntop/llmack/speech"
// )

// // MockTTS ...
// type MockTTS struct {
// 	// q       chan []byte
// 	buffer  [][]byte
// 	calback func([]byte) error
// }

// // NewMockTTS ...
// func NewMockTTS(c func([]byte) error) speech.TTS {
// 	t := &MockTTS{calback: c}
// 	t.buffer = make([][]byte, 1000, 1000)

// 	audio, err := os.Open("./tts.wav")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer audio.Close()
// 	x := 0
// 	for {
// 		data := make([]byte, 8044)
// 		n, err := audio.Read(data)
// 		if err != nil {
// 			if err.Error() == "EOF" {
// 				break
// 			}
// 			fmt.Printf("read file error: %v\n", err)
// 			break
// 		}

// 		if n <= 0 {
// 			break
// 		}
// 		t.buffer[x] = data
// 		x++
// 	}
// 	return t
// }

// // Synthesize ...
// func (c *MockTTS) Synthesize(_ string) ([]byte, error) {
// 	return nil, nil
// }

// // Input ...
// func (c *MockTTS) Input(_ string) error {
// 	fmt.Printf("tts mock input data: %v\n", "len(x)")
// 	for i := 0; i < len(c.buffer); i++ {
// 		x := c.buffer[i]
// 		fmt.Printf("tts mock input data x: %v\n", len(x))
// 		c.calback(x)
// 	}
// 	return nil
// }

// // Start ...
// func (c *MockTTS) Start() error {
// 	return nil
// }

// // Close ...
// func (c *MockTTS) Close() {

// }
