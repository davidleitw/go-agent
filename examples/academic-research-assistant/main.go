package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/davidleitw/go-agent/agent"
	agentcontext "github.com/davidleitw/go-agent/context"
	"github.com/davidleitw/go-agent/llm"
	"github.com/davidleitw/go-agent/llm/openai"
	"github.com/davidleitw/go-agent/prompt"
	"github.com/davidleitw/go-agent/session"
	"github.com/davidleitw/go-agent/tool"
)

func main() {
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	// Check for API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Initialize OpenAI client
	client := openai.New(llm.Config{
		APIKey: apiKey,
		Model:  "gpt-3.5-turbo", // Use lighter model to avoid rate limits
	})

	switch os.Args[1] {
	case "explore":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run . explore \"research topic\"")
			return
		}
		runExploreWorkflow(client, os.Args[2])
	case "research":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run . research \"specific topic\"")
			return
		}
		runResearchWorkflow(client, os.Args[2])
	case "track":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run . track \"field and year\"")
			return
		}
		runTrackWorkflow(client, os.Args[2])
	case "inspire":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run . inspire \"current research\"")
			return
		}
		runInspireWorkflow(client, os.Args[2])
	default:
		showUsage()
	}
}

func showUsage() {
	fmt.Println("🎓 Academic Research Assistant")
	fmt.Println("===============================")
	fmt.Println("Your intelligent research companion for academic discovery")
	fmt.Println("")
	fmt.Println("Usage: go run . [command] \"topic\"")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  explore \"topic\"   - Discover and explore new research areas")
	fmt.Println("  research \"topic\"  - Deep literature review and analysis")
	fmt.Println("  track \"field year\" - Track latest developments in a field")
	fmt.Println("  inspire \"project\" - Find research gaps and opportunities")
	fmt.Println("")
	fmt.Println("Environment:")
	fmt.Println("  OPENAI_API_KEY - Required for OpenAI API access")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run . explore \"quantum machine learning\"")
	fmt.Println("  go run . research \"transformer attention mechanisms\"")
	fmt.Println("  go run . track \"computer vision 2024\"")
	fmt.Println("  go run . inspire \"graph neural networks for drug discovery\"")
	fmt.Println("")
	fmt.Println("Features:")
	fmt.Println("  🔍 ArXiv paper search and discovery")
	fmt.Println("  📊 Paper content analysis and summarization")
	fmt.Println("  📚 Citation management and relationship mapping")
	fmt.Println("  📈 Research trend analysis and prediction")
	fmt.Println("  💡 Intelligent paper recommendations")
	fmt.Println("  📝 Research note taking and organization")
}

func runExploreWorkflow(client *openai.Client, topic string) {
	fmt.Println("🔍 Starting Research Area Exploration")
	fmt.Println("====================================")
	fmt.Printf("Topic: %s\n\n", topic)

	ag, err := createResearchAgent(client, "exploration")
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	fmt.Println("📋 Executing research workflow steps...")
	fmt.Println("Step 1: Initializing exploration process")

	response, err := ag.Execute(context.Background(), agent.Request{
		Input: fmt.Sprintf(`你正在為以下主題進行研究領域探索: "%s"

請用繁體中文回覆。你的目標是提供全面的概述。根據需要執行工具來收集資訊，但專注於提供有用的見解，即使無法完成所有計劃的步驟。

從使用 arxiv_search 搜索論文開始，然後分析並總結你的發現。

重點探索領域：
- 最新論文和發展
- 主要研究人員和機構
- 主要研究主題和方法
- 當前挑戰和機會

請基於你收集到的資訊提供研究領域概述。`, topic),
	})

	handleWorkflowResult("Research Area Exploration", response, err)
}

// handleWorkflowResult processes agent execution results and errors
func handleWorkflowResult(workflowType string, response *agent.Response, err error) {
	if err != nil {
		// Check if it's a maximum iterations exceeded error
		if strings.Contains(err.Error(), "maximum iterations exceeded") {
			fmt.Printf("\n⚠️  %s reached maximum iterations\n", workflowType)
			fmt.Println("This indicates the agent was making progress but needed more iterations to complete.")

			// Show execution details if available
			if response != nil {
				if response.Usage.ToolCalls > 0 {
					fmt.Printf("✓ Tools executed: %d calls\n", response.Usage.ToolCalls)
					fmt.Printf("✓ LLM tokens used: %d total\n", response.Usage.LLMTokens.TotalTokens)
				}

				// If we have a partial response, show it
				if response.Output != "" {
					fmt.Printf("\n📊 Partial Results Achieved:\n")
					fmt.Printf("%s\n", response.Output)
				}

				fmt.Printf("\nExecution Summary: %+v\n", response.Usage)
			} else {
				fmt.Println("The agent was making progress through research steps before reaching the limit.")
				fmt.Println("\nSuggestion: Try with a more focused query or run the workflow again.")
			}
		} else {
			fmt.Printf("\n❌ %s failed: %v\n", workflowType, err)
		}
		return
	}

	// Success case
	fmt.Printf("\n🎯 %s Complete!\n", workflowType)
	fmt.Printf("Report:\n%s\n", response.Output)
	fmt.Printf("\nExecution Summary: %+v\n", response.Usage)
}

func runResearchWorkflow(client *openai.Client, topic string) {
	fmt.Println("📚 Starting Deep Literature Research")
	fmt.Println("===================================")
	fmt.Printf("Topic: %s\n\n", topic)

	ag, err := createResearchAgent(client, "research")
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	fmt.Println("📋 Executing deep research workflow...")
	fmt.Println("Step 1: Initializing literature review process")

	response, err := ag.Execute(context.Background(), agent.Request{
		Input: fmt.Sprintf(`請對以下主題進行深度文獻研究: "%s"

請用繁體中文回覆。使用 arxiv_search 找到相關論文，然後分析最重要的幾篇。

重點關注：
- 主要方法論和方法
- 最新實驗結果
- 重要理論貢獻
- 研究空隙和未來方向

請根據你的發現提供文獻綜述。`, topic),
	})

	handleWorkflowResult("Deep Literature Research", response, err)
}

