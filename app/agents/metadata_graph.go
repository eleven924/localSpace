package agents

import (
	"context"
	"fmt"
	"strings"

	"LocalSpace/app/models"
	"LocalSpace/app/tools"

	"github.com/cloudwego/eino/compose"
)

const (
	metadataGraphNormalize = "normalize_input"
	metadataGraphMerge     = "merge_evidence"
	metadataGraphAnalyze   = "analyze_agent"
	metadataGraphValidate  = "validate_result"
	metadataGraphRecommend = "recommendation_hints"
	metadataGraphQuality   = "quality_gate"

	metadataRouteImage     = "route_image"
	metadataRouteDocument  = "route_document"
	metadataRouteMedia     = "route_media"
	metadataRouteArchive   = "route_archive"
	metadataRouteGeneric   = "route_generic"
	metadataRouteAutoApply = "quality_auto_apply"
	metadataRouteReview    = "quality_needs_review"
	metadataRouteLow       = "quality_low_confidence"

	metadataAutoApplyThreshold = 0.85
	metadataReviewThreshold    = 0.60
)

// MetadataGraphState 是文件分析 Graph 的运行时状态。
// 文件路径、文件 ID 等后端字段只保留在这里，不会直接进入模型 Prompt。
type MetadataGraphState struct {
	Request        *AnalysisRequest
	AIConfig       *models.AIConfig
	AvailableTools []tools.Tool
	Result         *MetadataAnalysisResult
	Route          string
	CurrentStage   string
	StageHistory   []string
	Errors         []string
	Confidence     float64
	NeedsReview    bool
	QualityStatus  string
}

// MetadataAnalysisGraph 使用 Eino Graph 编排文件元数据分析流程。
// Graph 本身负责确定性的业务流程，Agent 只负责局部的证据判断和结构化生成。
type MetadataAnalysisGraph struct {
	agent       MetadataAgent
	runnable    compose.Runnable[*MetadataGraphState, *MetadataGraphState]
	buildErr    error
	maxRunSteps int
}

// NewMetadataAnalysisGraph 创建并编译元数据分析 Graph。
func NewMetadataAnalysisGraph(agent MetadataAgent) *MetadataAnalysisGraph {
	graphRunner := &MetadataAnalysisGraph{
		agent:       agent,
		maxRunSteps: 32,
	}
	if agent == nil {
		graphRunner.buildErr = fmt.Errorf("metadata graph agent is not configured")
		return graphRunner
	}

	graph := compose.NewGraph[*MetadataGraphState, *MetadataGraphState]()
	if err := addMetadataGraphNodes(graph, agent); err != nil {
		graphRunner.buildErr = err
		return graphRunner
	}
	if err := addMetadataGraphEdges(graph); err != nil {
		graphRunner.buildErr = err
		return graphRunner
	}

	runnable, err := graph.Compile(context.Background(), compose.WithMaxRunSteps(graphRunner.maxRunSteps))
	if err != nil {
		graphRunner.buildErr = fmt.Errorf("compile metadata graph: %w", err)
		return graphRunner
	}
	graphRunner.runnable = runnable
	return graphRunner
}

func addMetadataGraphNodes(graph *compose.Graph[*MetadataGraphState, *MetadataGraphState], agent MetadataAgent) error {
	if err := graph.AddLambdaNode(metadataGraphNormalize, compose.InvokableLambda(metadataNormalizeNode)); err != nil {
		return fmt.Errorf("add %s node: %w", metadataGraphNormalize, err)
	}
	for _, route := range []string{
		metadataRouteImage,
		metadataRouteDocument,
		metadataRouteMedia,
		metadataRouteArchive,
		metadataRouteGeneric,
	} {
		routeName := route
		if err := graph.AddLambdaNode(routeName, compose.InvokableLambda(func(ctx context.Context, state *MetadataGraphState) (*MetadataGraphState, error) {
			return metadataRouteNode(ctx, state, routeName)
		})); err != nil {
			return fmt.Errorf("add %s node: %w", routeName, err)
		}
	}
	if err := graph.AddLambdaNode(metadataGraphMerge, compose.InvokableLambda(metadataMergeEvidenceNode)); err != nil {
		return fmt.Errorf("add %s node: %w", metadataGraphMerge, err)
	}
	if err := graph.AddLambdaNode(metadataGraphAnalyze, compose.InvokableLambda(func(ctx context.Context, state *MetadataGraphState) (*MetadataGraphState, error) {
		return metadataAnalyzeNode(ctx, state, agent)
	})); err != nil {
		return fmt.Errorf("add %s node: %w", metadataGraphAnalyze, err)
	}
	if err := graph.AddLambdaNode(metadataGraphValidate, compose.InvokableLambda(metadataValidateNode)); err != nil {
		return fmt.Errorf("add %s node: %w", metadataGraphValidate, err)
	}
	if err := graph.AddLambdaNode(metadataGraphQuality, compose.InvokableLambda(metadataQualityGateNode)); err != nil {
		return fmt.Errorf("add %s node: %w", metadataGraphQuality, err)
	}
	if err := graph.AddLambdaNode(metadataGraphRecommend, compose.InvokableLambda(metadataRecommendationNode)); err != nil {
		return fmt.Errorf("add %s node: %w", metadataGraphRecommend, err)
	}
	for _, route := range []string{metadataRouteAutoApply, metadataRouteReview, metadataRouteLow} {
		routeName := route
		if err := graph.AddLambdaNode(routeName, compose.InvokableLambda(func(ctx context.Context, state *MetadataGraphState) (*MetadataGraphState, error) {
			return metadataQualityRouteNode(ctx, state, routeName)
		})); err != nil {
			return fmt.Errorf("add %s node: %w", routeName, err)
		}
	}
	return nil
}

