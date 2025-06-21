package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// EnhancedProjectIntelligence extends ProjectIntelligence with multi-language support
type EnhancedProjectIntelligence struct {
	*ProjectIntelligence
	languageDetector  *LanguageDetector
	pluginManager     *LanguagePluginManager
	detectedLanguages []DetectedLanguage
	projectTopology   *ProjectTopology
	mutex             sync.RWMutex
}

// NewEnhancedProjectIntelligence creates an enhanced project intelligence instance
func NewEnhancedProjectIntelligence(workspace string) *EnhancedProjectIntelligence {
	basePI := NewProjectIntelligence(workspace)

	return &EnhancedProjectIntelligence{
		ProjectIntelligence: basePI,
		languageDetector:    NewLanguageDetector(workspace),
		pluginManager:       NewLanguagePluginManager(),
		detectedLanguages:   []DetectedLanguage{},
		projectTopology:     &ProjectTopology{},
	}
}

// StartWatching begins enhanced monitoring with language detection
func (epi *EnhancedProjectIntelligence) StartWatching() {
	log.Printf("Starting enhanced project intelligence monitoring for: %s", epi.workspace)

	// Start base monitoring
	epi.ProjectIntelligence.StartWatching()

	// Start language-specific monitoring
	go epi.startLanguageMonitoring()

	// Generate initial enhanced snapshot
	go func() {
		time.Sleep(3 * time.Second) // Give watchers time to initialize
		epi.updateEnhancedSnapshot()

		// Update enhanced snapshot every 45 seconds
		ticker := time.NewTicker(45 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			epi.updateEnhancedSnapshot()
		}
	}()
}

// startLanguageMonitoring begins language-specific monitoring
func (epi *EnhancedProjectIntelligence) startLanguageMonitoring() {
	log.Println("Starting multi-language monitoring...")

	// Initial language detection
	epi.detectLanguages()

	// Periodic language detection and error analysis
	ticker := time.NewTicker(60 * time.Second) // Check every minute
	defer ticker.Stop()

	for range ticker.C {
		epi.detectLanguages()
		epi.analyzeLanguageErrors()
	}
}

// detectLanguages performs language detection
func (epi *EnhancedProjectIntelligence) detectLanguages() {
	epi.mutex.Lock()
	defer epi.mutex.Unlock()

	detected, err := epi.languageDetector.DetectLanguages()
	if err != nil {
		log.Printf("Error detecting languages: %v", err)
		return
	}

	epi.detectedLanguages = detected
	log.Printf("Detected %d languages in project", len(detected))

	// Log detected languages for debugging
	for _, lang := range detected {
		log.Printf("  %s: %d files, %d lines", lang.Language.Name, lang.FileCount, lang.LineCount)
	}
}

// analyzeLanguageErrors runs error analysis for all detected languages
func (epi *EnhancedProjectIntelligence) analyzeLanguageErrors() {
	allErrors, err := epi.pluginManager.AnalyzeAllErrors(epi.workspace)
	if err != nil {
		log.Printf("Error analyzing language errors: %v", err)
		return
	}

	if len(allErrors) > 0 {
		log.Printf("Found %d language-specific errors", len(allErrors))

		// Merge with existing errors
		epi.mutex.Lock()
		epi.errorWatcher.mutex.Lock()

		// Add language-specific errors to the error watcher
		for _, err := range allErrors {
			epi.errorWatcher.errors = append(epi.errorWatcher.errors, err)
		}

		epi.errorWatcher.mutex.Unlock()
		epi.mutex.Unlock()
	}
}

// updateEnhancedSnapshot creates an enhanced project snapshot with language information
func (epi *EnhancedProjectIntelligence) updateEnhancedSnapshot() {
	epi.mutex.Lock()
	defer epi.mutex.Unlock()

	log.Println("Updating enhanced project snapshot...")

	// Update base snapshot first
	epi.ProjectIntelligence.updateSnapshot()

	// Update language topology
	topology, err := epi.languageDetector.GetProjectTopology()
	if err != nil {
		log.Printf("Error getting project topology: %v", err)
		return
	}

	epi.projectTopology = topology
	log.Printf("Enhanced snapshot updated - %d languages detected", len(topology.Languages))
}

// GetEnhancedSnapshot returns the enhanced project snapshot
func (epi *EnhancedProjectIntelligence) GetEnhancedSnapshot() *EnhancedProjectSnapshot {
	epi.mutex.RLock()
	defer epi.mutex.RUnlock()

	baseSnapshot := epi.ProjectIntelligence.GetSnapshot()
	if baseSnapshot == nil {
		return nil
	}

	return &EnhancedProjectSnapshot{
		ProjectSnapshot: *baseSnapshot,
		Languages:       epi.detectedLanguages,
		Topology:        epi.projectTopology,
		UpdatedAt:       time.Now(),
	}
}

// GetDetectedLanguages returns the list of detected languages
func (epi *EnhancedProjectIntelligence) GetDetectedLanguages() []DetectedLanguage {
	epi.mutex.RLock()
	defer epi.mutex.RUnlock()

	return append([]DetectedLanguage{}, epi.detectedLanguages...)
}

// GetProjectTopology returns the current project topology
func (epi *EnhancedProjectIntelligence) GetProjectTopology() *ProjectTopology {
	epi.mutex.RLock()
	defer epi.mutex.RUnlock()

	if epi.projectTopology == nil {
		return &ProjectTopology{
			Languages: []DetectedLanguage{},
			Services:  []ServiceInfo{},
			Databases: []DatabaseInfo{},
			APIs:      []APIEndpoint{},
			Relations: []ServiceRelation{},
			UpdatedAt: time.Now(),
		}
	}

	return epi.projectTopology
}