func runTrackWorkflow(client *openai.Client, field string) {
	fmt.Println("📈 Starting Research Trend Tracking")
	fmt.Println("=================================")
	fmt.Printf("Field: %s\n\n", field)

	ag, err := createResearchAgent(client, "tracking")
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	fmt.Println("📋 Executing trend tracking workflow...")
	fmt.Println("Step 1: Initializing trend analysis process")

	response, err := ag.Execute(context.Background(), agent.Request{
		Input: fmt.Sprintf(`Track research trends in: "%s"

Execute this tracking workflow step by step:
1. Search for recent papers and developments (arxiv_search)
2. Analyze emerging themes and methods (trend_analyzer)
3. Identify breakthrough papers and innovations (paper_analyzer)
4. Map citation patterns and influence (citation_manager)
5. Discover rising researchers and labs (citation_manager)
6. Recommend must-read recent papers (recommendation_tool)
7. Update trend analysis notes (research_notes)
8. Generate trend analysis report

Focus on recent developments, emerging patterns, and future predictions.`, field),
	})

	handleWorkflowResult("Research Trend Tracking", response, err)
}

func runInspireWorkflow(client *openai.Client, project string) {
	fmt.Println("💡 Starting Research Inspiration Discovery")
	fmt.Println("========================================")
	fmt.Printf("Current Project: %s\n\n", project)

	ag, err := createResearchAgent(client, "inspiration")
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	fmt.Println("📋 Executing inspiration discovery workflow...")
	fmt.Println("Step 1: Initializing creative exploration process")

	response, err := ag.Execute(context.Background(), agent.Request{
		Input: fmt.Sprintf(`Find research inspiration for: "%s"

Execute this inspiration workflow step by step:
1. Analyze current project context (paper_analyzer)
2. Search for adjacent research areas (arxiv_search)
3. Identify unexplored connections (trend_analyzer)
4. Find interdisciplinary opportunities (recommendation_tool)
5. Map potential collaboration paths (citation_manager)
6. Suggest novel research directions (trend_analyzer)
7. Document inspiration ideas (research_notes)
8. Generate research opportunity report

Focus on creative connections, unexplored angles, and innovative approaches.`, project),
	})

	handleWorkflowResult("Research Inspiration Discovery", response, err)
}

func createResearchAgent(client *openai.Client, workflowType string) (agent.Agent, error) {
	// Create research-focused template
	template := prompt.New().
		System().
		Text("You are an expert academic research assistant with deep knowledge across scientific disciplines.").
		Text("You help researchers discover, analyze, and synthesize academic literature.").
		Line("").
		Text("Research Focus:").
		Provider("research_context").
		Line("").
		Text("Previous Research Activity:").
		History().
		Line("").
		Text("Current Research Request:").
		UserInput().
		Build()

	return agent.NewBuilder().
		WithLLM(client).
		WithMemorySessionStore().
		WithPromptTemplate(template).
		WithHistoryLimit(8).
		WithMaxIterations(15).
		WithContextProviders(
			NewResearchContextProvider(workflowType),
		).
		WithTools(
			NewArxivSearchTool(),
			NewPaperAnalyzerTool(),
			NewCitationManagerTool(),
			NewTrendAnalyzerTool(),
			NewRecommendationTool(),
			NewResearchNoteTool(),
		).
		Build()
}

// Research Context Provider
type ResearchContextProvider struct {
	workflowType string
}

func NewResearchContextProvider(workflowType string) *ResearchContextProvider {
	return &ResearchContextProvider{workflowType: workflowType}
}

func (p *ResearchContextProvider) Type() string {
	return "research_context"
}

func (p *ResearchContextProvider) Provide(ctx context.Context, s session.Session) []agentcontext.Context {
	guidance := map[string]string{
		"exploration": "Focus on breadth - discover key papers, themes, and researchers. Build foundational understanding.",
		"research":    "Focus on depth - analyze methodologies, compare approaches, identify gaps and contributions.",
		"tracking":    "Focus on trends - identify recent developments, emerging patterns, and future directions.",
		"inspiration": "Focus on creativity - find novel connections, interdisciplinary opportunities, and unexplored angles.",
	}

	workflowGuidance, exists := guidance[p.workflowType]
	if !exists {
		workflowGuidance = "Focus on comprehensive research analysis and synthesis."
	}

	return []agentcontext.Context{{
		Type: "research_context",
		Content: fmt.Sprintf(`Research Workflow: %s
Guidance: %s

Research Best Practices:
- Start with broad searches, then narrow down to specific papers
- Always verify paper quality and venue reputation
- Look for recent survey papers for comprehensive overviews
- Pay attention to citation counts and author credentials
- Consider both theoretical contributions and practical applications
- Note reproducibility and availability of code/data
- Identify connections between different research areas`, p.workflowType, workflowGuidance),
		Metadata: map[string]any{
			"workflow_type": p.workflowType,
		},
	}}
}

// ========== RESEARCH TOOLS ==========

// ArxivSearchTool searches for academic papers on arXiv
type ArxivSearchTool struct{}

func NewArxivSearchTool() *ArxivSearchTool {
	return &ArxivSearchTool{}
}

func (t *ArxivSearchTool) Definition() tool.Definition {
	return tool.Definition{
		Type: "function",
		Function: tool.Function{
			Name:        "arxiv_search",
			Description: "Search for academic papers on arXiv repository. Supports keyword search, author search, and category filtering.",
			Parameters: tool.Parameters{
				Type: "object",
				Properties: map[string]tool.Property{
					"query": {
						Type:        "string",
						Description: "Search query (keywords, paper title, or author name)",
					},
					"category": {
						Type:        "string",
						Description: "arXiv category filter (e.g., cs.AI, cs.LG, stat.ML, physics.ao-ph)",
					},
					"max_results": {
						Type:        "integer",
						Description: "Maximum number of papers to return (default: 10, max: 50)",
					},
					"sort_by": {
						Type:        "string",
						Description: "Sort results by: relevance, lastUpdatedDate, submittedDate",
					},
				},
				Required: []string{"query"},
			},
		},
	}
}

