package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AIAnalysisManager manages AI-powered code analysis
type AIAnalysisManager struct {
	analyzer         *AICodeAnalyzer
	analysisResults  map[string]*ProjectAnalysisResult
	isAnalyzing      bool
	lastAnalysisTime time.Time
	mutex            sync.RWMutex
}

// NewAIAnalysisManager creates a new AI analysis manager
func NewAIAnalysisManager(workspace string) *AIAnalysisManager {
	return &AIAnalysisManager{
		analyzer:        NewAICodeAnalyzer(workspace),
		analysisResults: make(map[string]*ProjectAnalysisResult),
		isAnalyzing:     false,
	}
}

// StartBackgroundAnalysis begins continuous AI analysis
func (aam *AIAnalysisManager) StartBackgroundAnalysis() {
	log.Println("Starting background AI analysis...")
	
	// Run initial analysis
	go aam.runFullAnalysis()
	
	// Schedule periodic analysis (every 30 minutes)
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			aam.runPeriodicAnalysis()
		}
	}()
}

// runFullAnalysis performs a complete project analysis
func (aam *AIAnalysisManager) runFullAnalysis() {
	aam.mutex.Lock()
	defer aam.mutex.Unlock()
	
	if aam.isAnalyzing {
		log.Println("Analysis already in progress, skipping...")
		return
	}
	
	aam.isAnalyzing = true
	defer func() { aam.isAnalyzing = false }()
	
	log.Println("Running full AI project analysis...")
	start := time.Now()
	
	result, err := aam.analyzer.AnalyzeProject()
	if err != nil {
		log.Printf("Error during AI analysis: %v", err)
		return
	}
	
	// Store results
	analysisID := fmt.Sprintf("analysis_%d", time.Now().Unix())
	aam.analysisResults[analysisID] = result
	aam.lastAnalysisTime = time.Now()
	
	duration := time.Since(start)
	log.Printf("AI analysis completed in %v. Found %d issues across %d files", 
		duration, 
		result.Summary.TotalSecurityIssues + result.Summary.TotalPerformanceIssues + result.Summary.TotalCodeSmells,
		result.Summary.AnalyzedFiles)
	
	// Log summary of critical issues
	if result.Summary.TotalSecurityIssues > 0 {
		log.Printf("⚠️ CRITICAL: %d security issues found", result.Summary.TotalSecurityIssues)
	}
	
	if result.Summary.AverageComplexity > 70 {
		log.Printf("⚠️ High average complexity: %.1f", result.Summary.AverageComplexity)
	}
}

// runPeriodicAnalysis runs a lighter analysis for changed files
func (aam *AIAnalysisManager) runPeriodicAnalysis() {
	aam.mutex.RLock()
	defer aam.mutex.RUnlock()
	
	if aam.isAnalyzing {
		return
	}
	
	log.Println("Running periodic AI analysis for recent changes...")
	// This could be optimized to only analyze changed files
	go aam.runFullAnalysis()
}

// GetLatestAnalysis returns the most recent analysis results
func (aam *AIAnalysisManager) GetLatestAnalysis() *ProjectAnalysisResult {
	aam.mutex.RLock()
	defer aam.mutex.RUnlock()
	
	var latest *ProjectAnalysisResult
	var latestTime time.Time
	
	for _, result := range aam.analysisResults {
		if result.StartTime.After(latestTime) {
			latest = result
			latestTime = result.StartTime
		}
	}
	
	return latest
}

// GetAnalysisStatus returns current analysis status
func (aam *AIAnalysisManager) GetAnalysisStatus() map[string]interface{} {
	aam.mutex.RLock()
	defer aam.mutex.RUnlock()
	
	status := map[string]interface{}{
		"is_analyzing":       aam.isAnalyzing,
		"last_analysis_time": aam.lastAnalysisTime,
		"total_analyses":     len(aam.analysisResults),
	}
	
	if latest := aam.GetLatestAnalysis(); latest != nil {
		status["latest_summary"] = latest.Summary
	}
	
	return status
}

