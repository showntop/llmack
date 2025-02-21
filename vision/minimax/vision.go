package minimax

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/showntop/llmack/vision"
)

// Name ...
var Name = "minimax"

func init() {
	vision.Register(Name, &Vision{})
}

type Vision struct {
}

func (v *Vision) GenerateImage(ctx context.Context, prompt string, optFuncs ...vision.InvokeOption) (string, error) {
	options := vision.InvokeOptions{}
	for _, optFunc := range optFuncs {
		optFunc(&options)
	}

	url := "https://api.minimax.chat/v1/image_generation"
	api_key := options.ApiKey

	payload := map[string]interface{}{
		"model":            "image-01",
		"prompt":           "men Dressing in white t shirt, full-body stand front view image :25, outdoor, Venice beach sign, full-body image, Los Angeles, Fashion photography of 90s, documentary, Film grain, photorealistic",
		"aspect_ratio":     "16:9",
		"response_format":  "base64",
		"n":                3,
		"prompt_optimizer": true,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "null", err
	}
	req.Header.Add("Authorization", "Bearer "+api_key)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "null", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "null", err
	}
	fmt.Println("zx", string(raw))
	var result Result
	json.Unmarshal(raw, &result)

	for i := 0; i < len(result.Data.ImageBase64); i++ {
		data, _ := base64.RawStdEncoding.DecodeString(result.Data.ImageBase64[i])
		ff, _ := os.Create(fmt.Sprintf("image_%d.jpg", i))
		ff.Write(data)
	}

	return "null", nil
}

type Result struct {
	Id   string `json:"id"`
	Data struct {
		ImageBase64 []string `json:"image_base64"`
	} `json:"data"`
	Metadata struct {
		FailedCount  string `json:"failed_count"`
		SuccessCount string `json:"success_count"`
	} `json:"metadata"`
	BaseResp struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
}
