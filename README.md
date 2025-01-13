

# GoLang Fullstack LLM Framework

## Overview

The GoLang Fullstack LLM Framework is a comprehensive toolkit designed to facilitate the development of applications leveraging Large Language Models (LLMs). This framework integrates various components such as LLM interaction, prompt management, Retrieval-Augmented Generation (RAG), speech processing, programmatic control, optimization, and an engine to orchestrate these components seamlessly.

## Features

- **LLM Integration**: Interact with various Large Language Models (e.g., GPT, Claude) for text generation, summarization, and more.
- **Prompt Management**: Easily manage and optimize prompts for different tasks and models.
- **Retrieval-Augmented Generation (RAG)**: Combine retrieval-based methods with generative models to enhance response quality.
- **Speech Processing**: Convert speech to text and vice versa, enabling voice-based interactions.
- **Programmatic Control**: Define and execute complex workflows programmatically.
- **Optimizer**: Fine-tune model parameters and prompts for better performance.
- **Engine**: A core engine to manage and orchestrate all components efficiently.

## Usage

To use the GoLang Fullstack LLM Framework, ensure you have Go installed on your system. Then, run the following command:

```bash
go get github.com/showntop/llmack
```

### LLM Integration

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/zhipu"
)

func main() {
	ctx := context.Background()
	llm.WithSingleConfig(map[string]any{
		"api_key": os.Getenv("zhipu_api_key"),
	})

	resp, err := llm.NewInstance(zhipu.Name).Invoke(ctx, []llm.Message{
		llm.UserPromptMessage("hello"),
	}, []llm.PromptMessageTool{}, llm.WithModel("GLM-4-Flash"))
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Result())
}
```

More examples can be found in the [example](example) directory.

## Contributing

We welcome contributions! Please see our [CONTRIBUTING.md](CONTRIBUTING.md) for more details on how to get started.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to the Go community for their excellent tools and libraries.
- Special thanks to the developers of the various LLMs and related technologies that make this framework possible.

## Contact

For any questions or feedback, please reach out to us at [rongtaoxiao@gmail.com](mailto:rongtaoxiao@gmail.com).

---

This README provides a basic overview of the GoLang Fullstack LLM Framework. For more detailed documentation, please refer to the individual package documentation within the repository.