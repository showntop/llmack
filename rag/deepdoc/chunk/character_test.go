package chunk

import (
	"reflect"
	"testing"
)

func TestCharacterChunker_Chunk(t *testing.T) {
	type fields struct {
		separator   string
		chunkSize   int
		overlapSize int
	}
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				separator:   "",
				chunkSize:   5,
				overlapSize: 1,
			},
			args: args{
				content: "大家好，我是果粒奶优有果粒，欢迎关注我，让我们一起探索AI。",
			},
			want: []string{"大家好，我", "我是果粒奶", "奶优有果粒", "粒，欢迎关", "关注我，让", "让我们一起", "起探索AI", "I。"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCharacterChunker(tt.fields.chunkSize, tt.fields.overlapSize, tt.fields.separator)
			got, err := c.Chunk(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("CharacterChunker.Chunk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i := 0; i < len(got); i++ {
				t.Log(got[i])
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CharacterChunker.Chunk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecursiveCharacterChunker_Chunk(t *testing.T) {
	type fields struct {
		separators []string
		Base       *Base
	}
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				separators: []string{" ", "，", "。", "？", "！"},
				Base: &Base{
					chunkSize:   50,
					overlapSize: 1,
				},
			},
			args: args{
				content: `
为什么文本切割在RAG中很重要？RAG（Retrieval-Augmented Generation）是一种将检索机制集成到生成式语言模型中的技术，目的是通过从大量文档或知识库中检索相关信息来增强模型的生成能力。这种方法特别适用于需要广泛背景知识的任务，如问答、文章撰写或详细解释特定主题。在RAG架构中，文本切割（即将长文本分割成较短片段的过程）非常重要，原因如下：

1. **提高检索效率：** 对于大规模的文档库，直接在整个库上执行检索任务既不切实际也不高效。通过将长文本切割成较短的片段，可以使检索过程更加高效，因为短文本片段更容易被比较和索引。这样可以加快检索速度，提高整体性能。

2. **提升结果相关性：** 当查询特定信息时，与查询最相关的内容往往只占据文档中的一小部分。通过文本切割，可以更精确地匹配查询和文档片段之间的相关性，从而提高检索到的信息的准确性和相关性。这对于生成高质量、相关性强的回答尤为重要。

3. **内存和处理限制：** 当代的语言模型，尽管强大，但处理长文本时仍受到内存和计算资源的限制。将长文本分割成较短的片段可以帮助减轻这些限制，因为模型可以分别处理这些较短的文本片段，而不是一次性处理整个长文档。

4. **提高生成质量：** 在RAG框架中，检索到的文本片段将直接影响生成模块的输出。通过确保检索到高质量和高相关性的文本片段，可以提高最终生成内容的质量和准确性。

5. **适应性和灵活性：** 文本切割允许模型根据需要处理不同长度的文本，增加了模型处理各种数据源的能力。这种灵活性对于处理多样化的查询和多种格式的文档特别重要。

总之，文本切割在RAG中非常重要，因为它直接影响到检索效率、结果的相关性、系统的处理能力，以及最终生成内容的质量和准确性。通过优化文本切割策略，可以显著提升RAG系统的整体性能和用户满意度。
`,
			},
			want: []string{
				"为什么文本切割在RAG中很重要？RAG（Retrieval-Augmented",
				"Generation）是一种将检索机制集成到生成式语言模型中的技术，目的是通过从大量文档或知识库中",
				"中检索相关信息来增强模型的生成能力。这种方法特别适用于需要广泛背景知识的任务，如问答、文章撰写或详细",
				"细解释特定主题。在RAG架构中，文本切割（即将长文本分割成较短片段的过程）非常重要，原因如下：",
				"1. **提高检索效率：**",
				"对于大规模的文档库，直接在整个库上执行检索任务既不切实际也不高效。通过将长文本切割成较短的片段，可",
				"可以使检索过程更加高效，因为短文本片段更容易被比较和索引。这样可以加快检索速度，提高整体性能。",
				"2. **提升结果相关性：**",
				"当查询特定信息时，与查询最相关的内容往往只占据文档中的一小部分。通过文本切割，可以更精确地匹配查询",
				"询和文档片段之间的相关性，从而提高检索到的信息的准确性和相关性。这对于生成高质量、相关性强的回答尤为",
				"为重要。",
				"3. **内存和处理限制：**",
				"当代的语言模型，尽管强大，但处理长文本时仍受到内存和计算资源的限制。将长文本分割成较短的片段可以帮",
				"帮助减轻这些限制，因为模型可以分别处理这些较短的文本片段，而不是一次性处理整个长文档。",
				"4. **提高生成质量：**",
				"在RAG框架中，检索到的文本片段将直接影响生成模块的输出。通过确保检索到高质量和高相关性的文本片段",
				"段，可以提高最终生成内容的质量和准确性。",
				"5. **适应性和灵活性：**",
				"文本切割允许模型根据需要处理不同长度的文本，增加了模型处理各种数据源的能力。这种灵活性对于处理多样",
				"样化的查询和多种格式的文档特别重要。",
				"总之，文本切割在RAG中非常重要，因为它直接影响到检索效率、结果的相关性、系统的处理能力，以及最终",
				"终生成内容的质量和准确性。通过优化文本切割策略，可以显著提升RAG系统的整体性能和用户满意度。",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewRecursiveCharacterChunker(tt.fields.Base.chunkSize, tt.fields.Base.overlapSize, tt.fields.separators)
			got, err := c.Chunk(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecursiveCharacterChunker.Chunk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RecursiveCharacterChunker.Chunk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecursiveCharacterChunker_chunk(t *testing.T) {
	type fields struct {
		separators []string
		Base       *Base
	}
	type args struct {
		text          string
		separators    []string
		chunkSize     int
		keepSeparator bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RecursiveCharacterChunker{
				separators: tt.fields.separators,
				Base:       tt.fields.Base,
			}
			if got := c.chunk(tt.args.text, tt.args.separators, tt.args.chunkSize, tt.args.keepSeparator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RecursiveCharacterChunker.chunk() = %v, want %v", got, tt.want)
			}
		})
	}
}
