package agents

import (
	"fmt"
	"strings"
)

// PromptBuilder builds smart prompts for AI generation
type PromptBuilder struct {
	tagTemplates  map[string]string
	descTemplates map[string]string
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	builder := &PromptBuilder{
		tagTemplates:  make(map[string]string),
		descTemplates: make(map[string]string),
	}
	builder.initTemplates()
	return builder
}

// initTemplates initializes prompt templates
func (b *PromptBuilder) initTemplates() {
	// Tag generation templates
	b.tagTemplates["default"] = `文件名：%s
文件类型：%s
用户关键词：%s
用户标签：%s
用户描述：%s

请基于以上信息为这个文件生成3-5个精准的标签。
标签应该反映文件的内容、类型、用途或特征。
要求：
1. 标签要简洁明了，每个标签2-4个字
2. 优先使用用户提供的标签作为参考
3. 结合关键词信息生成相关标签
4. 用逗号分隔标签

生成的标签：`

	// Description generation templates
	b.descTemplates["default"] = `文件名：%s
文件类型：%s
用户关键词：%s
用户标签：%s
用户描述：%s

请基于以上信息为这个文件生成一个简洁准确的描述（1-2句话）。
描述应该概括文件的用途、内容或特点。
要求：
1. 描述要简洁明了，不超过100个字
2. 优先参考用户提供的描述
3. 结合关键词和标签信息

生成的描述：`
}

// BuildTagPrompt builds a prompt for tag generation
func (b *PromptBuilder) BuildTagPrompt(
	fileName, fileType string,
	userKeywords string,
	userTags []string,
	userDescription string,
) string {
	tagsStr := strings.Join(userTags, ", ")
	template := b.tagTemplates["default"]

	return fmt.Sprintf(template,
		fileName,
		fileType,
		userKeywords,
		tagsStr,
		userDescription,
	)
}

// BuildDescriptionPrompt builds a prompt for description generation
func (b *PromptBuilder) BuildDescriptionPrompt(
	fileName, fileType string,
	userKeywords string,
	userTags []string,
	userDescription string,
) string {
	tagsStr := strings.Join(userTags, ", ")
	template := b.descTemplates["default"]

	return fmt.Sprintf(template,
		fileName,
		fileType,
		userKeywords,
		tagsStr,
		userDescription,
	)
}

// BuildMetadataSystemPrompt builds a prompt for metadata JSON generation.
func (b *PromptBuilder) BuildMetadataSystemPrompt() string {
	return "你是一个文件元数据补全助手。请根据上下文生成 tags 和 description，并仅返回 JSON：{\"tags\":[],\"description\":\"\"}。"
}

// BuildMetadataUserPrompt builds a prompt for metadata JSON generation.
func (b *PromptBuilder) BuildMetadataUserPrompt(input *MetadataGenerationInput) string {
	tagsStr := ""
	if input != nil {
		tagsStr = strings.Join(input.UserTags, ", ")
		return fmt.Sprintf(
			"文件名：%s\n文件类型：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请输出 metadata JSON。",
			input.FileName,
			input.FileType,
			input.UserKeywords,
			tagsStr,
			input.UserDescription,
		)
	}

	return fmt.Sprintf(
		"文件名：%s\n文件类型：%s\n用户关键词：%s\n用户标签：%s\n用户描述：%s\n请输出 metadata JSON。",
		"", "", "", tagsStr, "",
	)
}