// AnalyzeLanguageErrors runs error analysis for a specific language
func (epi *EnhancedProjectIntelligence) AnalyzeLanguageErrors(languageName string) ([]ErrorInfo, error) {
	plugin, exists := epi.pluginManager.GetPlugin(languageName)
	if !exists {
		return nil, fmt.Errorf("language plugin not found: %s", languageName)
	}

	return plugin.AnalyzeErrors(epi.workspace)
}

// RunLanguageTests runs tests for a specific language
func (epi *EnhancedProjectIntelligence) RunLanguageTests(languageName string) (*TestResults, error) {
	plugin, exists := epi.pluginManager.GetPlugin(languageName)
	if !exists {
		return nil, fmt.Errorf("language plugin not found: %s", languageName)
	}

	return plugin.RunTests(epi.workspace)
}

// RunLanguageLinter runs linter for a specific language
func (epi *EnhancedProjectIntelligence) RunLanguageLinter(languageName string) ([]ErrorInfo, error) {
	plugin, exists := epi.pluginManager.GetPlugin(languageName)
	if !exists {
		return nil, fmt.Errorf("language plugin not found: %s", languageName)
	}

	return plugin.RunLinter(epi.workspace)
}

// GetLanguageDependencies gets dependencies for a specific language
func (epi *EnhancedProjectIntelligence) GetLanguageDependencies(languageName string) ([]DependencyInfo, error) {
	plugin, exists := epi.pluginManager.GetPlugin(languageName)
	if !exists {
		return nil, fmt.Errorf("language plugin not found: %s", languageName)
	}

	return plugin.GetDependencies(epi.workspace)
}

// FindLanguageServices finds services for a specific language
func (epi *EnhancedProjectIntelligence) FindLanguageServices(languageName string) ([]ServiceInfo, error) {
	plugin, exists := epi.pluginManager.GetPlugin(languageName)
	if !exists {
		return nil, fmt.Errorf("language plugin not found: %s", languageName)
	}

	return plugin.FindServices(epi.workspace)
}