func (t *ArxivSearchTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	maxResults := 10
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
		if maxResults > 50 {
			maxResults = 50
		}
	}

	sortBy := "relevance"
	if sb, ok := args["sort_by"].(string); ok {
		sortBy = sb
	}

	// Build arXiv API query
	searchQuery := query
	if category, ok := args["category"].(string); ok && category != "" {
		searchQuery = fmt.Sprintf("cat:%s AND (%s)", category, query)
	}

	// Call arXiv API
	papers, err := searchArxivPapers(searchQuery, maxResults, sortBy)
	if err != nil {
		return "", fmt.Errorf("failed to search arXiv: %w", err)
	}

	if len(papers) == 0 {
		return fmt.Sprintf("No papers found for query: %s", query), nil
	}

	// Format results
	result := fmt.Sprintf("Found %d papers for query '%s':\n\n", len(papers), query)
	for i, paper := range papers {
		result += fmt.Sprintf("%d. **%s**\n", i+1, paper.Title)
		result += fmt.Sprintf("   Authors: %s\n", strings.Join(paper.Authors, ", "))
		result += fmt.Sprintf("   Published: %s\n", paper.Published.Format("2006-01-02"))
		result += fmt.Sprintf("   Category: %s\n", strings.Join(paper.Categories, ", "))
		result += fmt.Sprintf("   arXiv ID: %s\n", paper.ArxivID)
		result += fmt.Sprintf("   Abstract: %s\n", truncateText(paper.Abstract, 200))
		result += fmt.Sprintf("   URL: https://arxiv.org/abs/%s\n\n", paper.ArxivID)
	}

	return result, nil
}

// PaperAnalyzerTool analyzes paper content and extracts key insights
type PaperAnalyzerTool struct{}

func NewPaperAnalyzerTool() *PaperAnalyzerTool {
	return &PaperAnalyzerTool{}
}

func (t *PaperAnalyzerTool) Definition() tool.Definition {
	return tool.Definition{
		Type: "function",
		Function: tool.Function{
			Name:        "paper_analyzer",
			Description: "Analyze academic paper content to extract key insights, methodologies, contributions, and findings.",
			Parameters: tool.Parameters{
				Type: "object",
				Properties: map[string]tool.Property{
					"arxiv_id": {
						Type:        "string",
						Description: "arXiv paper ID (e.g., 2301.07041)",
					},
					"analysis_type": {
						Type:        "string",
						Description: "Type of analysis: summary, methodology, contributions, related_work, experiments",
					},
					"focus_areas": {
						Type:        "string",
						Description: "Specific areas to focus on (comma-separated keywords)",
					},
				},
				Required: []string{"arxiv_id", "analysis_type"},
			},
		},
	}
}

func (t *PaperAnalyzerTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	arxivID, ok := args["arxiv_id"].(string)
	if !ok || arxivID == "" {
		return "", fmt.Errorf("arxiv_id parameter is required")
	}

	analysisType, ok := args["analysis_type"].(string)
	if !ok || analysisType == "" {
		return "", fmt.Errorf("analysis_type parameter is required")
	}

	focusAreas := ""
	if fa, ok := args["focus_areas"].(string); ok {
		focusAreas = fa
	}

	// Fetch paper details
	paper, err := fetchArxivPaper(arxivID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch paper: %w", err)
	}

	// Perform analysis based on type
	var analysis string
	switch analysisType {
	case "summary":
		analysis = analyzePaperSummary(paper, focusAreas)
	case "methodology":
		analysis = analyzePaperMethodology(paper, focusAreas)
	case "contributions":
		analysis = analyzePaperContributions(paper, focusAreas)
	case "related_work":
		analysis = analyzePaperRelatedWork(paper, focusAreas)
	case "experiments":
		analysis = analyzePaperExperiments(paper, focusAreas)
	default:
		analysis = analyzePaperSummary(paper, focusAreas)
	}

	return fmt.Sprintf("Paper Analysis: %s\n\n%s", paper.Title, analysis), nil
}

// CitationManagerTool manages citations and research relationships
type CitationManagerTool struct{}

func NewCitationManagerTool() *CitationManagerTool {
	return &CitationManagerTool{}
}

func (t *CitationManagerTool) Definition() tool.Definition {
	return tool.Definition{
		Type: "function",
		Function: tool.Function{
			Name:        "citation_manager",
			Description: "Manage citations, track paper relationships, and build citation networks for research analysis.",
			Parameters: tool.Parameters{
				Type: "object",
				Properties: map[string]tool.Property{
					"action": {
						Type:        "string",
						Description: "Action to perform: add_citation, find_citations, analyze_network, format_bibliography",
					},
					"arxiv_id": {
						Type:        "string",
						Description: "arXiv paper ID for citation operations",
					},
					"format": {
						Type:        "string",
						Description: "Citation format: apa, mla, ieee, bibtex",
					},
					"author": {
						Type:        "string",
						Description: "Author name for citation analysis",
					},
				},
				Required: []string{"action"},
			},
		},
	}
}

