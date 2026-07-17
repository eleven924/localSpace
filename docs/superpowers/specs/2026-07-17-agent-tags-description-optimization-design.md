# Agent生成标签和描述逻辑优化设计

**日期：** 2026-07-17
**版本：** 1.0
**状态：** 待审核

## 概述

本设计旨在优化LocalSpace中导入文件时的AI标签和描述生成逻辑，通过引入eino agent框架，实现更智能、更准确的元数据生成能力。

## 背景和需求

### 现状
- 当前AI服务仅基于文件名和文件类型生成标签和描述
- 缺少用户输入的上下文信息
- 没有外部信息获取能力
- 生成结果准确性有限

### 目标
- 支持用户提供关键词作为AI生成的提示
- 利用用户输入的标签和描述作为上下文
- 支持Agent自主决定是否需要网络搜索获取信息
- 整合多源信息生成更准确的标签和描述

## 架构设计

### 整体架构

```
用户输入 (关键词/标签/描述)
    ↓
FileService.ImportFile
    ↓
AgentService.GenerateTags/GenerateDescription
    ↓
Agent (自主决策是否使用工具)
    ↓
AI生成结果
    ↓
返回给调用方
```

### 组件设计

#### 1. AgentService

**职责：** 管理Agent实例，提供统一的接口给FileService调用

**位置：** `app/services/agent_service.go`

**主要方法：**
```go
type AgentService struct {
    tagAgent       *TagAgent
    descriptionAgent *DescriptionAgent
    configRepo     *repositories.ConfigRepository
    toolRegistry   *ToolRegistry
}

// 生成标签
func (s *AgentService) GenerateTags(
    ctx context.Context,
    fileName, fileType string,
    userKeywords string,
    userTags []string,
    userDescription string,
) ([]string, error)

// 生成描述
func (s *AgentService) GenerateDescription(
    ctx context.Context,
    fileName, fileType string,
    userKeywords string,
    userTags []string,
    userDescription string,
) (string, error)

// 批量生成（保留接口，暂不实现）
func (s *AgentService) GenerateBatch(
    ctx context.Context,
    files []FileContext,
) ([]*models.AIAnalysis, error)
```

#### 2. TagAgent

**职责：** 使用eino agent框架生成文件标签，具备工具调用能力

**位置：** `app/agents/tag_agent.go`

**核心能力：**
- 接收多源输入（文件名、类型、用户关键词、标签、描述）
- 构建智能提示词
- 决策是否需要调用网络搜索工具
- 整合搜索结果生成最终标签列表

**主要方法：**
```go
type TagAgent struct {
    chatModel  model.ChatModel
    tools      []Tool
    promptBuilder *PromptBuilder
}

// 生成标签
func (a *TagAgent) Generate(
    ctx context.Context,
    input *TagGenerationInput,
) ([]string, error)

// 决策是否需要搜索
func (a *TagAgent) decideNeedSearch(
    ctx context.Context,
    input *TagGenerationInput,
) (bool, error)
```

#### 3. DescriptionAgent

**职责：** 使用eino agent框架生成文件描述，具备工具调用能力

**位置：** `app/agents/description_agent.go`

**核心能力：**
- 接收多源输入（文件名、类型、用户关键词、标签、描述）
- 构建智能提示词
- 决策是否需要调用网络搜索工具
- 整合搜索结果生成最终描述文本

**主要方法：**
```go
type DescriptionAgent struct {
    chatModel  model.ChatModel
    tools      []Tool
    promptBuilder *PromptBuilder
}

// 生成描述
func (a *DescriptionAgent) Generate(
    ctx context.Context,
    input *DescriptionGenerationInput,
) (string, error)

// 决策是否需要搜索
func (a *DescriptionAgent) decideNeedSearch(
    ctx context.Context,
    input *DescriptionGenerationInput,
) (bool, error)
```

#### 4. WebSearchTool

**职责：** 提供网络搜索能力，供Agent调用

**位置：** `app/tools/web_search_tool.go`

**主要功能：**
- 执行网络搜索
- 返回结构化搜索结果
- 处理搜索错误和超时

**接口定义：**
```go
type WebSearchTool struct {
    searchClient SearchClient
    timeout      time.Duration
}

// 执行搜索
func (t *WebSearchTool) Search(
    ctx context.Context,
    query string,
) (*SearchResult, error)

// 工具描述（供Agent理解）
func (t *WebSearchTool) Description() string
```

#### 5. PromptBuilder

**职责：** 构建智能提示词，整合多源信息

**位置：** `app/agents/prompt_builder.go`

