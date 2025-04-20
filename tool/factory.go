package tool

import "sync"

// 注册表锁
var registerLock sync.RWMutex

// Repo 工具仓库
type Repo interface {
	// FetchTool(context.Context, int64, string) (*Metadata, error)
}

// Factory 工具工厂
type Factory struct {
	repo Repo
}

// NewFactory ...
func NewFactory(repo Repo) *Factory {
	return &Factory{repo: repo}
}

// Spawn ...
func (f *Factory) Spawn(name string) *Tool {
	registerLock.RLock()
	defer registerLock.RUnlock()
	if t, ok := tools[name]; ok {
		return t
	}
	return nil
}

var defaultFactory = NewFactory(nil)

// WithRepo ...
func WithRepo(repo Repo) {
	defaultFactory.repo = repo
}

// Spawn ...
func Spawn(name string) *Tool {
	x := defaultFactory.Spawn(name)
	if x == nil {
		return NilTool
	}
	return x
}

// Tools 工具注册表
var tools map[string]*Tool = make(map[string]*Tool)

// Register 注册工具
func Register(t *Tool) {
	registerLock.Lock()
	defer registerLock.Unlock()
	tools[t.Name] = t
}