// EnhancedProjectSnapshot extends ProjectSnapshot with language information
type EnhancedProjectSnapshot struct {
	ProjectSnapshot
	Languages []DetectedLanguage `json:"languages"`
	Topology  *ProjectTopology   `json:"topology"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// EnhancedIntelligenceServer extends IntelligenceServer with language capabilities
type EnhancedIntelligenceServer struct {
	*IntelligenceServer
	epi                  *EnhancedProjectIntelligence
	snapshotManager      *SnapshotManager
	investigationTracker *InvestigationTracker
	versionManager       *VersionManager
	snapshotExporter     *SnapshotExporter
	snapshotSharer       *SnapshotSharer
}

// NewEnhancedIntelligenceServer creates an enhanced intelligence server
func NewEnhancedIntelligenceServer(workspace string) *EnhancedIntelligenceServer {
	baseServer := NewIntelligenceServer(workspace)
	epi := NewEnhancedProjectIntelligence(workspace)
	snapshotManager := NewSnapshotManager(filepath.Join(workspace, ".argus", "snapshots"))
	investigationTracker := NewInvestigationTracker()
	versionManager := NewVersionManager(snapshotManager)
	snapshotExporter := NewSnapshotExporter(snapshotManager)
	snapshotSharer := NewSnapshotSharer()

	server := &EnhancedIntelligenceServer{
		IntelligenceServer:   baseServer,
		epi:                  epi,
		snapshotManager:      snapshotManager,
		investigationTracker: investigationTracker,
		versionManager:       versionManager,
		snapshotExporter:     snapshotExporter,
		snapshotSharer:       snapshotSharer,
	}

	// Replace the base PI with enhanced PI
	server.IntelligenceServer.pi = epi.ProjectIntelligence

	// Setup enhanced routes
	server.setupEnhancedRoutes()

	// Start enhanced monitoring
	epi.StartWatching()

	return server
}

// setupEnhancedRoutes adds language-specific API endpoints
func (eis *EnhancedIntelligenceServer) setupEnhancedRoutes() {
	// Language detection and analysis endpoints
	eis.app.Get("/api/languages", eis.languagesHandler)
	eis.app.Get("/api/languages/:language/errors", eis.languageErrorsHandler)
	eis.app.Get("/api/languages/:language/dependencies", eis.languageDependenciesHandler)
	eis.app.Get("/api/languages/:language/services", eis.languageServicesHandler)
	eis.app.Post("/api/languages/:language/lint", eis.languageLintHandler)
	eis.app.Post("/api/languages/:language/test", eis.languageTestHandler)

	// Enhanced project topology
	eis.app.Get("/api/topology", eis.topologyHandler)
	eis.app.Get("/api/topology/detailed", eis.detailedTopologyHandler)
	eis.app.Get("/api/topology/mindmap", eis.mindmapDataHandler)
	eis.app.Get("/api/enhanced-snapshot", eis.enhancedSnapshotHandler)

	// Mind map specific endpoints
	eis.app.Get("/api/mindmap/nodes", eis.mindmapNodesHandler)
	eis.app.Get("/api/mindmap/relationships", eis.mindmapRelationshipsHandler)
	eis.app.Get("/mindmap", eis.mindmapPageHandler)
	eis.app.Get("/snapshots", eis.snapshotsPageHandler)
	eis.app.Get("/", eis.enhancedDashboardHandler)

	// Universal analysis endpoints
	eis.app.Post("/api/analyze/all-languages", eis.analyzeAllLanguagesHandler)
	eis.app.Get("/api/project-overview", eis.projectOverviewHandler)

	// Snapshot system endpoints
	eis.app.Post("/api/snapshots", eis.createSnapshotHandler)
	eis.app.Get("/api/snapshots", eis.listSnapshotsHandler)
	eis.app.Get("/api/snapshots/:id", eis.getSnapshotHandler)
	eis.app.Put("/api/snapshots/:id", eis.updateSnapshotHandler)
	eis.app.Delete("/api/snapshots/:id", eis.deleteSnapshotHandler)
	eis.app.Post("/api/snapshots/:id/restore", eis.restoreSnapshotHandler)
	eis.app.Post("/api/snapshots/:id/fork", eis.forkSnapshotHandler)

	// Investigation tracking endpoints
	eis.app.Post("/api/investigation/questions", eis.addQuestionHandler)
	eis.app.Put("/api/investigation/questions/:id", eis.updateQuestionHandler)
	eis.app.Post("/api/investigation/hypotheses", eis.addHypothesisHandler)
	eis.app.Put("/api/investigation/hypotheses/:id", eis.updateHypothesisHandler)
	eis.app.Post("/api/investigation/findings", eis.addFindingHandler)
	eis.app.Post("/api/investigation/blockers", eis.addBlockerHandler)
	eis.app.Put("/api/investigation/blockers/:id", eis.resolveBlockerHandler)

	// Collaboration endpoints
	eis.app.Post("/api/collaboration/ai-recommendation", eis.aiRecommendationHandler)
	eis.app.Post("/api/collaboration/human-decision", eis.humanDecisionHandler)
	eis.app.Post("/api/collaboration/communication", eis.communicationHandler)
	eis.app.Get("/api/collaboration/session", eis.getSessionHandler)

	// Investigation state endpoint
	eis.app.Get("/api/investigation/state", eis.getInvestigationStateHandler)

	// Versioning endpoints
	eis.app.Get("/api/snapshots/:id/version", eis.getVersionInfoHandler)
	eis.app.Get("/api/snapshots/versions", eis.getAllVersionsHandler)

	// Export and sharing endpoints
	eis.app.Post("/api/snapshots/:id/export", eis.exportSnapshotHandler)
	eis.app.Post("/api/snapshots/:id/share", eis.createShareLinkHandler)
	eis.app.Get("/api/shared/:linkId", eis.accessSharedSnapshotHandler)
}

// Language-specific handlers

func (eis *EnhancedIntelligenceServer) languagesHandler(c *fiber.Ctx) error {
	languages := eis.epi.GetDetectedLanguages()
	return c.JSON(fiber.Map{
		"languages": languages,
		"count":     len(languages),
		"timestamp": time.Now(),
	})
}

func (eis *EnhancedIntelligenceServer) languageErrorsHandler(c *fiber.Ctx) error {
	language := c.Params("language")
	errors, err := eis.epi.AnalyzeLanguageErrors(language)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"language": language,
		"errors":   errors,
		"count":    len(errors),
	})
}

func (eis *EnhancedIntelligenceServer) languageDependenciesHandler(c *fiber.Ctx) error {
	language := c.Params("language")
	deps, err := eis.epi.GetLanguageDependencies(language)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"language":     language,
		"dependencies": deps,
		"count":        len(deps),
	})
}

func (eis *EnhancedIntelligenceServer) languageServicesHandler(c *fiber.Ctx) error {
	language := c.Params("language")
	services, err := eis.epi.FindLanguageServices(language)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"language": language,
		"services": services,
		"count":    len(services),
	})
}

func (eis *EnhancedIntelligenceServer) languageLintHandler(c *fiber.Ctx) error {
	language := c.Params("language")
	errors, err := eis.epi.RunLanguageLinter(language)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"language":    language,
		"lint_errors": errors,
		"count":       len(errors),
		"timestamp":   time.Now(),
	})
}

func (eis *EnhancedIntelligenceServer) languageTestHandler(c *fiber.Ctx) error {
	language := c.Params("language")
	results, err := eis.epi.RunLanguageTests(language)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"language":     language,
		"test_results": results,
		"timestamp":    time.Now(),
	})
}

func (eis *EnhancedIntelligenceServer) topologyHandler(c *fiber.Ctx) error {
	topology := eis.epi.GetProjectTopology()
	return c.JSON(topology)
}

func (eis *EnhancedIntelligenceServer) enhancedSnapshotHandler(c *fiber.Ctx) error {
	snapshot := eis.epi.GetEnhancedSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Enhanced snapshot not ready yet",
		})
	}

	return c.JSON(snapshot)
}

func (eis *EnhancedIntelligenceServer) analyzeAllLanguagesHandler(c *fiber.Ctx) error {
	// Trigger analysis for all detected languages
	go func() {
		eis.epi.analyzeLanguageErrors()
		eis.epi.updateEnhancedSnapshot()
	}()

	return c.JSON(fiber.Map{
		"message":   "Multi-language analysis started",
		"timestamp": time.Now(),
	})
}

func (eis *EnhancedIntelligenceServer) projectOverviewHandler(c *fiber.Ctx) error {
	languages := eis.epi.GetDetectedLanguages()
	topology := eis.epi.GetProjectTopology()
	baseSnapshot := eis.epi.GetSnapshot()

	if baseSnapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}

	// Create comprehensive project overview
	overview := fiber.Map{
		"project_type": "multi-language",
		"languages": fiber.Map{
			"detected": languages,
			"count":    len(languages),
			"primary":  getPrimaryLanguage(languages),
		},
		"health": baseSnapshot.Health,
		"structure": fiber.Map{
			"total_files": baseSnapshot.Structure.TotalFiles,
			"total_size":  baseSnapshot.Structure.TotalSize,
		},
		"services": fiber.Map{
			"detected": topology.Services,
			"count":    len(topology.Services),
		},
		"errors": fiber.Map{
			"active": baseSnapshot.ActiveErrors,
			"count":  len(baseSnapshot.ActiveErrors),
		},
		"last_updated": topology.UpdatedAt,
	}

	return c.JSON(overview)
}

// Helper function to determine primary language
func getPrimaryLanguage(languages []DetectedLanguage) string {
	if len(languages) == 0 {
		return "unknown"
	}

	// Return the language with the most files
	maxFiles := 0
	primary := "unknown"

	for _, lang := range languages {
		if lang.FileCount > maxFiles {
			maxFiles = lang.FileCount
			primary = lang.Language.Name
		}
	}

	return primary
}

// Enhanced topology handlers for mind map visualization

func (eis *EnhancedIntelligenceServer) detailedTopologyHandler(c *fiber.Ctx) error {
	topology := eis.epi.GetProjectTopology()
	languages := eis.epi.GetDetectedLanguages()

	// Create detailed topology with relationships
	detailedTopology := fiber.Map{
		"languages":      languages,
		"services":       topology.Services,
		"databases":      topology.Databases,
		"apis":           topology.APIs,
		"relationships":  eis.generateRelationships(languages, topology.Services),
		"dependencies":   eis.getAllDependencies(),
		"health_metrics": eis.getHealthMetrics(),
		"updated_at":     topology.UpdatedAt,
	}

	return c.JSON(detailedTopology)
}

func (eis *EnhancedIntelligenceServer) mindmapDataHandler(c *fiber.Ctx) error {
	view := c.Query("view", "human") // human or ai

	languages := eis.epi.GetDetectedLanguages()
	topology := eis.epi.GetProjectTopology()
	snapshot := eis.epi.GetEnhancedSnapshot()

	// Create mind map specific data structure
	mindmapData := fiber.Map{
		"view":  view,
		"nodes": eis.generateMindmapNodes(languages, topology, snapshot, view),
		"links": eis.generateMindmapLinks(languages, topology, snapshot, view),
		"metadata": fiber.Map{
			"total_files":    len(snapshot.Structure.Files),
			"health_score":   snapshot.Health.Score,
			"error_count":    len(snapshot.ActiveErrors),
			"language_count": len(languages),
			"service_count":  len(topology.Services),
		},
		"timestamp": time.Now(),
	}

	return c.JSON(mindmapData)
}

func (eis *EnhancedIntelligenceServer) mindmapNodesHandler(c *fiber.Ctx) error {
	view := c.Query("view", "human")

	languages := eis.epi.GetDetectedLanguages()
	topology := eis.epi.GetProjectTopology()
	snapshot := eis.epi.GetEnhancedSnapshot()

	nodes := eis.generateMindmapNodes(languages, topology, snapshot, view)

	return c.JSON(fiber.Map{
		"nodes": nodes,
		"count": len(nodes),
	})
}

func (eis *EnhancedIntelligenceServer) mindmapRelationshipsHandler(c *fiber.Ctx) error {
	view := c.Query("view", "human")

	languages := eis.epi.GetDetectedLanguages()
	topology := eis.epi.GetProjectTopology()
	snapshot := eis.epi.GetEnhancedSnapshot()

	links := eis.generateMindmapLinks(languages, topology, snapshot, view)

	return c.JSON(fiber.Map{
		"relationships": links,
		"count":         len(links),
	})
}

func (eis *EnhancedIntelligenceServer) mindmapPageHandler(c *fiber.Ctx) error {
	return c.SendFile("./mindmap.html")
}

func (eis *EnhancedIntelligenceServer) snapshotsPageHandler(c *fiber.Ctx) error {
	return c.SendFile("./snapshots.html")
}

func (eis *EnhancedIntelligenceServer) enhancedDashboardHandler(c *fiber.Ctx) error {
	// Return enhanced dashboard showing new features
	enhancedEndpoints := []string{
		"/api/enhanced-snapshot - Enhanced project snapshot with language detection",
		"/api/languages - Multi-language project analysis",
		"/api/topology/mindmap - Interactive mind-map visualization data",
		"/api/snapshots - Investigation snapshot management",
		"/api/investigation/state - Current investigation tracking",
		"/snapshots - Investigation snapshots dashboard",
		"/mindmap - Interactive project mind-map",
		"/api/topology - Complete project topology",
		"--- Original Argus Endpoints ---",
		"/api/snapshot - Complete project snapshot",
		"/api/structure - Project file structure",
		"/api/changes - Recent file changes",
		"/api/git - Git repository status",
		"/api/errors - Active errors and warnings",
		"/api/build - Build status",
		"/api/processes - Running processes",
		"/api/dependencies - Project dependencies",
		"/api/todos - TODO items in code",
		"/api/health - Project health metrics",
		"/api/search?q=query - Search across project",
	}

	snapshot := eis.epi.GetEnhancedSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Enhanced snapshot not ready"})
	}

	return c.JSON(fiber.Map{
		"service":   "Enhanced Project Argus",
		"version":   "2.0.0",
		"status":    "running",
		"workspace": eis.epi.workspace,
		"timestamp": time.Now(),
		"features": []string{
			"Multi-language detection (9 languages)",
			"Investigation snapshot system",
			"Interactive mind-map visualization",
			"Human-AI collaboration tracking",
			"Export & sharing capabilities",
			"Version management & lineage",
		},
		"health":    snapshot.Health,
		"languages": len(snapshot.Languages),
		"endpoints": enhancedEndpoints,
		"dashboards": fiber.Map{
			"snapshots": "/snapshots - Investigation management dashboard",
			"mindmap":   "/mindmap - Interactive project visualization",
		},
	})
}

// Helper functions for mind map data generation

func (eis *EnhancedIntelligenceServer) generateMindmapNodes(languages []DetectedLanguage, topology *ProjectTopology, snapshot *EnhancedProjectSnapshot, view string) []fiber.Map {
	nodes := []fiber.Map{}

	// Project root node
	nodes = append(nodes, fiber.Map{
		"id":    "project",
		"type":  "project",
		"name":  "Project",
		"size":  40,
		"color": "#64ffda",
		"metadata": fiber.Map{
			"health_score": snapshot.Health.Score,
			"total_files":  len(snapshot.Structure.Files),
			"total_size":   snapshot.Structure.TotalSize,
		},
	})

	// Language nodes
	for _, lang := range languages {
		nodeColor := eis.getLanguageColor(lang.Language.Name)
		nodeSize := eis.calculateNodeSize(lang.FileCount, 20, 35)

		node := fiber.Map{
			"id":    fmt.Sprintf("lang-%s", lang.Language.Name),
			"type":  "language",
			"name":  lang.Language.Name,
			"size":  nodeSize,
			"color": nodeColor,
			"metadata": fiber.Map{
				"file_count": lang.FileCount,
				"line_count": lang.LineCount,
				"frameworks": lang.Frameworks,
				"extensions": lang.Language.Extensions,
			},
		}
		nodes = append(nodes, node)
	}

	// Service nodes
	for _, service := range topology.Services {
		statusColor := "#ff9800" // default warning
		if service.Status == "running" {
			statusColor = "#4caf50"
		} else if service.Status == "error" {
			statusColor = "#f44336"
		}

		node := fiber.Map{
			"id":    fmt.Sprintf("service-%s", service.ID),
			"type":  "service",
			"name":  service.Name,
			"size":  25,
			"color": statusColor,
			"metadata": fiber.Map{
				"language":      service.Language,
				"framework":     service.Framework,
				"port":          service.Port,
				"status":        service.Status,
				"start_command": service.StartCommand,
			},
		}
		nodes = append(nodes, node)
	}

	// Error nodes (AI view only)
	if view == "ai" && len(snapshot.ActiveErrors) > 0 {
		errorGroups := eis.groupErrorsByFile(snapshot.ActiveErrors)
		for file, errors := range errorGroups {
			nodeSize := eis.calculateNodeSize(len(errors), 15, 30)

			node := fiber.Map{
				"id":    fmt.Sprintf("error-%s", file),
				"type":  "error",
				"name":  fmt.Sprintf("%s (%d errors)", file, len(errors)),
				"size":  nodeSize,
				"color": "#f44336",
				"metadata": fiber.Map{
					"file":        file,
					"error_count": len(errors),
					"errors":      errors[:min(len(errors), 5)], // Limit to first 5 errors
				},
			}
			nodes = append(nodes, node)
		}
	}

	// Dependency nodes (AI view only)
	if view == "ai" {
		deps := eis.getAllDependencies()
		depGroups := eis.groupDependenciesByLanguage(deps)

		for language, langDeps := range depGroups {
			if len(langDeps) > 5 { // Only show if significant dependencies
				node := fiber.Map{
					"id":    fmt.Sprintf("deps-%s", language),
					"type":  "dependencies",
					"name":  fmt.Sprintf("%s Dependencies (%d)", language, len(langDeps)),
					"size":  eis.calculateNodeSize(len(langDeps), 18, 28),
					"color": "#4caf50",
					"metadata": fiber.Map{
						"language":         language,
						"dependency_count": len(langDeps),
						"dependencies":     langDeps[:min(len(langDeps), 10)], // Limit to first 10
					},
				}
				nodes = append(nodes, node)
			}
		}
	}

	return nodes
}

func (eis *EnhancedIntelligenceServer) generateMindmapLinks(languages []DetectedLanguage, topology *ProjectTopology, snapshot *EnhancedProjectSnapshot, view string) []fiber.Map {
	links := []fiber.Map{}

	// Connect languages to project
	for _, lang := range languages {
		link := fiber.Map{
			"source":   "project",
			"target":   fmt.Sprintf("lang-%s", lang.Language.Name),
			"type":     "hierarchy",
			"strength": 1.0,
		}
		links = append(links, link)
	}

	// Connect services to their languages
	for _, service := range topology.Services {
		langTarget := fmt.Sprintf("lang-%s", service.Language)
		serviceTarget := fmt.Sprintf("service-%s", service.ID)

		link := fiber.Map{
			"source":   langTarget,
			"target":   serviceTarget,
			"type":     "association",
			"strength": 0.8,
		}
		links = append(links, link)
	}

	// Connect errors to languages (AI view)
	if view == "ai" {
		errorGroups := eis.groupErrorsByFile(snapshot.ActiveErrors)
		for file, _ := range errorGroups {
			language := eis.detectLanguageFromFile(file)
			langTarget := fmt.Sprintf("lang-%s", language)
			errorTarget := fmt.Sprintf("error-%s", file)

			link := fiber.Map{
				"source":   langTarget,
				"target":   errorTarget,
				"type":     "error",
				"strength": 0.6,
			}
			links = append(links, link)
		}
	}

	// Connect dependencies to languages (AI view)
	if view == "ai" {
		deps := eis.getAllDependencies()
		depGroups := eis.groupDependenciesByLanguage(deps)

		for language, langDeps := range depGroups {
			if len(langDeps) > 5 {
				langTarget := fmt.Sprintf("lang-%s", language)
				depTarget := fmt.Sprintf("deps-%s", language)

				link := fiber.Map{
					"source":   langTarget,
					"target":   depTarget,
					"type":     "dependency",
					"strength": 0.7,
				}
				links = append(links, link)
			}
		}
	}

	return links
}

// Helper utility functions

func (eis *EnhancedIntelligenceServer) generateRelationships(languages []DetectedLanguage, services []ServiceInfo) []fiber.Map {
	relationships := []fiber.Map{}

	for _, service := range services {
		// Find the language this service belongs to
		for _, lang := range languages {
			if lang.Language.Name == service.Language {
				relationship := fiber.Map{
					"type":        "implements",
					"from":        fmt.Sprintf("language-%s", lang.Language.Name),
					"to":          fmt.Sprintf("service-%s", service.ID),
					"description": fmt.Sprintf("%s service using %s framework", service.Name, service.Framework),
				}
				relationships = append(relationships, relationship)
				break
			}
		}
	}

	return relationships
}

func (eis *EnhancedIntelligenceServer) getAllDependencies() []DependencyInfo {
	deps := []DependencyInfo{}
	languages := eis.epi.GetDetectedLanguages()

	for _, lang := range languages {
		langDeps, err := eis.epi.GetLanguageDependencies(lang.Language.Name)
		if err == nil {
			deps = append(deps, langDeps...)
		}
	}

	return deps
}

func (eis *EnhancedIntelligenceServer) getHealthMetrics() fiber.Map {
	snapshot := eis.epi.GetEnhancedSnapshot()
	if snapshot == nil {
		return fiber.Map{"score": 0, "status": "unknown"}
	}

	return fiber.Map{
		"score":          snapshot.Health.Score,
		"error_count":    snapshot.Health.ErrorCount,
		"warning_count":  snapshot.Health.WarningCount,
		"technical_debt": snapshot.Health.TechnicalDebt,
		"last_check":     snapshot.Health.LastHealthCheck,
	}
}

func (eis *EnhancedIntelligenceServer) groupErrorsByFile(errors []ErrorInfo) map[string][]ErrorInfo {
	groups := make(map[string][]ErrorInfo)
	for _, err := range errors {
		file := err.File
		if file == "" {
			file = "unknown"
		}
		groups[file] = append(groups[file], err)
	}
	return groups
}

func (eis *EnhancedIntelligenceServer) groupDependenciesByLanguage(deps []DependencyInfo) map[string][]DependencyInfo {
	groups := make(map[string][]DependencyInfo)
	for _, dep := range deps {
		language := eis.detectLanguageFromSource(dep.Source)
		groups[language] = append(groups[language], dep)
	}
	return groups
}

func (eis *EnhancedIntelligenceServer) getLanguageColor(language string) string {
	colors := map[string]string{
		"javascript": "#f7df1e",
		"typescript": "#3178c6",
		"python":     "#3776ab",
		"go":         "#00add8",
		"java":       "#ed8b00",
		"csharp":     "#239120",
		"rust":       "#000000",
		"php":        "#777bb4",
		"ruby":       "#cc342d",
	}

	if color, exists := colors[strings.ToLower(language)]; exists {
		return color
	}
	return "#64ffda" // default color
}

func (eis *EnhancedIntelligenceServer) calculateNodeSize(count, minSize, maxSize int) int {
	size := minSize + (count * 2)
	if size > maxSize {
		return maxSize
	}
	return size
}

func (eis *EnhancedIntelligenceServer) detectLanguageFromFile(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	extMap := map[string]string{
		".js":   "javascript",
		".jsx":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".py":   "python",
		".go":   "go",
		".java": "java",
		".cs":   "csharp",
		".rs":   "rust",
		".php":  "php",
		".rb":   "ruby",
	}

	if lang, exists := extMap[ext]; exists {
		return lang
	}
	return "unknown"
}

func (eis *EnhancedIntelligenceServer) detectLanguageFromSource(source string) string {
	sourceMap := map[string]string{
		"package.json":     "javascript",
		"go.mod":           "go",
		"requirements.txt": "python",
		"Cargo.toml":       "rust",
		"pom.xml":          "java",
		"composer.json":    "php",
		"Gemfile":          "ruby",
	}

	if lang, exists := sourceMap[source]; exists {
		return lang
	}
	return "unknown"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Snapshot system handlers

func (eis *EnhancedIntelligenceServer) createSnapshotHandler(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Snapshot name is required"})
	}

	// Get current project state
	enhancedSnapshot := eis.epi.GetEnhancedSnapshot()
	if enhancedSnapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Project state not ready"})
	}

	// Create snapshot
	snapshot, err := eis.snapshotManager.CreateSnapshot(req.Name, req.Description, eis.epi.workspace, enhancedSnapshot)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to create snapshot: %v", err)})
	}

	// Save to storage
	if err := eis.snapshotManager.SaveSnapshot(snapshot); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to save snapshot: %v", err)})
	}

	// Register with version manager
	eis.versionManager.RegisterSnapshot(snapshot)

	return c.Status(201).JSON(fiber.Map{
		"message":  "Snapshot created successfully",
		"snapshot": snapshot,
	})
}

func (eis *EnhancedIntelligenceServer) listSnapshotsHandler(c *fiber.Ctx) error {
	snapshots, err := eis.snapshotManager.ListSnapshots()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to list snapshots: %v", err)})
	}

	// Return lightweight snapshot info
	snapshotSummaries := make([]fiber.Map, len(snapshots))
	for i, snapshot := range snapshots {
		snapshotSummaries[i] = fiber.Map{
			"id":          snapshot.ID,
			"name":        snapshot.Name,
			"description": snapshot.Description,
			"created_at":  snapshot.CreatedAt,
			"updated_at":  snapshot.UpdatedAt,
			"version":     snapshot.Version,
			"tags":        snapshot.Metadata.Tags,
			"size":        snapshot.Metadata.Size,
		}
	}

	return c.JSON(fiber.Map{
		"snapshots": snapshotSummaries,
		"count":     len(snapshots),
	})
}

func (eis *EnhancedIntelligenceServer) getSnapshotHandler(c *fiber.Ctx) error {
	snapshotID := c.Params("id")

	snapshot, err := eis.snapshotManager.LoadSnapshot(snapshotID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Snapshot not found"})
	}

	return c.JSON(snapshot)
}

func (eis *EnhancedIntelligenceServer) updateSnapshotHandler(c *fiber.Ctx) error {
	snapshotID := c.Params("id")

	snapshot, err := eis.snapshotManager.LoadSnapshot(snapshotID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Snapshot not found"})
	}

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update allowed fields
	if name, ok := updates["name"].(string); ok {
		snapshot.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		snapshot.Description = description
	}
	if tags, ok := updates["tags"].([]interface{}); ok {
		stringTags := make([]string, len(tags))
		for i, tag := range tags {
			if s, ok := tag.(string); ok {
				stringTags[i] = s
			}
		}
		snapshot.Metadata.Tags = stringTags
	}

	// Save updates
	if err := eis.snapshotManager.SaveSnapshot(snapshot); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to update snapshot: %v", err)})
	}

	return c.JSON(fiber.Map{
		"message":  "Snapshot updated successfully",
		"snapshot": snapshot,
	})
}

func (eis *EnhancedIntelligenceServer) deleteSnapshotHandler(c *fiber.Ctx) error {
	snapshotID := c.Params("id")

	if err := eis.snapshotManager.DeleteSnapshot(snapshotID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to delete snapshot: %v", err)})
	}

	return c.JSON(fiber.Map{"message": "Snapshot deleted successfully"})
}

func (eis *EnhancedIntelligenceServer) restoreSnapshotHandler(c *fiber.Ctx) error {
	snapshotID := c.Params("id")

	snapshot, err := eis.snapshotManager.LoadSnapshot(snapshotID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Snapshot not found"})
	}

	// This would restore the project state from the snapshot
	// For now, just return the snapshot data
	return c.JSON(fiber.Map{
		"message":        "Snapshot restoration initiated",
		"snapshot_state": snapshot.ProjectState,
		"investigation":  snapshot.Investigation,
		"collaboration":  snapshot.Collaboration,
	})
}

func (eis *EnhancedIntelligenceServer) forkSnapshotHandler(c *fiber.Ctx) error {
	parentID := c.Params("id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Fork name is required"})
	}

	childSnapshot, err := eis.snapshotManager.CreateChildSnapshot(parentID, req.Name, req.Description)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to fork snapshot: %v", err)})
	}

	if err := eis.snapshotManager.SaveSnapshot(childSnapshot); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to save forked snapshot: %v", err)})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":  "Snapshot forked successfully",
		"snapshot": childSnapshot,
	})
}

// Investigation tracking handlers

func (eis *EnhancedIntelligenceServer) addQuestionHandler(c *fiber.Ctx) error {
	var req struct {
		Question string `json:"question"`
		Priority string `json:"priority"`
		Source   string `json:"source"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Question == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Question is required"})
	}

	if req.Source == "" {
		req.Source = "human"
	}

	question := eis.investigationTracker.AddQuestion(req.Question, req.Priority, req.Source)

	return c.JSON(fiber.Map{
		"message":  "Question added to investigation",
		"question": question,
	})
}

