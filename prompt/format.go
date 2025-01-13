package prompt

import (
	"fmt"
	"regexp"
)

var regex = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]{0,29}|#histories#|#query#|#context#)\}\}`)
var withVariableTmplRegex = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]{0,29}|#[a-zA-Z0-9_]{1,50}\.[a-zA-Z0-9_\.]{1,100}#|#histories#|#query#|#context#)\}\}`)

// TemplateFormatter ...
type TemplateFormatter struct {
	template         string
	withVariableTmpl bool
	regex            *regexp.Regexp
	variableKeys     []string
}

// NewTemplateFormatter ...
func NewTemplateFormatter(template string, withVariableTmpl bool) *TemplateFormatter {
	regex := regex
	if withVariableTmpl {
		regex = withVariableTmplRegex
	}
	return &TemplateFormatter{
		template:         template,
		withVariableTmpl: withVariableTmpl,
		regex:            regex,
		variableKeys:     extract(template, regex),
	}
}

func extract(template string, regex *regexp.Regexp) []string {
	return regex.FindAllString(template, -1)
}

// Format ...
func (p *TemplateFormatter) Format(inputs map[string]any, removeVars bool) string {
	replacer := func(match string) string {
		key := regex.FindStringSubmatch(match)[1]
		value, ok := inputs[key]
		if !ok {
			value = match
		} else {
			value = fmt.Sprintf("%+v", inputs[key])
		}
		valuex, _ := value.(string)

		if removeVars {
			return removeTemplateVariables(valuex, p.withVariableTmpl)
		}
		return valuex
	}
	prompt := regex.ReplaceAllStringFunc(p.template, replacer)
	prompt = regexp.MustCompile(`<\|.*?\|>`).ReplaceAllString(prompt, "")
	return prompt
}

func removeTemplateVariables(text string, withVariableTmpl bool) string {
	regex := regex
	if withVariableTmpl {
		regex = withVariableTmplRegex
	}
	return regex.ReplaceAllString(text, "{$1}")
}
