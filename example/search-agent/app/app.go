package app

import (
	"context"
	"fmt"

	"github.com/showntop/llmack/engine"
	"github.com/showntop/llmack/example/search-agent/workflow"
	"github.com/showntop/llmack/llm"
)

// Application ...
type Application struct {
}

// NewApplication ...
func NewApplication() *Application {
	return &Application{}
}

// SearchCommand ...
type SearchCommand struct {
	Query string
}

// Search ...
func (app *Application) Search(ctx context.Context, command SearchCommand) (SearchEventStream, error) {
	events := NewSearchEventStream()
	go func() {
		defer events.Close()

		settings := engine.DefaultSettings()
		settings.Workflow = workflow.BuildWorkflow()
		runx := engine.NewWorkflowEngine(settings)
		esm := runx.Execute(ctx, engine.Input{Query: command.Query})
		for evt := esm.Next(); evt != nil; evt = esm.Next() {
			fmt.Printf("main event name:%v data: %+v \n ", evt.Name, evt.Data)
			if evt.Source == "related" {
				events.Push(&SearchEvent{Status: "related", Related: evt.Data.(*llm.Chunk).Delta.Message.Content()})
			} else if evt.Source == "answer" {
				events.Push(&SearchEvent{Status: "answer", Answer: evt.Data.(*llm.Chunk).Delta.Message.Content()})
			} else if evt.Source == "sources" {
				events.Push(&SearchEvent{Status: "sources", Sources: evt.Data})
			} else if evt.Source == "images" {
				events.Push(&SearchEvent{Status: "images", Images: evt.Data})
			} else if evt.Source == "videos" {
				events.Push(&SearchEvent{Status: "videos", Videos: evt.Data})
			}
		}
	}()
	return events, nil
}