func addMetadataGraphEdges(graph *compose.Graph[*MetadataGraphState, *MetadataGraphState]) error {
	if err := graph.AddEdge(compose.START, metadataGraphNormalize); err != nil {
		return fmt.Errorf("connect start to normalize: %w", err)
	}
	endNodes := map[string]bool{
		metadataRouteImage:    true,
		metadataRouteDocument: true,
		metadataRouteMedia:    true,
		metadataRouteArchive:  true,
		metadataRouteGeneric:  true,
	}
	if err := graph.AddBranch(metadataGraphNormalize, compose.NewGraphBranch(metadataRouteCondition, endNodes)); err != nil {
		return fmt.Errorf("add metadata route branch: %w", err)
	}
	for _, route := range []string{
		metadataRouteImage,
		metadataRouteDocument,
		metadataRouteMedia,
		metadataRouteArchive,
		metadataRouteGeneric,
	} {
		if err := graph.AddEdge(route, metadataGraphMerge); err != nil {
			return fmt.Errorf("connect %s to merge: %w", route, err)
		}
	}
	if err := graph.AddEdge(metadataGraphMerge, metadataGraphAnalyze); err != nil {
		return fmt.Errorf("connect merge to analyze: %w", err)
	}
	if err := graph.AddEdge(metadataGraphAnalyze, metadataGraphValidate); err != nil {
		return fmt.Errorf("connect analyze to validate: %w", err)
	}
	if err := graph.AddEdge(metadataGraphValidate, metadataGraphRecommend); err != nil {
		return fmt.Errorf("connect validate to recommendations: %w", err)
	}
	if err := graph.AddEdge(metadataGraphRecommend, metadataGraphQuality); err != nil {
		return fmt.Errorf("connect recommendations to quality: %w", err)
	}
	qualityEndNodes := map[string]bool{
		metadataRouteAutoApply: true,
		metadataRouteReview:    true,
		metadataRouteLow:       true,
	}
	if err := graph.AddBranch(metadataGraphQuality, compose.NewGraphBranch(metadataQualityRouteCondition, qualityEndNodes)); err != nil {
		return fmt.Errorf("add quality route branch: %w", err)
	}
	for _, route := range []string{metadataRouteAutoApply, metadataRouteReview, metadataRouteLow} {
		if err := graph.AddEdge(route, compose.END); err != nil {
			return fmt.Errorf("connect %s to end: %w", route, err)
		}
	}
	return nil
}

// Run 执行一次有界的文件分析 Graph。
func (g *MetadataAnalysisGraph) Run(ctx context.Context, input *AnalysisRequest, config *models.AIConfig, availableTools []tools.Tool) (*MetadataAnalysisResult, error) {
	if g == nil {
		return nil, fmt.Errorf("metadata graph is nil")
	}
	if g.buildErr != nil {
		return nil, g.buildErr
	}
	if g.runnable == nil {
		return nil, fmt.Errorf("metadata graph is not compiled")
	}
	if input == nil {
		return nil, fmt.Errorf("metadata graph input is nil")
	}
	if config == nil {
		return nil, fmt.Errorf("metadata graph ai config is nil")
	}

	state, err := g.runnable.Invoke(ctx, &MetadataGraphState{
		Request:        input,
		AIConfig:       config,
		AvailableTools: append([]tools.Tool(nil), availableTools...),
		StageHistory:   make([]string, 0, 8),
	})
	if err != nil {
		return nil, err
	}
	if state == nil || state.Result == nil {
		return nil, fmt.Errorf("metadata graph returned empty result")
	}
	if state.Result.Trace == nil {
		state.Result.Trace = &MetadataTrace{}
	}
	state.Result.Trace.CurrentStage = state.CurrentStage
	state.Result.Trace.StageHistory = append([]string(nil), state.StageHistory...)
	state.Result.Trace.Errors = append([]string(nil), state.Errors...)
	state.Result.Trace.Confidence = state.Confidence
	state.Result.Trace.NeedsReview = state.NeedsReview
	state.Result.Trace.QualityStatus = state.QualityStatus
	return state.Result, nil
}

