package vision

// InvokeOption is a function that configures a InvokeOptions.
type InvokeOption func(*InvokeOptions)

// InvokeOptions ...
type InvokeOptions struct {
	ApiKey string
}

// WithApiKey ...
func WithApiKey(apiKey string) InvokeOption {
	return func(o *InvokeOptions) {
		o.ApiKey = apiKey
	}
}