// AnalyzeFile performs on-demand analysis of a specific file
func (aam *AIAnalysisManager) AnalyzeFile(filePath string) (*CodeAnalysisResult, error) {
	return aam.analyzer.analyzeFile(filePath)
}

// GetAISuggestions returns AI suggestions for the project
func (aam *AIAnalysisManager) GetAISuggestions() []AISuggestion {
	latest := aam.GetLatestAnalysis()
	if latest == nil {
		return []AISuggestion{}
	}
	
	suggestions := []AISuggestion{}
	for _, file := range latest.Files {
		suggestions = append(suggestions, file.Suggestions...)
	}
	
	return suggestions
}

// Integration with Enhanced Intelligence Server

// setupAIAnalysisRoutes adds AI analysis endpoints to the server
func (eis *EnhancedIntelligenceServer) setupAIAnalysisRoutes() {
	// Initialize AI analysis manager
	eis.aiAnalysisManager = NewAIAnalysisManager(eis.epi.workspace)
	eis.aiAnalysisManager.StartBackgroundAnalysis()
	
	// API endpoints
	eis.app.Get("/api/ai/analysis/status", eis.aiAnalysisStatusHandler)
	eis.app.Get("/api/ai/analysis/latest", eis.latestAnalysisHandler)
	eis.app.Post("/api/ai/analysis/run", eis.triggerAnalysisHandler)
	eis.app.Get("/api/ai/suggestions", eis.aiSuggestionsHandler)
	eis.app.Get("/api/ai/analysis/file/:path", eis.analyzeFileHandler)
	eis.app.Get("/api/ai/code-quality", eis.codeQualityHandler)
	eis.app.Get("/api/ai/security-report", eis.securityReportHandler)
	eis.app.Get("/api/ai/performance-report", eis.performanceReportHandler)
	eis.app.Get("/api/ai/recommendations", eis.aiRecommendationsHandler)
}

// HTTP Handlers

func (eis *EnhancedIntelligenceServer) aiAnalysisStatusHandler(c *fiber.Ctx) error {
	status := eis.aiAnalysisManager.GetAnalysisStatus()
	return c.JSON(fiber.Map{
		"status":    "success",
		"data":      status,
		"timestamp": time.Now(),
	})
}

func (eis *EnhancedIntelligenceServer) latestAnalysisHandler(c *fiber.Ctx) error {
	analysis := eis.aiAnalysisManager.GetLatestAnalysis()
	if analysis == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "No analysis results available",
		})
	}
	
	return c.JSON(fiber.Map{
		"status":   "success",
		"analysis": analysis,
	})
}

func (eis *EnhancedIntelligenceServer) triggerAnalysisHandler(c *fiber.Ctx) error {
	// Trigger new analysis
	go eis.aiAnalysisManager.runFullAnalysis()
	
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "AI analysis triggered",
		"eta":     "2-5 minutes",
	})
}

func (eis *EnhancedIntelligenceServer) aiSuggestionsHandler(c *fiber.Ctx) error {
	suggestions := eis.aiAnalysisManager.GetAISuggestions()
	
	// Filter by category if requested
	category := c.Query("category")
	if category != "" {
		filtered := []AISuggestion{}
		for _, suggestion := range suggestions {
			if suggestion.Category == category {
				filtered = append(filtered, suggestion)
			}
		}
		suggestions = filtered
	}
	
	// Limit results
	limit := 50
	if c.Query("limit") != "" {
		if l, err := parseLimit(c.Query("limit")); err == nil {
			limit = l
		}
	}
	
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}
	
	return c.JSON(fiber.Map{
		"status":      "success",
		"suggestions": suggestions,
		"total":       len(suggestions),
		"categories":  []string{"security", "performance", "maintainability", "refactoring"},
	})
}