func (eis *EnhancedIntelligenceServer) updateQuestionHandler(c *fiber.Ctx) error {
	questionID := c.Params("id")

	var req struct {
		Status string `json:"status"`
		Answer string `json:"answer"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := eis.investigationTracker.UpdateQuestion(questionID, req.Status, req.Answer); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":     "Question updated",
		"question_id": questionID,
		"updates":     req,
	})
}

func (eis *EnhancedIntelligenceServer) addHypothesisHandler(c *fiber.Ctx) error {
	var req struct {
		Statement  string   `json:"statement"`
		Confidence float64  `json:"confidence"`
		Evidence   []string `json:"evidence"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Statement == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Statement is required"})
	}

	hypothesis := eis.investigationTracker.AddHypothesis(req.Statement, req.Confidence, req.Evidence)

	return c.JSON(fiber.Map{
		"message":    "Hypothesis added to investigation",
		"hypothesis": hypothesis,
	})
}

func (eis *EnhancedIntelligenceServer) updateHypothesisHandler(c *fiber.Ctx) error {
	hypothesisID := c.Params("id")

	var req struct {
		Status     string  `json:"status"`
		Confidence float64 `json:"confidence"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	return c.JSON(fiber.Map{
		"message":       "Hypothesis updated",
		"hypothesis_id": hypothesisID,
		"updates":       req,
	})
}

func (eis *EnhancedIntelligenceServer) addFindingHandler(c *fiber.Ctx) error {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Impact      string `json:"impact"`
		Category    string `json:"category"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Title is required"})
	}

	finding := eis.investigationTracker.AddFinding(req.Title, req.Description, req.Impact, req.Category, []Evidence{})

	return c.JSON(fiber.Map{
		"message": "Finding added to investigation",
		"finding": finding,
	})
}

