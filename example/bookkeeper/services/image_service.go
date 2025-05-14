package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/bookkeeper-ai/bookkeeper/config"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/zhipu"
)

type ImageService struct {
	uploadDir string
	llmClient *llm.Instance
}

func NewImageService(cfg *config.Config) (*ImageService, error) {
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %v", err)
	}

	llmClient := llm.NewInstance(zhipu.Name)
	// llm.WithSingleConfig(map[string]any{
	// 	"api_key": os.Getenv("ZHIPU_API_KEY"),
	// })

	return &ImageService{
		uploadDir: uploadDir,
		llmClient: llmClient,
	}, nil
}

// UploadAndAnalyze 上传并分析图片
func (s *ImageService) UploadAndAnalyze(file *multipart.FileHeader) (string, map[string]interface{}, error) {
	// 保存文件
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	filepath := filepath.Join(s.uploadDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(filepath)
	if err != nil {
		return "", nil, fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 使用 LLM 分析图片
	ctx := context.Background()
	resp, err := s.llmClient.Invoke(ctx, []llm.Message{
		llm.NewUserTextMessage(fmt.Sprintf("请分析这张消费小票图片，提取以下信息：1. 消费金额 2. 消费类别 3. 消费日期 4. 商家名称。图片路径：%s", filepath)),
	}, nil, llm.WithModel("GLM-4-Flash"))

	if err != nil {
		return "", nil, fmt.Errorf("分析图片失败: %v", err)
	}

	// 解析 LLM 返回的结果
	// TODO: 根据实际返回格式解析结果
	result := map[string]interface{}{
		"amount":     0.0,
		"category":   "",
		"date":       time.Now(),
		"merchant":   "",
		"raw_result": resp.Result(),
	}

	return filepath, result, nil
}