**主要方法：**
```go
type PromptBuilder struct {
    templates map[string]string
}

// 构建标签生成提示词
func (b *PromptBuilder) BuildTagPrompt(
    fileName, fileType string,
    userKeywords string,
    userTags []string,
    userDescription string,
) string

// 构建描述生成提示词
func (b *PromptBuilder) BuildDescriptionPrompt(
    fileName, fileType string,
    userKeywords string,
    userTags []string,
    userDescription string,
) string
```

### 数据模型扩展

#### ImportFileRequest 扩展

```go
type ImportFileRequest struct {
    FilePath    string   // 文件路径
    FileName    string   // 文件名
    Description string   // 用户提供的描述
    Tags        []string // 用户提供的标签
    Keywords    string   // 新增：用户输入的关键词
}
```

#### Agent输入模型

```go
// 标签生成输入
type TagGenerationInput struct {
    FileName         string   // 文件名
    FileType         string   // 文件类型
    UserKeywords     string   // 用户关键词
    UserTags         []string // 用户标签
    UserDescription  string   // 用户描述
    Metadata         models.Metadata // 文件元数据
}

// 描述生成输入
type DescriptionGenerationInput struct {
    FileName         string   // 文件名
    FileType         string   // 文件类型
    UserKeywords     string   // 用户关键词
    UserTags         []string // 用户标签
    UserDescription  string   // 用户描述
    Metadata         models.Metadata // 文件元数据
}
```

## 数据流设计

### 1. 标签生成流程

```
用户输入关键词 → FileService.ImportFile
    ↓
构建ImportFileRequest (包含Keywords)
    ↓
AgentService.GenerateTags
    ↓
构建TagGenerationInput
    ↓
调用TagAgent.Generate
    ↓
Agent分析输入，构建智能提示词
    ↓
Agent自主决策：
  - 如果需要更多信息 → 调用WebSearchTool
  - 如果信息充足 → 直接生成
    ↓
整合所有信息，调用LLM生成标签
    ↓
解析和验证标签结果
    ↓
返回优化后的标签列表
```

### 2. 描述生成流程

```
用户输入关键词 → FileService.ImportFile
    ↓
构建ImportFileRequest (包含Keywords)
    ↓
AgentService.GenerateDescription
    ↓
构建DescriptionGenerationInput
    ↓
调用DescriptionAgent.Generate
    ↓
Agent分析输入，构建智能提示词
    ↓
Agent自主决策：
  - 如果需要更多信息 → 调用WebSearchTool
  - 如果信息充足 → 直接生成
    ↓
整合所有信息，调用LLM生成描述
    ↓
返回优化后的描述文本
```

### 3. 网络搜索决策流程

```
Agent接收输入信息
    ↓
分析文件名、关键词、现有标签
    ↓
判断信息是否充足：
  - 文件名包含专有名词？
  - 关键词是否过于模糊？
  - 文件类型是否需要外部信息？
    ↓
如果判断需要搜索 → 调用WebSearchTool
    ↓
获取搜索结果
    ↓
整合搜索结果到提示词
    ↓
重新生成标签/描述
    ↓
返回最终结果
```

## 错误处理设计

### 1. Agent调用错误处理

**错误分类：**
- **配置错误** - AI配置缺失或无效，返回友好提示
- **网络错误** - 网络搜索失败，降级为不使用搜索
- **API错误** - LLM调用失败，返回基础标签/描述
- **超时错误** - 设置合理超时，避免长时间阻塞

**处理策略：**
```go
func (s *AgentService) GenerateTags(
    ctx context.Context,
    fileName, fileType string,
    userKeywords string,
    userTags []string,
    userDescription string,
) ([]string, error) {
    // 检查AI配置
    config, err := s.configRepo.GetAIConfig()
    if err != nil || !config.Enabled {
        return []string{}, nil // 返回空标签，不阻塞导入
    }

    // 尝试Agent生成
    input := &TagGenerationInput{
        FileName:        fileName,
        FileType:        fileType,
        UserKeywords:    userKeywords,
        UserTags:        userTags,
        UserDescription: userDescription,
    }

    tags, err := s.tagAgent.Generate(ctx, input)
    if err != nil {
        // 降级处理：返回基础标签
        return s.generateBasicTags(fileName, fileType), nil
    }

    return tags, nil
}
```

### 2. 工具调用错误处理

**网络搜索工具：**
- 搜索失败时不阻塞主流程
- 记录日志便于调试
- Agent可以基于已有信息继续生成

```go
func (t *WebSearchTool) Search(
    ctx context.Context,
    query string,
) (*SearchResult, error) {
    // 设置超时
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    result, err := t.searchClient.Search(ctx, query)
    if err != nil {
        // 记录错误但不中断流程
        log.Printf("Web search failed: %v", err)
        return &SearchResult{}, nil // 返回空结果，允许继续
    }

    return result, nil
}
```

### 3. 导入流程集成