func metadataNormalizeNode(_ context.Context, state *MetadataGraphState) (*MetadataGraphState, error) {
	if state == nil || state.Request == nil {
		return nil, fmt.Errorf("normalize input state is nil")
	}
	metadataGraphStage(state, metadataGraphNormalize)
	state.Route = metadataRouteForType(state.Request.FileType, state.Request.FileName)
	return state, nil
}

func metadataRouteCondition(_ context.Context, state *MetadataGraphState) (string, error) {
	if state == nil || state.Route == "" {
		return metadataRouteGeneric, nil
	}
	return state.Route, nil
}

func metadataRouteNode(_ context.Context, state *MetadataGraphState, route string) (*MetadataGraphState, error) {
	if state == nil {
		return nil, fmt.Errorf("route state is nil")
	}
	metadataGraphStage(state, route)
	return state, nil
}

func metadataMergeEvidenceNode(_ context.Context, state *MetadataGraphState) (*MetadataGraphState, error) {
	if state == nil || state.Request == nil {
		return nil, fmt.Errorf("merge evidence state is nil")
	}
	// Phase 2 已经完成本地证据采集，这个节点只负责为后续 Agent 统一证据入口。
	metadataGraphStage(state, metadataGraphMerge)
	return state, nil
}

func metadataAnalyzeNode(ctx context.Context, state *MetadataGraphState, agent MetadataAgent) (*MetadataGraphState, error) {
	if state == nil || state.Request == nil || state.AIConfig == nil {
		return nil, fmt.Errorf("analyze agent state is incomplete")
	}
	metadataGraphStage(state, metadataGraphAnalyze)
	result, err := agent.Analyze(ctx, state.Request, state.AIConfig, state.AvailableTools)
	if err != nil {
		state.Errors = append(state.Errors, err.Error())
		return nil, fmt.Errorf("%s: %w", metadataGraphAnalyze, err)
	}
	state.Result = result
	return state, nil
}

func metadataValidateNode(_ context.Context, state *MetadataGraphState) (*MetadataGraphState, error) {
	if state == nil || state.Result == nil || state.Result.Analysis == nil {
		return nil, fmt.Errorf("%s: metadata result is incomplete", metadataGraphValidate)
	}
	metadataGraphStage(state, metadataGraphValidate)
	state.Result.Analysis.Tags = normalizeMetadataTags(state.Result.Analysis.Tags)
	state.Result.Analysis.Description = strings.TrimSpace(state.Result.Analysis.Description)
	if len(state.Result.Analysis.Tags) == 0 && state.Result.Analysis.Description == "" {
		return nil, fmt.Errorf("%s: metadata analysis is empty", metadataGraphValidate)
	}
	return state, nil
}

// metadataQualityGateNode 根据证据和结果完整度计算门禁状态，避免完全信任模型自报的置信度。
func metadataQualityGateNode(_ context.Context, state *MetadataGraphState) (*MetadataGraphState, error) {
	if state == nil || state.Request == nil || state.Result == nil || state.Result.Analysis == nil {
		return nil, fmt.Errorf("%s: quality input is incomplete", metadataGraphQuality)
	}
	metadataGraphStage(state, metadataGraphQuality)

	confidence := 0.35
	if len(state.Result.Analysis.Tags) > 0 {
		confidence += 0.20
	}
	if strings.TrimSpace(state.Result.Analysis.Description) != "" {
		confidence += 0.20
	}
	if len(state.Request.Evidence) > 0 {
		confidence += 0.15
	}
	if len(state.Request.Evidence) >= 2 {
		confidence += 0.10
	}
	if hasEvidenceSource(state.Request.Evidence, "local:file_metadata") {
		confidence += 0.05
	}
	if strings.TrimSpace(state.Request.UserKeywords) != "" || len(state.Request.UserTags) > 0 || strings.TrimSpace(state.Request.UserDescription) != "" {
		confidence += 0.05
	}
	if confidence > 1 {
		confidence = 1
	}

	state.Confidence = confidence
	switch {
	case confidence >= metadataAutoApplyThreshold:
		state.QualityStatus = "auto_apply"
		state.NeedsReview = false
	case confidence >= metadataReviewThreshold:
		state.QualityStatus = "needs_review"
		state.NeedsReview = true
	default:
		state.QualityStatus = "low_confidence"
		state.NeedsReview = true
	}
	return state, nil
}

