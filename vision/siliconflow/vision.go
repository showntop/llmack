package siliconflow

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/showntop/llmack/vision"
)

// Name ...
var Name = "siliconflow"

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

	url := "https://api.siliconflow.cn/v1/images/generations"

	params := map[string]any{
		"prompt": prompt,
		"model":  "black-forest-labs/FLUX.1-schnell",
		"seed":   4999999999,
	}
	payload, _ := json.Marshal(params)
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return "null", err
	}
	req.Header.Add("Authorization", "Bearer "+options.ApiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "null", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "null", err
	}
	var result Result
	json.Unmarshal(body, &result)

	return result.Images[0].Url, nil
}

type Result struct {
	Images []struct {
		Url string `json:"url"`
	} `json:"images"`
	Timings struct {
		Inference float64 `json:"inference"`
	} `json:"timings"`
	Seed     int    `json:"seed"`
	SharedId string `json:"shared_id"`
	Data     []struct {
		Url string `json:"url"`
	} `json:"data"`
	Created int `json:"created"`
}