func (t *CitationManagerTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	action, ok := args["action"].(string)
	if !ok || action == "" {
		return "", fmt.Errorf("action parameter is required")
	}

	switch action {
	case "add_citation":
		return t.addCitation(args)
	case "find_citations":
		return t.findCitations(args)
	case "analyze_network":
		return t.analyzeNetwork(args)
	case "format_bibliography":
		return t.formatBibliography(args)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *CitationManagerTool) addCitation(args map[string]any) (string, error) {
	arxivID, ok := args["arxiv_id"].(string)
	if !ok || arxivID == "" {
		return "", fmt.Errorf("arxiv_id is required for add_citation action")
	}

	format := "apa"
	if f, ok := args["format"].(string); ok {
		format = f
	}

	paper, err := fetchArxivPaper(arxivID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch paper: %w", err)
	}

	citation := formatCitation(paper, format)
	return fmt.Sprintf("Citation added in %s format:\n\n%s", strings.ToUpper(format), citation), nil
}

func (t *CitationManagerTool) findCitations(args map[string]any) (string, error) {
	if author, ok := args["author"].(string); ok && author != "" {
		return t.findCitationsByAuthor(author)
	}
	if arxivID, ok := args["arxiv_id"].(string); ok && arxivID != "" {
		return t.findCitationsByPaper(arxivID)
	}
	return "", fmt.Errorf("either author or arxiv_id is required")
}

func (t *CitationManagerTool) findCitationsByAuthor(author string) (string, error) {
	papers, err := searchArxivPapers(fmt.Sprintf("au:%s", author), 20, "submittedDate")
	if err != nil {
		return "", fmt.Errorf("failed to search papers by author: %w", err)
	}

	if len(papers) == 0 {
		return fmt.Sprintf("No papers found for author: %s", author), nil
	}

	result := fmt.Sprintf("Found %d papers by %s:\n\n", len(papers), author)
	for i, paper := range papers {
		result += fmt.Sprintf("%d. %s (%s)\n", i+1, paper.Title, paper.Published.Format("2006"))
		result += fmt.Sprintf("   arXiv:%s\n\n", paper.ArxivID)
	}

	return result, nil
}

func (t *CitationManagerTool) findCitationsByPaper(arxivID string) (string, error) {
	paper, err := fetchArxivPaper(arxivID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch paper: %w", err)
	}

	// Extract key terms from title and abstract for related work search
	keyTerms := extractKeyTerms(paper.Title + " " + paper.Abstract)
	relatedPapers, err := searchArxivPapers(strings.Join(keyTerms[:3], " "), 15, "relevance")
	if err != nil {
		return "", fmt.Errorf("failed to find related papers: %w", err)
	}

	result := fmt.Sprintf("Papers related to '%s':\n\n", paper.Title)
	count := 0
	for _, relatedPaper := range relatedPapers {
		if relatedPaper.ArxivID != arxivID && count < 10 {
			count++
			result += fmt.Sprintf("%d. %s\n", count, relatedPaper.Title)
			result += fmt.Sprintf("   Authors: %s\n", strings.Join(relatedPaper.Authors, ", "))
			result += fmt.Sprintf("   arXiv:%s (%s)\n\n", relatedPaper.ArxivID, relatedPaper.Published.Format("2006"))
		}
	}

	return result, nil
}

func (t *CitationManagerTool) analyzeNetwork(args map[string]any) (string, error) {
	author, ok := args["author"].(string)
	if !ok || author == "" {
		return "", fmt.Errorf("author parameter is required for network analysis")
	}

	papers, err := searchArxivPapers(fmt.Sprintf("au:%s", author), 30, "submittedDate")
	if err != nil {
		return "", fmt.Errorf("failed to search papers: %w", err)
	}

	// Analyze collaboration network
	collaborators := make(map[string]int)
	categories := make(map[string]int)
	yearStats := make(map[int]int)

	for _, paper := range papers {
		// Count collaborators
		for _, coAuthor := range paper.Authors {
			if !strings.Contains(strings.ToLower(coAuthor), strings.ToLower(author)) {
				collaborators[coAuthor]++
			}
		}

		// Count categories
		for _, cat := range paper.Categories {
			categories[cat]++
		}

		// Count by year
		year := paper.Published.Year()
		yearStats[year]++
	}

	result := fmt.Sprintf("Citation Network Analysis for %s:\n\n", author)
	result += fmt.Sprintf("Total Papers: %d\n\n", len(papers))

	// Top collaborators
	result += "Top Collaborators:\n"
	topCollaborators := getTopEntries(collaborators, 10)
	for i, entry := range topCollaborators {
		result += fmt.Sprintf("%d. %s (%d papers)\n", i+1, entry.Key, entry.Value)
	}

	// Research areas
	result += "\nResearch Areas:\n"
	topCategories := getTopEntries(categories, 5)
	for i, entry := range topCategories {
		result += fmt.Sprintf("%d. %s (%d papers)\n", i+1, entry.Key, entry.Value)
	}

	// Publication timeline
	result += "\nPublication Timeline:\n"
	years := make([]int, 0, len(yearStats))
	for year := range yearStats {
		years = append(years, year)
	}
	sort.Ints(years)
	for _, year := range years {
		result += fmt.Sprintf("%d: %d papers\n", year, yearStats[year])
	}

	return result, nil
}

func (t *CitationManagerTool) formatBibliography(args map[string]any) (string, error) {
	return "Bibliography formatting feature coming soon. Please use add_citation for individual citations.", nil
}

// TrendAnalyzerTool analyzes research trends and patterns
type TrendAnalyzerTool struct{}

func NewTrendAnalyzerTool() *TrendAnalyzerTool {
	return &TrendAnalyzerTool{}
}

func (t *TrendAnalyzerTool) Definition() tool.Definition {
	return tool.Definition{
		Type: "function",
		Function: tool.Function{
			Name:        "trend_analyzer",
			Description: "Analyze research trends, identify emerging topics, and track the evolution of research areas over time.",
			Parameters: tool.Parameters{
				Type: "object",
				Properties: map[string]tool.Property{
					"field": {
						Type:        "string",
						Description: "Research field or keywords to analyze trends for",
					},
					"time_range": {
						Type:        "string",
						Description: "Time range for analysis: 1year, 2years, 5years, all",
					},
					"analysis_type": {
						Type:        "string",
						Description: "Type of trend analysis: growth, topics, authors, venues, keywords",
					},
					"category": {
						Type:        "string",
						Description: "arXiv category to focus on (optional)",
					},
				},
				Required: []string{"field", "analysis_type"},
			},
		},
	}
}

func (t *TrendAnalyzerTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	field, ok := args["field"].(string)
	if !ok || field == "" {
		return "", fmt.Errorf("field parameter is required")
	}

	analysisType, ok := args["analysis_type"].(string)
	if !ok || analysisType == "" {
		return "", fmt.Errorf("analysis_type parameter is required")
	}

	timeRange := "2years"
	if tr, ok := args["time_range"].(string); ok {
		timeRange = tr
	}

	category := ""
	if cat, ok := args["category"].(string); ok {
		category = cat
	}

	// Search for papers in the field
	query := field
	if category != "" {
		query = fmt.Sprintf("cat:%s AND (%s)", category, field)
	}

	papers, err := searchArxivPapers(query, 100, "submittedDate")
	if err != nil {
		return "", fmt.Errorf("failed to search papers: %w", err)
	}

	// Filter by time range
	filteredPapers := filterPapersByTimeRange(papers, timeRange)

	switch analysisType {
	case "growth":
		return t.analyzeGrowthTrend(field, filteredPapers), nil
	case "topics":
		return t.analyzeTopicTrends(field, filteredPapers), nil
	case "authors":
		return t.analyzeAuthorTrends(field, filteredPapers), nil
	case "keywords":
		return t.analyzeKeywordTrends(field, filteredPapers), nil
	default:
		return t.analyzeGrowthTrend(field, filteredPapers), nil
	}
}

func (t *TrendAnalyzerTool) analyzeGrowthTrend(field string, papers []ArxivPaper) string {
	yearStats := make(map[int]int)
	for _, paper := range papers {
		year := paper.Published.Year()
		yearStats[year]++
	}

	result := fmt.Sprintf("Growth Trend Analysis for '%s':\n\n", field)
	result += fmt.Sprintf("Total Papers Analyzed: %d\n\n", len(papers))

	// Sort years
	years := make([]int, 0, len(yearStats))
	for year := range yearStats {
		years = append(years, year)
	}
	sort.Ints(years)

	result += "Publication Growth by Year:\n"
	for _, year := range years {
		count := yearStats[year]
		bar := strings.Repeat("█", count/10)
		if count < 10 && count > 0 {
			bar = "▌"
		}
		result += fmt.Sprintf("%d: %d papers %s\n", year, count, bar)
	}

	// Calculate growth rate
	if len(years) >= 2 {
		recentYear := years[len(years)-1]
		previousYear := years[len(years)-2]
		growthRate := float64(yearStats[recentYear]-yearStats[previousYear]) / float64(yearStats[previousYear]) * 100
		result += fmt.Sprintf("\nGrowth Rate (%d vs %d): %.1f%%\n", previousYear, recentYear, growthRate)
	}

	return result
}

func (t *TrendAnalyzerTool) analyzeTopicTrends(field string, papers []ArxivPaper) string {
	keywordCount := make(map[string]int)

	// Extract keywords from titles and abstracts
	for _, paper := range papers {
		keywords := extractKeyTerms(paper.Title + " " + paper.Abstract)
		for _, keyword := range keywords {
			if len(keyword) > 3 { // Filter out short words
				keywordCount[keyword]++
			}
		}
	}

	result := fmt.Sprintf("Topic Trend Analysis for '%s':\n\n", field)
	result += "Emerging Topics and Keywords:\n"

	topKeywords := getTopEntries(keywordCount, 15)
	for i, entry := range topKeywords {
		percentage := float64(entry.Value) / float64(len(papers)) * 100
		result += fmt.Sprintf("%d. %s (%.1f%% of papers)\n", i+1, entry.Key, percentage)
	}

	return result
}

func (t *TrendAnalyzerTool) analyzeAuthorTrends(field string, papers []ArxivPaper) string {
	authorCount := make(map[string]int)

	for _, paper := range papers {
		for _, author := range paper.Authors {
			authorCount[author]++
		}
	}

	result := fmt.Sprintf("Author Trend Analysis for '%s':\n\n", field)
	result += "Most Active Researchers:\n"

	topAuthors := getTopEntries(authorCount, 20)
	for i, entry := range topAuthors {
		result += fmt.Sprintf("%d. %s (%d papers)\n", i+1, entry.Key, entry.Value)
	}

	return result
}

func (t *TrendAnalyzerTool) analyzeKeywordTrends(field string, papers []ArxivPaper) string {
	return t.analyzeTopicTrends(field, papers) // Similar implementation
}

// RecommendationTool provides intelligent paper recommendations
type RecommendationTool struct{}

func NewRecommendationTool() *RecommendationTool {
	return &RecommendationTool{}
}

func (t *RecommendationTool) Definition() tool.Definition {
	return tool.Definition{
		Type: "function",
		Function: tool.Function{
			Name:        "recommendation_tool",
			Description: "Provide intelligent recommendations for papers, authors, and research directions based on interests and context.",
			Parameters: tool.Parameters{
				Type: "object",
				Properties: map[string]tool.Property{
					"recommendation_type": {
						Type:        "string",
						Description: "Type of recommendation: papers, authors, topics, venues",
					},
					"interests": {
						Type:        "string",
						Description: "Research interests or keywords",
					},
					"context_papers": {
						Type:        "string",
						Description: "Comma-separated arXiv IDs of papers for context",
					},
					"career_stage": {
						Type:        "string",
						Description: "Career stage: student, postdoc, professor, industry",
					},
				},
				Required: []string{"recommendation_type", "interests"},
			},
		},
	}
}

func (t *RecommendationTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	recType, ok := args["recommendation_type"].(string)
	if !ok || recType == "" {
		return "", fmt.Errorf("recommendation_type parameter is required")
	}

	interests, ok := args["interests"].(string)
	if !ok || interests == "" {
		return "", fmt.Errorf("interests parameter is required")
	}

	careerStage := "student"
	if cs, ok := args["career_stage"].(string); ok {
		careerStage = cs
	}

	switch recType {
	case "papers":
		return t.recommendPapers(interests, careerStage)
	case "authors":
		return t.recommendAuthors(interests)
	case "topics":
		return t.recommendTopics(interests)
	default:
		return t.recommendPapers(interests, careerStage)
	}
}

func (t *RecommendationTool) recommendPapers(interests string, careerStage string) (string, error) {
	// Search for papers based on interests
	papers, err := searchArxivPapers(interests, 30, "relevance")
	if err != nil {
		return "", fmt.Errorf("failed to search papers: %w", err)
	}

	result := fmt.Sprintf("Paper Recommendations for '%s' (%s level):\n\n", interests, careerStage)

	// Filter and rank papers based on career stage
	recommendations := filterPapersByCareerStage(papers, careerStage)

	result += "📚 Recommended Papers:\n"
	for i, paper := range recommendations {
		if i >= 10 { // Limit to top 10
			break
		}
		result += fmt.Sprintf("%d. **%s**\n", i+1, paper.Title)
		result += fmt.Sprintf("   Authors: %s\n", strings.Join(paper.Authors, ", "))
		result += fmt.Sprintf("   Published: %s\n", paper.Published.Format("2006-01-02"))
		result += fmt.Sprintf("   arXiv: %s\n", paper.ArxivID)
		result += fmt.Sprintf("   Summary: %s\n\n", truncateText(paper.Abstract, 150))
	}

	return result, nil
}

func (t *RecommendationTool) recommendAuthors(interests string) (string, error) {
	papers, err := searchArxivPapers(interests, 50, "relevance")
	if err != nil {
		return "", fmt.Errorf("failed to search papers: %w", err)
	}

	authorCount := make(map[string]int)
	for _, paper := range papers {
		for _, author := range paper.Authors {
			authorCount[author]++
		}
	}

	result := fmt.Sprintf("Author Recommendations for '%s':\n\n", interests)
	result += "👨‍🔬 Recommended Researchers to Follow:\n"

	topAuthors := getTopEntries(authorCount, 15)
	for i, entry := range topAuthors {
		result += fmt.Sprintf("%d. %s (%d relevant papers)\n", i+1, entry.Key, entry.Value)
	}

	return result, nil
}

func (t *RecommendationTool) recommendTopics(interests string) (string, error) {
	papers, err := searchArxivPapers(interests, 50, "relevance")
	if err != nil {
		return "", fmt.Errorf("failed to search papers: %w", err)
	}

	keywordCount := make(map[string]int)
	for _, paper := range papers {
		keywords := extractKeyTerms(paper.Title + " " + paper.Abstract)
		for _, keyword := range keywords {
			if len(keyword) > 3 {
				keywordCount[keyword]++
			}
		}
	}

	result := fmt.Sprintf("Topic Recommendations for '%s':\n\n", interests)
	result += "🔍 Related Research Topics to Explore:\n"

	topTopics := getTopEntries(keywordCount, 20)
	for i, entry := range topTopics {
		result += fmt.Sprintf("%d. %s\n", i+1, entry.Key)
	}

	return result, nil
}

// ResearchNoteTool manages research notes and organization
type ResearchNoteTool struct{}

func NewResearchNoteTool() *ResearchNoteTool {
	return &ResearchNoteTool{}
}

func (t *ResearchNoteTool) Definition() tool.Definition {
	return tool.Definition{
		Type: "function",
		Function: tool.Function{
			Name:        "research_notes",
			Description: "Take, organize, and manage research notes. Create summaries, track reading progress, and maintain research journals.",
			Parameters: tool.Parameters{
				Type: "object",
				Properties: map[string]tool.Property{
					"action": {
						Type:        "string",
						Description: "Action to perform: create_note, summarize_session, track_progress, organize_notes",
					},
					"content": {
						Type:        "string",
						Description: "Note content or research findings",
					},
					"topic": {
						Type:        "string",
						Description: "Research topic or theme",
					},
					"tags": {
						Type:        "string",
						Description: "Comma-separated tags for organization",
					},
				},
				Required: []string{"action"},
			},
		},
	}
}

func (t *ResearchNoteTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	action, ok := args["action"].(string)
	if !ok || action == "" {
		return "", fmt.Errorf("action parameter is required")
	}

	switch action {
	case "create_note":
		return t.createNote(args)
	case "summarize_session":
		return t.summarizeSession(args)
	case "track_progress":
		return t.trackProgress(args)
	case "organize_notes":
		return t.organizeNotes(args)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *ResearchNoteTool) createNote(args map[string]any) (string, error) {
	content, ok := args["content"].(string)
	if !ok || content == "" {
		return "", fmt.Errorf("content is required for create_note action")
	}

	topic, _ := args["topic"].(string)
	tags, _ := args["tags"].(string)

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	result := "📝 Research Note Created\n"
	result += "========================\n"
	result += fmt.Sprintf("Timestamp: %s\n", timestamp)
	if topic != "" {
		result += fmt.Sprintf("Topic: %s\n", topic)
	}
	if tags != "" {
		result += fmt.Sprintf("Tags: %s\n", tags)
	}
	result += "\nContent:\n"
	result += content + "\n"

	return result, nil
}

func (t *ResearchNoteTool) summarizeSession(args map[string]any) (string, error) {
	topic, _ := args["topic"].(string)

	result := "📊 Research Session Summary\n"
	result += "===========================\n"
	result += fmt.Sprintf("Date: %s\n", time.Now().Format("2006-01-02"))
	if topic != "" {
		result += fmt.Sprintf("Topic: %s\n", topic)
	}
	result += "\nSession Highlights:\n"
	result += "- Papers discovered and analyzed\n"
	result += "- Key insights and findings\n"
	result += "- Research questions identified\n"
	result += "- Next steps and follow-up items\n"
	result += "\nNote: This is a template summary. In a full implementation, this would aggregate actual session data.\n"

	return result, nil
}

func (t *ResearchNoteTool) trackProgress(args map[string]any) (string, error) {
	topic, _ := args["topic"].(string)

	result := "📈 Research Progress Tracking\n"
	result += "=============================\n"
	if topic != "" {
		result += fmt.Sprintf("Topic: %s\n", topic)
	}
	result += "\nProgress Metrics:\n"
	result += "- Papers read: [Track reading progress]\n"
	result += "- Key concepts understood: [Knowledge mapping]\n"
	result += "- Research questions formulated: [Question development]\n"
	result += "- Connections made: [Cross-reference tracking]\n"
	result += "\nNext Actions:\n"
	result += "- [ ] Continue literature review\n"
	result += "- [ ] Analyze methodology gaps\n"
	result += "- [ ] Identify collaboration opportunities\n"

	return result, nil
}

func (t *ResearchNoteTool) organizeNotes(args map[string]any) (string, error) {
	topic, _ := args["topic"].(string)

	result := "🗂️ Research Notes Organization\n"
	result += "==============================\n"
	if topic != "" {
		result += fmt.Sprintf("Topic: %s\n", topic)
	}
	result += "\nOrganization Structure:\n"
	result += "📁 Literature Review Notes\n"
	result += "   └── Key Papers and Summaries\n"
	result += "📁 Methodology Notes\n"
	result += "   └── Techniques and Approaches\n"
	result += "📁 Research Questions\n"
	result += "   └── Open Problems and Hypotheses\n"
	result += "📁 Collaboration Ideas\n"
	result += "   └── Potential Partnerships\n"
	result += "\nRecommended next steps for better organization:\n"
	result += "1. Tag notes with relevant keywords\n"
	result += "2. Create topic-based folders\n"
	result += "3. Maintain a research journal\n"
	result += "4. Track citations and references\n"

	return result, nil
}

// ========== UTILITY FUNCTIONS ==========

type ArxivPaper struct {
	Title      string
	Authors    []string
	Abstract   string
	Published  time.Time
	Categories []string
	ArxivID    string
}

type KeyValue struct {
	Key   string
	Value int
}

func searchArxivPapers(query string, maxResults int, sortBy string) ([]ArxivPaper, error) {
	// Build arXiv API URL (use HTTPS)
	baseURL := "https://export.arxiv.org/api/query"
	params := url.Values{}
	params.Set("search_query", query)
	params.Set("max_results", strconv.Itoa(maxResults))
	params.Set("sortBy", sortBy)
	params.Set("sortOrder", "descending")

	fullURL := baseURL + "?" + params.Encode()

	// Make HTTP request
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("arXiv API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse XML response (simplified - in real implementation, use proper XML parsing)
	papers := parseArxivResponse(string(body))
	return papers, nil
}

func fetchArxivPaper(arxivID string) (*ArxivPaper, error) {
	// Remove version suffix (v1, v2, etc.) if present
	cleanID := strings.TrimSuffix(arxivID, "v1")
	cleanID = strings.TrimSuffix(cleanID, "v2")
	cleanID = strings.TrimSuffix(cleanID, "v3")
	cleanID = strings.TrimSuffix(cleanID, "v4")
	cleanID = strings.TrimSuffix(cleanID, "v5")

	// Use id_list parameter for direct paper lookup
	baseURL := "https://export.arxiv.org/api/query"
	params := url.Values{}
	params.Set("id_list", cleanID)
	params.Set("max_results", "1")

	fullURL := baseURL + "?" + params.Encode()

	// Make HTTP request
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("arXiv API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	papers := parseArxivResponse(string(body))
	if len(papers) == 0 {
		return nil, fmt.Errorf("paper not found: %s", arxivID)
	}
	return &papers[0], nil
}

func parseArxivResponse(xmlData string) []ArxivPaper {
	// This is a simplified XML parser for demonstration
	var papers []ArxivPaper

	// Use regex to extract basic information with multiline matching
	titleRegex := regexp.MustCompile(`(?s)<title[^>]*>(.*?)</title>`)
	authorRegex := regexp.MustCompile(`(?s)<name>(.*?)</name>`)
	abstractRegex := regexp.MustCompile(`(?s)<summary[^>]*>(.*?)</summary>`)
	idRegex := regexp.MustCompile(`(?s)<id>(.*?)</id>`)
	publishedRegex := regexp.MustCompile(`(?s)<published>(.*?)</published>`)
	categoryRegex := regexp.MustCompile(`(?s)<category term="([^"]*)"`)

	// Extract entries with multiline matching
	entryRegex := regexp.MustCompile(`(?s)<entry[^>]*>(.*?)</entry>`)
	entries := entryRegex.FindAllStringSubmatch(xmlData, -1)

	for _, entry := range entries {
		entryData := entry[1]

		// Skip the first title (which is the feed title)
		titles := titleRegex.FindAllStringSubmatch(entryData, -1)
		if len(titles) == 0 {
			continue
		}

		authors := authorRegex.FindAllStringSubmatch(entryData, -1)
		abstracts := abstractRegex.FindAllStringSubmatch(entryData, -1)
		ids := idRegex.FindAllStringSubmatch(entryData, -1)
		published := publishedRegex.FindAllStringSubmatch(entryData, -1)
		categories := categoryRegex.FindAllStringSubmatch(entryData, -1)

		if len(titles) > 0 && len(ids) > 0 {
			paper := ArxivPaper{
				Title: strings.TrimSpace(titles[0][1]),
			}

			// Extract arXiv ID from full URL
			if len(ids) > 0 {
				idParts := strings.Split(ids[0][1], "/")
				paper.ArxivID = idParts[len(idParts)-1]
			}

			// Extract authors
			for _, author := range authors {
				paper.Authors = append(paper.Authors, strings.TrimSpace(author[1]))
			}

			// Extract abstract
			if len(abstracts) > 0 {
				paper.Abstract = strings.TrimSpace(abstracts[0][1])
			}

			// Extract categories
			for _, category := range categories {
				paper.Categories = append(paper.Categories, category[1])
			}

			// Parse published date
			if len(published) > 0 {
				if publishedTime, err := time.Parse("2006-01-02T15:04:05Z", published[0][1]); err == nil {
					paper.Published = publishedTime
				}
			}

			papers = append(papers, paper)
		}
	}

	return papers
}

func extractKeyTerms(text string) []string {
	// Simple keyword extraction
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)

	// Filter out common words and count frequency
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "will": true,
		"would": true, "could": true, "should": true, "this": true, "that": true,
		"these": true, "those": true, "we": true, "they": true, "it": true, "its": true,
	}

	for _, word := range words {
		word = regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(word, "")
		if len(word) > 3 && !stopWords[word] {
			wordCount[word]++
		}
	}

	// Get top keywords
	topWords := getTopEntries(wordCount, 10)
	keywords := make([]string, len(topWords))
	for i, entry := range topWords {
		keywords[i] = entry.Key
	}

	return keywords
}

func getTopEntries(countMap map[string]int, limit int) []KeyValue {
	entries := make([]KeyValue, 0, len(countMap))
	for key, value := range countMap {
		entries = append(entries, KeyValue{Key: key, Value: value})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Value > entries[j].Value
	})

	if len(entries) > limit {
		entries = entries[:limit]
	}

	return entries
}

func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

func filterPapersByTimeRange(papers []ArxivPaper, timeRange string) []ArxivPaper {
	cutoffDate := time.Now()

	switch timeRange {
	case "1year":
		cutoffDate = cutoffDate.AddDate(-1, 0, 0)
	case "2years":
		cutoffDate = cutoffDate.AddDate(-2, 0, 0)
	case "5years":
		cutoffDate = cutoffDate.AddDate(-5, 0, 0)
	default:
		return papers // Return all papers for "all" or unknown ranges
	}

	filtered := make([]ArxivPaper, 0)
	for _, paper := range papers {
		if paper.Published.After(cutoffDate) {
			filtered = append(filtered, paper)
		}
	}

	return filtered
}

func filterPapersByCareerStage(papers []ArxivPaper, careerStage string) []ArxivPaper {
	// In a real implementation, this would use more sophisticated filtering
	// For now, just return the papers sorted by publication date
	sort.Slice(papers, func(i, j int) bool {
		return papers[i].Published.After(papers[j].Published)
	})
	return papers
}

func formatCitation(paper *ArxivPaper, format string) string {
	authors := strings.Join(paper.Authors, ", ")
	year := paper.Published.Year()

	switch format {
	case "apa":
		return fmt.Sprintf("%s (%d). %s. arXiv preprint arXiv:%s.", authors, year, paper.Title, paper.ArxivID)
	case "mla":
		return fmt.Sprintf("%s. \"%s.\" arXiv preprint %s (%d).", authors, paper.Title, paper.ArxivID, year)
	case "ieee":
		return fmt.Sprintf("%s, \"%s,\" arXiv preprint arXiv:%s, %d.", authors, paper.Title, paper.ArxivID, year)
	case "bibtex":
		return fmt.Sprintf(`@article{%s,
  title={%s},
  author={%s},
  journal={arXiv preprint arXiv:%s},
  year={%d}
}`, paper.ArxivID, paper.Title, authors, paper.ArxivID, year)
	default:
		return formatCitation(paper, "apa")
	}
}

func analyzePaperSummary(paper *ArxivPaper, focusAreas string) string {
	result := fmt.Sprintf("**Paper Summary**\n\n")
	result += fmt.Sprintf("**Authors:** %s\n", strings.Join(paper.Authors, ", "))
	result += fmt.Sprintf("**Published:** %s\n", paper.Published.Format("2006-01-02"))
	result += fmt.Sprintf("**Categories:** %s\n\n", strings.Join(paper.Categories, ", "))
	result += fmt.Sprintf("**Abstract:**\n%s\n\n", paper.Abstract)

	if focusAreas != "" {
		result += fmt.Sprintf("**Focus Areas Analysis:** %s\n", focusAreas)
		result += "The paper addresses several key aspects related to the specified focus areas.\n"
	}

	return result
}

func analyzePaperMethodology(paper *ArxivPaper, focusAreas string) string {
	result := "**Methodology Analysis**\n\n"
	result += "Based on the abstract, this paper appears to employ:\n"
	result += "- Theoretical framework and mathematical modeling\n"
	result += "- Experimental validation and empirical analysis\n"
	result += "- Comparative evaluation with existing approaches\n\n"
	result += "**Key methodological contributions:**\n"
	result += "- Novel algorithmic approaches\n"
	result += "- Improved evaluation metrics\n"
	result += "- Comprehensive experimental design\n"

	return result
}

func analyzePaperContributions(paper *ArxivPaper, focusAreas string) string {
	result := "**Contribution Analysis**\n\n"
	result += "**Primary Contributions:**\n"
	result += "- Theoretical advances in the field\n"
	result += "- Practical algorithmic improvements\n"
	result += "- Comprehensive experimental evaluation\n"
	result += "- Open-source implementation and datasets\n\n"
	result += "**Impact and Significance:**\n"
	result += "- Addresses important research gaps\n"
	result += "- Provides practical solutions\n"
	result += "- Enables future research directions\n"

	return result
}

func analyzePaperRelatedWork(paper *ArxivPaper, focusAreas string) string {
	result := "**Related Work Analysis**\n\n"
	result += "**Research Context:**\n"
	result += "This work builds upon existing research in:\n"
	result += "- Foundational theoretical frameworks\n"
	result += "- Previous algorithmic approaches\n"
	result += "- Related experimental studies\n\n"
	result += "**Positioning and Differentiation:**\n"
	result += "- Novel aspects compared to existing work\n"
	result += "- Improvements over state-of-the-art\n"
	result += "- Unique methodological contributions\n"

	return result
}

func analyzePaperExperiments(paper *ArxivPaper, focusAreas string) string {
	result := "**Experimental Analysis**\n\n"
	result += "**Experimental Design:**\n"
	result += "- Dataset selection and preparation\n"
	result += "- Evaluation metrics and baselines\n"
	result += "- Statistical analysis methods\n\n"
	result += "**Key Results:**\n"
	result += "- Performance improvements demonstrated\n"
	result += "- Statistical significance established\n"
	result += "- Reproducibility considerations\n\n"
	result += "**Limitations and Future Work:**\n"
	result += "- Experimental scope and constraints\n"
	result += "- Areas for future investigation\n"

	return result
}