// metadataRecommendationNode 只生成可选提示，不改变主标签，也不访问本地搜索能力。
func metadataRecommendationNode(_ context.Context, state *MetadataGraphState) (*MetadataGraphState, error) {
	if state == nil || state.Request == nil || state.Result == nil || state.Result.Analysis == nil {
		return nil, fmt.Errorf("%s: recommendation input is incomplete", metadataGraphRecommend)
	}
	metadataGraphStage(state, metadataGraphRecommend)

	primaryTags := make(map[string]struct{}, len(state.Result.Analysis.Tags))
	for _, tag := range state.Result.Analysis.Tags {
		primaryTags[strings.ToLower(strings.TrimSpace(tag))] = struct{}{}
	}
	related := make([]string, 0, 5)
	seen := make(map[string]struct{}, 5)
	for _, tag := range state.Request.UserTags {
		appendRelatedTag(&related, seen, primaryTags, tag)
	}
	for _, keyword := range strings.FieldsFunc(state.Request.UserKeywords, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == ' ' || r == '\t' || r == '\n'
	}) {
		appendRelatedTag(&related, seen, primaryTags, keyword)
	}
	if len(related) > 0 {
		state.Result.Analysis.RelatedTags = related
		state.Result.Analysis.RecommendationReasons = append(state.Result.Analysis.RecommendationReasons, "复用用户已提供的标签和关键词")
	}
	if len(state.Request.Evidence) > 0 {
		state.Result.Analysis.RecommendationReasons = append(state.Result.Analysis.RecommendationReasons, "已有本地证据，可在后续接入本地库后生成合集推荐")
	}
	return state, nil
}

func appendRelatedTag(related *[]string, seen, primary map[string]struct{}, raw string) {
	tag := strings.ToLower(strings.TrimSpace(raw))
	if tag == "" || len(*related) >= 5 {
		return
	}
	if _, exists := primary[tag]; exists {
		return
	}
	if _, exists := seen[tag]; exists {
		return
	}
	seen[tag] = struct{}{}
	*related = append(*related, tag)
}

func metadataQualityRouteCondition(_ context.Context, state *MetadataGraphState) (string, error) {
	if state == nil {
		return metadataRouteLow, nil
	}
	switch state.QualityStatus {
	case "auto_apply":
		return metadataRouteAutoApply, nil
	case "needs_review":
		return metadataRouteReview, nil
	default:
		return metadataRouteLow, nil
	}
}

func metadataQualityRouteNode(_ context.Context, state *MetadataGraphState, route string) (*MetadataGraphState, error) {
	if state == nil {
		return nil, fmt.Errorf("quality route state is nil")
	}
	metadataGraphStage(state, route)
	return state, nil
}

func hasEvidenceSource(evidence []tools.EvidenceItem, source string) bool {
	for _, item := range evidence {
		if item.Source == source {
			return true
		}
	}
	return false
}

func metadataGraphStage(state *MetadataGraphState, stage string) {
	state.CurrentStage = stage
	state.StageHistory = append(state.StageHistory, stage)
}

func metadataRouteForType(fileType, fileName string) string {
	typeName := strings.ToLower(strings.TrimSpace(fileType))
	if typeName == "" {
		name := strings.ToLower(strings.TrimSpace(fileName))
		switch {
		case strings.HasSuffix(name, ".jpg"), strings.HasSuffix(name, ".jpeg"), strings.HasSuffix(name, ".png"), strings.HasSuffix(name, ".gif"), strings.HasSuffix(name, ".webp"):
			typeName = "image"
		case strings.HasSuffix(name, ".mp4"), strings.HasSuffix(name, ".mkv"), strings.HasSuffix(name, ".avi"), strings.HasSuffix(name, ".mov"):
			typeName = "video"
		case strings.HasSuffix(name, ".mp3"), strings.HasSuffix(name, ".flac"), strings.HasSuffix(name, ".wav"), strings.HasSuffix(name, ".m4a"):
			typeName = "music"
		}
	}
	switch typeName {
	case "image", "photo", "picture":
		return metadataRouteImage
	case "document", "text", "pdf", "word", "spreadsheet", "presentation":
		return metadataRouteDocument
	case "video", "music", "audio", "media":
		return metadataRouteMedia
	case "archive", "zip", "rar", "7z", "tar":
		return metadataRouteArchive
	default:
		return metadataRouteGeneric
	}
}

func normalizeMetadataTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}