**FileService集成策略：**
- AI生成失败不影响文件导入成功
- 提供异步生成选项
- 支持手动重新生成

```go
func (s *FileService) ImportFile(req ImportFileRequest) error {
    // ... 现有的导入逻辑 ...

    // AI生成标签和描述
    var tags []string
    var description string

    if s.aiService != nil {
        // 尝试使用AgentService生成
        if agentTags, err := s.agentService.GenerateTags(
            ctx, req.FileName, fileType, req.Keywords, req.Tags, req.Description,
        ); err == nil {
            tags = agentTags
        }

        if agentDesc, err := s.agentService.GenerateDescription(
            ctx, req.FileName, fileType, req.Keywords, req.Tags, req.Description,
        ); err == nil {
            description = agentDesc
        }
    }

    // 如果AI生成失败，使用用户输入或空值
    if len(tags) == 0 && len(req.Tags) > 0 {
        tags = req.Tags
    }
    if description == "" && req.Description != "" {
        description = req.Description
    }

    // 创建文件记录
    file := &models.File{
        // ... 其他字段 ...
        Tags:        tags,
        Description: description,
    }

    // ... 保存逻辑 ...
}
```

## 测试策略

### 1. 单元测试

**AgentService测试：**
- 测试各种输入组合
- 测试错误处理路径
- Mock Agent和工具调用

**Agent测试：**
- 测试提示词构建逻辑
- 测试决策逻辑
- Mock LLM和工具调用

**工具测试：**
- 测试网络搜索工具
- 测试错误处理
- Mock外部API

### 2. 集成测试

**端到端测试：**
- 测试完整导入流程
- 测试AI生成集成
- 测试各种文件类型

**工具集成测试：**
- 测试MCP工具集成
- 测试网络搜索实际调用

### 3. 用户场景测试

**典型场景：**
- 用户不提供任何输入 → AI基于文件名生成
- 用户提供关键词 → AI结合关键词生成
- 不熟悉的软件名 → Agent自动搜索
- 网络不可用 → 降级处理

## 性能考虑

### 1. 响应时间优化

**缓存策略：**
- 缓存常见文件类型的标签生成结果
- 缓存网络搜索结果（相同查询）
- 设置合理的缓存过期时间

**异步处理：**
- AI生成不阻塞文件导入
- 提供生成状态查询
- 支持后台队列处理

### 2. 资源管理

**并发控制：**
- 限制同时进行的AI调用数量
- 避免内存和连接泄漏
- 合理的超时设置

**成本控制：**
- 记录API调用次数和成本
- 设置使用限制
- 提供成本监控

### 3. 用户体验优化

**渐进式增强：**
- 先返回基础标签/描述
- 后台优化生成结果
- 支持结果刷新

**状态反馈：**
- 显示AI生成状态
- 提供进度指示
- 错误友好提示

## 向后兼容性

### 1. 现有API兼容

**FileService.ImportFile：**
- 保持现有方法签名不变
- `Keywords` 字段为可选，默认为空字符串
- 现有调用方无需修改

**AIService保留：**
- 保留现有 `AIService` 作为降级方案
- 新的 `AgentService` 作为主要实现
- 配置切换机制

### 2. 数据库兼容

**无schema变更：**
- 不修改现有数据库结构
- `Keywords` 只在内存中传递
- 保持现有字段不变

### 3. 配置兼容

**AI配置扩展：**
- 现有AI配置继续有效
- 新增Agent相关配置项
- 默认禁用高级功能

```go
type AIConfig struct {
    ID      uint   `json:"id"`
    APIKey  string `json:"apiKey"`
    Model   string `json:"model"`
    BaseURL string `json:"baseURL"`
    Enabled bool   `json:"enabled"`

    // 新增字段
    EnableAgent     bool `json:"enableAgent"`      // 启用Agent功能
    EnableWebSearch bool `json:"enableWebSearch"`  // 启用网络搜索
    MaxTokens       int  `json:"maxTokens"`        // 最大token数
    Timeout         int  `json:"timeout"`          // 超时时间（秒）
}
```

## 安全考虑

### 1. 数据隐私

**用户输入保护：**
- 关键词和标签不记录到日志
- 网络搜索查询匿名化处理
- 敏感信息过滤

**网络搜索安全：**
- 避免在搜索中泄露文件路径
- 过滤敏感关键词
- 使用安全搜索API

### 2. API安全

**密钥管理：**
- AI API密钥安全存储
- 不在代码中硬编码
- 定期轮换机制

**访问控制：**
- 验证AI配置有效性
- 限制API调用频率
- 异常访问监控

### 3. 输入验证

**用户输入验证：**
- 关键词长度限制（最大500字符）
- 特殊字符过滤
- 防止注入攻击

**Agent输出验证：**
- 标签格式验证（3-5个标签）
- 描述长度限制（最大200字符）
- 敏感内容过滤