func (eis *EnhancedIntelligenceServer) analyzeFileHandler(c *fiber.Ctx) error {
	filePath := c.Params("path")
	if filePath == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "File path is required",
		})
	}
	
	result, err := eis.aiAnalysisManager.AnalyzeFile(filePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Analysis failed: %v", err),
		})
	}
	
	return c.JSON(fiber.Map{
		"status": "success",
		"result": result,
	})
}

func (eis *EnhancedIntelligenceServer) codeQualityHandler(c *fiber.Ctx) error {
	analysis := eis.aiAnalysisManager.GetLatestAnalysis()
	if analysis == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "No analysis data available",
		})
	}
	
	// Calculate quality metrics
	qualityReport := map[string]interface{}{
		"overall_score":        analysis.Summary.OverallScore,
		"average_complexity":   analysis.Summary.AverageComplexity,
		"total_files":          analysis.Summary.AnalyzedFiles,
		"technical_debt_hours": analysis.Summary.TechnicalDebtHours,
		"maintainability_distribution": eis.calculateMaintainabilityDistribution(analysis),
		"complexity_distribution":      eis.calculateComplexityDistribution(analysis),
		"language_quality":             eis.calculateLanguageQuality(analysis),
		"quality_trends":               eis.getQualityTrends(),
		"recommendations":              analysis.Recommendations,
	}
	
	return c.JSON(fiber.Map{
		"status": "success",
		"report": qualityReport,
	})
}

func (eis *EnhancedIntelligenceServer) securityReportHandler(c *fiber.Ctx) error {
	analysis := eis.aiAnalysisManager.GetLatestAnalysis()
	if analysis == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "No analysis data available",
		})
	}
	
	// Aggregate security issues
	securityIssues := []SecurityIssue{}
	severityCount := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}
	
	for _, file := range analysis.Files {
		for _, issue := range file.SecurityIssues {
			securityIssues = append(securityIssues, issue)
			severityCount[issue.Severity]++
		}
	}
	
	securityReport := map[string]interface{}{
		"total_issues":     len(securityIssues),
		"severity_breakdown": severityCount,
		"issues":           securityIssues,
		"risk_score":       eis.calculateSecurityRiskScore(severityCount),
		"recommendations":  eis.generateSecurityRecommendations(severityCount),
	}
	
	return c.JSON(fiber.Map{
		"status": "success",
		"report": securityReport,
	})
}

func (eis *EnhancedIntelligenceServer) performanceReportHandler(c *fiber.Ctx) error {
	analysis := eis.aiAnalysisManager.GetLatestAnalysis()
	if analysis == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "No analysis data available",
		})
	}
	
	// Aggregate performance issues
	performanceIssues := []PerformanceIssue{}
	typeCount := map[string]int{}
	
	for _, file := range analysis.Files {
		for _, issue := range file.PerformanceIssues {
			performanceIssues = append(performanceIssues, issue)
			typeCount[issue.Type]++
		}
	}
	
	performanceReport := map[string]interface{}{
		"total_issues":       len(performanceIssues),
		"issue_types":        typeCount,
		"issues":             performanceIssues,
		"performance_score":  eis.calculatePerformanceScore(analysis),
		"optimization_priority": eis.prioritizeOptimizations(performanceIssues),
	}
	
	return c.JSON(fiber.Map{
		"status": "success",
		"report": performanceReport,
	})
}

func (eis *EnhancedIntelligenceServer) aiRecommendationsHandler(c *fiber.Ctx) error {
	analysis := eis.aiAnalysisManager.GetLatestAnalysis()
	if analysis == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "No analysis data available",
		})
	}
	
	recommendations := map[string]interface{}{
		"immediate_actions":   eis.getImmediateActions(analysis),
		"short_term_goals":    eis.getShortTermGoals(analysis),
		"long_term_strategy":  eis.getLongTermStrategy(analysis),
		"learning_resources":  eis.getLearningResources(analysis),
		"tools_suggestions":   eis.getToolSuggestions(analysis),
	}
	
	return c.JSON(fiber.Map{
		"status":          "success",
		"recommendations": recommendations,
		"generated_at":    time.Now(),
	})
}

