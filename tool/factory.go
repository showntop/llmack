package tool

import "context"

// Repo 工具仓库
type Repo interface {
	FetchAPITool(context.Context, int64, string) (*APIToolBundle, error)
}

// Factory 工具工厂
type Factory struct {
	repo Repo
}

// NewFactory ...
func NewFactory(repo Repo) *Factory {
	return &Factory{repo: repo}
}

// Instantiate ...
func (f *Factory) Instantiate(providerID int64, providerKind string, name string) Tool {
	if providerKind == "api" {
		bundle, err := f.repo.FetchAPITool(context.TODO(), providerID, name)
		if err != nil {
			return &NilTool{Target: name}
		}
		return NewAPITool(*bundle)
	} else if providerKind == "code" {
		return NewCodeTool(name)
	}
	return &NilTool{Target: name}
}

var defaultFactory = NewFactory(nil)

// WithRepo ...
func WithRepo(repo Repo) {
	defaultFactory.repo = repo
}

// Instantiate ...
func Instantiate(providerID int64, providerKind string, name string) Tool {
	return defaultFactory.Instantiate(providerID, providerKind, name)
}