func (eis *EnhancedIntelligenceServer) addBlockerHandler(c *fiber.Ctx) error {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Severity    string `json:"severity"`
		Type        string `json:"type"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Title is required"})
	}

	blocker := eis.investigationTracker.AddBlocker(req.Title, req.Description, req.Severity, req.Type)

	return c.JSON(fiber.Map{
		"message": "Blocker added to investigation",
		"blocker": blocker,
	})
}

func (eis *EnhancedIntelligenceServer) resolveBlockerHandler(c *fiber.Ctx) error {
	blockerID := c.Params("id")

	var req struct {
		Resolution string `json:"resolution"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	return c.JSON(fiber.Map{
		"message":    "Blocker resolved",
		"blocker_id": blockerID,
		"resolution": req.Resolution,
	})
}

// Collaboration handlers

func (eis *EnhancedIntelligenceServer) aiRecommendationHandler(c *fiber.Ctx) error {
	var req struct {
		Type        string  `json:"type"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Confidence  float64 `json:"confidence"`
		Priority    string  `json:"priority"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	return c.JSON(fiber.Map{
		"message":        "AI recommendation recorded",
		"recommendation": req,
	})
}

func (eis *EnhancedIntelligenceServer) humanDecisionHandler(c *fiber.Ctx) error {
	var req struct {
		Decision  string `json:"decision"`
		Rationale string `json:"rationale"`
		Impact    string `json:"impact"`
		Category  string `json:"category"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	return c.JSON(fiber.Map{
		"message":  "Human decision recorded",
		"decision": req,
	})
}

func (eis *EnhancedIntelligenceServer) communicationHandler(c *fiber.Ctx) error {
	var req struct {
		Source     string   `json:"source"`
		Type       string   `json:"type"`
		Content    string   `json:"content"`
		References []string `json:"references"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	return c.JSON(fiber.Map{
		"message":       "Communication logged",
		"communication": req,
	})
}

func (eis *EnhancedIntelligenceServer) getSessionHandler(c *fiber.Ctx) error {
	// Return current collaboration session info
	return c.JSON(fiber.Map{
		"session_id":       "current_session",
		"started_at":       time.Now().Add(-time.Hour), // Mock data
		"interactions":     15,
		"active_snapshots": 2,
		"current_phase":    "analysis",
	})
}

func (eis *EnhancedIntelligenceServer) getInvestigationStateHandler(c *fiber.Ctx) error {
	state := eis.investigationTracker.GetCurrentState()
	return c.JSON(fiber.Map{
		"investigation_state": state,
		"timestamp":           time.Now(),
	})
}

// Version management handlers

func (eis *EnhancedIntelligenceServer) getVersionInfoHandler(c *fiber.Ctx) error {
	snapshotID := c.Params("id")

	versionInfo, err := eis.versionManager.GetVersionInfo(snapshotID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(versionInfo)
}

func (eis *EnhancedIntelligenceServer) getAllVersionsHandler(c *fiber.Ctx) error {
	versions := eis.versionManager.GetAllVersions()
	return c.JSON(versions)
}

// Export and sharing handlers

func (eis *EnhancedIntelligenceServer) exportSnapshotHandler(c *fiber.Ctx) error {
	snapshotID := c.Params("id")

	var options ExportOptions
	if err := c.BodyParser(&options); err != nil {
		// Set default options if body parsing fails
		options = ExportOptions{
			Format:               FormatJSON,
			IncludeProjectState:  true,
			IncludeFiles:         false,
			IncludeTimeline:      true,
			IncludeCollaboration: true,
			CompressOutput:       false,
		}
	}

	data, filename, err := eis.snapshotExporter.ExportSnapshot(snapshotID, options)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Export failed: %v", err)})
	}

	// Set appropriate content type based on format
	var contentType string
	switch options.Format {
	case FormatJSON:
		contentType = "application/json"
	case FormatMarkdown:
		contentType = "text/markdown"
	case FormatHTML:
		contentType = "text/html"
	case FormatZIP:
		contentType = "application/zip"
	default:
		contentType = "application/octet-stream"
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	return c.Send(data)
}

func (eis *EnhancedIntelligenceServer) createShareLinkHandler(c *fiber.Ctx) error {
	snapshotID := c.Params("id")

	var options map[string]interface{}
	if err := c.BodyParser(&options); err != nil {
		options = make(map[string]interface{})
	}

	// Verify snapshot exists
	_, err := eis.snapshotManager.LoadSnapshot(snapshotID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Snapshot not found"})
	}

	link, err := eis.snapshotSharer.CreateShareableLink(snapshotID, options)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to create share link: %v", err)})
	}

	return c.JSON(fiber.Map{
		"message":   "Share link created successfully",
		"link":      link,
		"share_url": fmt.Sprintf("/api/shared/%s", link.ID),
	})
}

func (eis *EnhancedIntelligenceServer) accessSharedSnapshotHandler(c *fiber.Ctx) error {
	linkID := c.Params("linkId")

	link, err := eis.snapshotSharer.GetShareableLink(linkID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	// Check permissions
	if !link.Permissions["read"] {
		return c.Status(403).JSON(fiber.Map{"error": "Read permission denied"})
	}

	// Record access
	if err := eis.snapshotSharer.AccessShareableLink(linkID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to record access"})
	}

	// Load and return snapshot
	snapshot, err := eis.snapshotManager.LoadSnapshot(link.SnapshotID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Snapshot not found"})
	}

	// Return limited snapshot data based on permissions
	responseData := fiber.Map{
		"id":            snapshot.ID,
		"name":          snapshot.Name,
		"description":   snapshot.Description,
		"created_at":    snapshot.CreatedAt,
		"version":       snapshot.Version,
		"investigation": snapshot.Investigation,
		"permissions":   link.Permissions,
		"access_info": fiber.Map{
			"access_count": link.AccessCount,
			"expires_at":   link.ExpiresAt,
		},
	}

	// Include additional data based on permissions
	if link.Permissions["download"] {
		responseData["export_options"] = []string{"json", "markdown", "html", "zip"}
	}

	return c.JSON(responseData)
}