// Helper functions for report generation

func (eis *EnhancedIntelligenceServer) calculateMaintainabilityDistribution(analysis *ProjectAnalysisResult) map[string]int {
	distribution := map[string]int{
		"excellent": 0,
		"good":      0,
		"fair":      0,
		"poor":      0,
		"critical":  0,
	}
	
	for _, file := range analysis.Files {
		distribution[file.Maintainability]++
	}
	
	return distribution
}

func (eis *EnhancedIntelligenceServer) calculateComplexityDistribution(analysis *ProjectAnalysisResult) map[string]int {
	distribution := map[string]int{
		"low":    0,  // 0-20
		"medium": 0,  // 21-50
		"high":   0,  // 51-80
		"very_high": 0, // 81+
	}
	
	for _, file := range analysis.Files {
		if file.ComplexityScore <= 20 {
			distribution["low"]++
		} else if file.ComplexityScore <= 50 {
			distribution["medium"]++
		} else if file.ComplexityScore <= 80 {
			distribution["high"]++
		} else {
			distribution["very_high"]++
		}
	}
	
	return distribution
}

func (eis *EnhancedIntelligenceServer) calculateLanguageQuality(analysis *ProjectAnalysisResult) map[string]interface{} {
	languageStats := make(map[string]map[string]interface{})
	
	for _, file := range analysis.Files {
		if _, exists := languageStats[file.Language]; !exists {
			languageStats[file.Language] = map[string]interface{}{
				"files":            0,
				"total_complexity": 0,
				"total_issues":     0,
			}
		}
		
		stats := languageStats[file.Language]
		stats["files"] = stats["files"].(int) + 1
		stats["total_complexity"] = stats["total_complexity"].(int) + file.ComplexityScore
		stats["total_issues"] = stats["total_issues"].(int) + len(file.SecurityIssues) + len(file.PerformanceIssues) + len(file.CodeSmells)
	}
	
	// Calculate averages
	for lang, stats := range languageStats {
		fileCount := stats["files"].(int)
		if fileCount > 0 {
			stats["avg_complexity"] = float64(stats["total_complexity"].(int)) / float64(fileCount)
			stats["avg_issues"] = float64(stats["total_issues"].(int)) / float64(fileCount)
		}
		languageStats[lang] = stats
	}
	
	return map[string]interface{}{
		"by_language": languageStats,
	}
}

func (eis *EnhancedIntelligenceServer) getQualityTrends() map[string]interface{} {
	// This would track quality over time - simplified for demo
	return map[string]interface{}{
		"trend": "improving",
		"note":  "Quality trends require historical data",
	}
}

func (eis *EnhancedIntelligenceServer) calculateSecurityRiskScore(severityCount map[string]int) int {
	score := severityCount["critical"]*10 + severityCount["high"]*7 + severityCount["medium"]*4 + severityCount["low"]*1
	if score > 100 {
		score = 100
	}
	return score
}

func (eis *EnhancedIntelligenceServer) generateSecurityRecommendations(severityCount map[string]int) []string {
	recommendations := []string{}
	
	if severityCount["critical"] > 0 {
		recommendations = append(recommendations, "🚨 Address critical security issues immediately")
	}
	if severityCount["high"] > 0 {
		recommendations = append(recommendations, "🔒 Fix high-priority security vulnerabilities")
	}
	if severityCount["medium"]+severityCount["low"] > 10 {
		recommendations = append(recommendations, "🛡️ Consider implementing automated security scanning")
	}
	
	return recommendations
}

