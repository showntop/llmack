package wikipedia

import (
	"context"
	"fmt"

	"github.com/showntop/llmack/tool"
	gowiki "github.com/trietmn/go-wiki"
)

func init() {
	t := tool.CodeTool{}
	t.Meta.Name = "wikipedia_search"
	t.Meta.Description = "A tool for performing a Wikipedia search and extracting snippets and webpages. Input should be a search query."
	t.Meta.Parameters = append(t.Meta.Parameters, tool.Parameter{
		Name: "query", Type: tool.String, Required: true, LLMDescrition: `key words for searching, this should be in the language of "language" parameter`,
		Options: []string{"de", "en", "fr", "hi", "ja", "ko", "pl", "pt", "ro", "uk", "vi", "zh"},
	}, tool.Parameter{
		Name: "language", Type: tool.String, Required: true, LLMDescrition: `
		language of the wikipedia to be searched,
		only "de" for German,
		"en" for English,
		"fr" for French,
		"hi" for Hindi,
		"ja" for Japanese,
		"ko" for Korean,
		"pl" for Polish,
		"pt" for Portuguese,
		"ro" for Romanian,
		"uk" for Ukrainian,
		"vi" for Vietnamese,
		and "zh" for Chinese are supported
		`,
	})
	t.Invokex = Invoke
	tool.Register(t.Name(), &t)
}

// Invoke ...
// func Invoke(ctx context.Context, args map[string]any) interface{} {
func Invoke(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	lang, _ := args["language"].(string)
	if query == "" {
		return "Please input query", fmt.Errorf("query is empty")
	}
	_ = lang

	// Search for the Wikipedia page title
	searchResult, _, err := gowiki.Search("Why is the sky blue", 3, false)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("This is your search result: %v\n", searchResult)

	// Get the page
	page, err := gowiki.GetPage("Rafael Nadal", -1, false, true)
	if err != nil {
		fmt.Println(err)
	}

	// Get the content of the page
	content, err := page.GetContent()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("This is the page content: %v\n", content)

	return content, nil
}