## 实施步骤

### 阶段1：基础架构搭建
1. 创建 `app/agents/` 目录结构
2. 实现 `AgentService` 基础框架
3. 集成eino agent依赖
4. 配置管理扩展

### 阶段2：Agent实现
1. 实现 `TagAgent` 基础功能
2. 实现 `DescriptionAgent` 基础功能
3. 智能提示词构建逻辑
4. 基础测试

### 阶段3：工具集成
1. 实现MCP网络搜索工具
2. Agent工具调用能力
3. 错误处理和降级
4. 集成测试

### 阶段4：服务集成
1. FileService集成AgentService
2. 导入流程改造
3. UI适配（关键词输入框）
4. 端到端测试

### 阶段5：优化和完善
1. 性能优化
2. 缓存实现
3. 监控和日志
4. 文档完善

## 技术栈

### 核心依赖
- **eino agent框架** - Agent实现和工具调用
- **现有OpenAI集成** - LLM调用保持不变
- **MCP工具** - 网络搜索能力

### 目录结构

```
app/
├── services/
│   ├── agent_service.go       # Agent服务管理
│   └── ai_service.go          # 保留原有AI服务
├── agents/
│   ├── agent_base.go          # Agent基类
│   ├── tag_agent.go           # 标签生成Agent
│   ├── description_agent.go   # 描述生成Agent
│   └── prompt_builder.go      # 提示词构建器
├── tools/
│   ├── web_search_tool.go     # 网络搜索工具
│   └── tool_registry.go       # 工具注册表
├── models/
│   ├── file.go                # 扩展ImportFileRequest
│   └── config.go              # 扩展AI配置
└── repositories/
    └── config_repository.go   # 配置仓储
```

## 成功标准

### 功能完整性
- ✅ 支持用户输入关键词
- ✅ Agent能够自主决策是否需要网络搜索
- ✅ 整合多源信息生成准确标签和描述
- ✅ 错误处理和降级机制完善
- ✅ 向后兼容现有功能

### 性能指标
- 标签生成响应时间 < 3秒（无网络搜索）
- 标签生成响应时间 < 8秒（有网络搜索）
- 描述生成响应时间 < 3秒（无网络搜索）
- 描述生成响应时间 < 8秒（有网络搜索）
- 并发支持 ≥ 10个同时请求

### 用户体验
- AI生成失败不影响文件导入
- 提供清晰的状态反馈
- 错误提示友好易懂

## 风险和限制

### 技术风险
- **eino agent框架稳定性** - 框架可能存在不稳定性，需要充分测试
- **网络搜索依赖** - 依赖外部搜索服务，可能存在服务中断风险
- **API成本** - 频繁的API调用可能增加成本

### 限制
- **网络依赖** - 网络搜索功能需要互联网连接
- **语言支持** - 主要优化中文文件名，其他语言可能效果不佳
- **文件类型** - 某些特殊文件类型可能需要定制化处理

## 后续扩展

### 功能扩展
- 支持更多文件类型的定制化生成策略
- 支持用户自定义生成模板
- 支持批量生成功能
- 支持生成结果的手动修正和学习

### 技术优化
- 引入向量数据库实现语义搜索
- 支持本地模型部署
- 实现增量学习和用户反馈机制

## 附录

### 示例提示词模板

**标签生成提示词：**
```
文件名：{fileName}
文件类型：{fileType}
用户关键词：{userKeywords}
用户标签：{userTags}
用户描述：{userDescription}

请基于以上信息为这个文件生成3-5个精准的标签。
标签应该反映文件的内容、类型、用途或特征。
要求：
1. 标签要简洁明了，每个标签2-4个字
2. 优先使用用户提供的标签作为参考
3. 结合关键词信息生成相关标签
4. 如果需要更多信息，可以使用搜索工具
5. 用逗号分隔标签

生成的标签：
```

**描述生成提示词：**
```
文件名：{fileName}
文件类型：{fileType}
用户关键词：{userKeywords}
用户标签：{userTags}
用户描述：{userDescription}

请基于以上信息为这个文件生成一个简洁准确的描述（1-2句话）。
描述应该概括文件的用途、内容或特点。
要求：
1. 描述要简洁明了，不超过50个字
2. 优先参考用户提供的描述
3. 结合关键词和标签信息
4. 如果需要更多信息，可以使用搜索工具

生成的描述：
```

### 配置示例

**AI配置（config.json）：**
```json
{
  "ai": {
    "enabled": true,
    "apiKey": "sk-...",
    "model": "gpt-4",
    "baseURL": "https://api.openai.com/v1",
    "enableAgent": true,
    "enableWebSearch": true,
    "maxTokens": 500,
    "timeout": 30
  }
}
```