func (eis *EnhancedIntelligenceServer) calculatePerformanceScore(analysis *ProjectAnalysisResult) int {
	totalIssues := analysis.Summary.TotalPerformanceIssues
	totalFiles := analysis.Summary.AnalyzedFiles
	
	if totalFiles == 0 {
		return 100
	}
	
	issueRatio := float64(totalIssues) / float64(totalFiles)
	score := 100 - int(issueRatio*50) // Simplified scoring
	
	if score < 0 {
		score = 0
	}
	
	return score
}

func (eis *EnhancedIntelligenceServer) prioritizeOptimizations(issues []PerformanceIssue) []map[string]interface{} {
	priority := []map[string]interface{}{}
	
	// Group by type and prioritize
	typeCount := make(map[string]int)
	for _, issue := range issues {
		typeCount[issue.Type]++
	}
	
	for issueType, count := range typeCount {
		priority = append(priority, map[string]interface{}{
			"type":  issueType,
			"count": count,
			"priority": eis.getOptimizationPriority(issueType),
		})
	}
	
	return priority
}

func (eis *EnhancedIntelligenceServer) getOptimizationPriority(issueType string) string {
	priorities := map[string]string{
		"nested_loops":         "high",
		"string_concatenation": "medium",
		"inefficient_query":    "high",
		"memory_leak":          "critical",
	}
	
	if priority, exists := priorities[issueType]; exists {
		return priority
	}
	return "medium"
}

func (eis *EnhancedIntelligenceServer) getImmediateActions(analysis *ProjectAnalysisResult) []string {
	actions := []string{}
	
	if analysis.Summary.TotalSecurityIssues > 0 {
		actions = append(actions, "Fix security vulnerabilities")
	}
	if analysis.Summary.AverageComplexity > 80 {
		actions = append(actions, "Refactor high-complexity functions")
	}
	if analysis.Summary.TotalPerformanceIssues > 5 {
		actions = append(actions, "Optimize performance bottlenecks")
	}
	
	return actions
}

func (eis *EnhancedIntelligenceServer) getShortTermGoals(analysis *ProjectAnalysisResult) []string {
	goals := []string{
		"Establish automated code quality gates",
		"Implement comprehensive test coverage",
		"Set up continuous security scanning",
		"Create coding standards documentation",
	}
	return goals
}

func (eis *EnhancedIntelligenceServer) getLongTermStrategy(analysis *ProjectAnalysisResult) []string {
	strategy := []string{
		"Implement microservices architecture",
		"Adopt advanced monitoring and observability",
		"Establish DevSecOps practices",
		"Create automated deployment pipelines",
	}
	return strategy
}

func (eis *EnhancedIntelligenceServer) getLearningResources(analysis *ProjectAnalysisResult) []map[string]string {
	resources := []map[string]string{
		{"title": "Clean Code Practices", "type": "documentation", "priority": "high"},
		{"title": "Security Best Practices", "type": "training", "priority": "critical"},
		{"title": "Performance Optimization", "type": "workshop", "priority": "medium"},
	}
	return resources
}

func (eis *EnhancedIntelligenceServer) getToolSuggestions(analysis *ProjectAnalysisResult) []map[string]string {
	tools := []map[string]string{
		{"name": "SonarQube", "purpose": "Code quality analysis", "priority": "high"},
		{"name": "OWASP ZAP", "purpose": "Security testing", "priority": "high"},
		{"name": "Go vet", "purpose": "Static analysis", "priority": "medium"},
	}
	return tools
}

// Utility functions

func parseLimit(limitStr string) (int, error) {
	// Simple limit parsing - in production, add proper validation
	limit := 50
	if l, err := json.Number(limitStr).Int64(); err == nil && l > 0 && l <= 1000 {
		limit = int(l)
	}
	return limit, nil
}

// Add AI analysis manager to EnhancedIntelligenceServer
type EnhancedIntelligenceServerWithAI struct {
	*EnhancedIntelligenceServer
	aiAnalysisManager *AIAnalysisManager
}