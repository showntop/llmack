package optimizer

// Example ...
type Example struct {
	store     map[string]any
	inputKeys []string
}

// Examplex ...
func Examplex(pairs ...any) *Example {
	ex := &Example{
		store: make(map[string]any),
	}
	if len(pairs)%2 != 0 {

	}
	for i := 0; i < len(pairs); i += 2 {
		ex.store[pairs[i].(string)] = pairs[i+1]
	}
	return ex
}

// Set ...
func (ex *Example) Set(pairs ...string) {
	if len(pairs)%2 != 0 {

	}
	for i := 0; i < len(pairs); i += 2 {
		ex.store[pairs[i]] = pairs[i+1]
	}
}

// Get ...
func (ex *Example) Get(key string) any {
	return ex.store[key]
}

// WithInputKeys ...
func (ex *Example) WithInputKeys(keys ...string) *Example {
	ex.inputKeys = keys
	return ex
}

// Inputs ...
func (ex *Example) Inputs(keys ...string) map[string]any {
	if len(ex.inputKeys) == 0 {
		return ex.store
	}
	result := make(map[string]any)
	for i := 0; i < len(ex.inputKeys); i++ {
		result[ex.inputKeys[i]] = ex.store[ex.inputKeys[i]]
	}
	return result
}
