package file

import (
	"context"

	"github.com/showntop/llmack/tool"
)

const ReadFile = "ReadFile"

func init() {
	t := &tool.Tool{}
	t.Name = ReadFile
	t.Kind = "code"
	t.Description = "Reads the file content in a specified location."
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name:          "file_name",
		LLMDescrition: "Path of the file to read.",
		Type:          tool.String,
		Required:      true,
	}, tool.Parameter{
		Name:          "content",
		LLMDescrition: "File content to write.",
		Type:          tool.String,
		Required:      true,
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		// fileName, _ := args["file_name"].(string)
		// // 获取最终文件路径
		// finalPath, err := ResourceHelper.GetAgentReadResourcePath(fileName,
		// 	Agent.GetAgentFromID(t.ToolkitConfig.Session, t.AgentID),
		// 	AgentExecution.GetAgentExecutionFromID(t.ToolkitConfig.Session, t.AgentExecutionID))
		// if err != nil {
		// 	return "", err
		// }

		// var temporaryFilePath string
		// finalName := filepath.Base(finalPath)

		// // 处理 S3 存储
		// // if GetStorageType(GetConfig("STORAGE_TYPE", StorageType.FILE.Value)) == StorageType.S3 {
		// // 	if strings.HasSuffix(strings.ToLower(finalName), ".txt") {
		// // 		return S3Helper{}.ReadFromS3(finalPath)
		// // 	} else {
		// // 		saveDirectory := "/"
		// // 		temporaryFilePath = filepath.Join(saveDirectory, fileName)
		// // 		f, err := os.Create(temporaryFilePath)
		// // 		if err != nil {
		// // 			return "", err
		// // 		}
		// // 		defer f.Close()

		// // 		contents, err := S3Helper{}.ReadBinaryFromS3(finalPath)
		// // 		if err != nil {
		// // 			return "", err
		// // 		}
		// // 		if _, err := f.Write(contents); err != nil {
		// // 			return "", err
		// // 		}
		// // 	}
		// // }

		// // 检查文件是否存在
		// if finalPath == "" || (!fileExists(finalPath) && temporaryFilePath == "") {
		// 	return "", fmt.Errorf("File '%s' not found", fileName)
		// }

		// // 创建目录
		// directory := filepath.Dir(finalPath)
		// if err := os.MkdirAll(directory, 0755); err != nil {
		// 	return "", err
		// }

		// if temporaryFilePath != "" {
		// 	finalPath = temporaryFilePath
		// }

		// var content string
		// // 处理 epub 文件
		// if strings.HasSuffix(strings.ToLower(finalPath), ".epub") {
		// 	book, err := epub.Open(finalPath)
		// 	if err != nil {
		// 		return "", err
		// 	}

		// 	var contents []string
		// 	for _, item := range book.Sections() {
		// 		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(item.Content)))
		// 		if err != nil {
		// 			continue
		// 		}
		// 		contents = append(contents, doc.Text())
		// 	}
		// 	content = strings.Join(contents, "\n")
		// } else {
		// 	// 处理其他类型文件
		// 	if strings.HasSuffix(finalPath, ".csv") {
		// 		if err := CorrectCSVEncoding(finalPath); err != nil {
		// 			return "", err
		// 		}
		// 	}

		// 	elements, err := Partition(finalPath)
		// 	if err != nil {
		// 		return "", err
		// 	}

		// 	var elementStrings []string
		// 	for _, el := range elements {
		// 		elementStrings = append(elementStrings, fmt.Sprintf("%v", el))
		// 	}
		// 	content = strings.Join(elementStrings, "\n\n")
		// }

		// // 清理临时文件
		// if temporaryFilePath != "" {
		// 	os.Remove(temporaryFilePath)
		// }

		// return content, nil

		return "", nil
	}
	tool.Register(t)
}
