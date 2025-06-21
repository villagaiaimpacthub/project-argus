package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// EnhancedProjectIntelligence extends ProjectIntelligence with multi-language support
type EnhancedProjectIntelligence struct {
	*ProjectIntelligence
	languageDetector *LanguageDetector
	pluginManager    *LanguagePluginManager
	detectedLanguages []DetectedLanguage
	projectTopology  *ProjectTopology
	mutex            sync.RWMutex
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
	epi *EnhancedProjectIntelligence
}

// NewEnhancedIntelligenceServer creates an enhanced intelligence server
func NewEnhancedIntelligenceServer(workspace string) *EnhancedIntelligenceServer {
	baseServer := NewIntelligenceServer(workspace)
	epi := NewEnhancedProjectIntelligence(workspace)
	
	server := &EnhancedIntelligenceServer{
		IntelligenceServer: baseServer,
		epi:                epi,
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
	eis.app.Get("/api/enhanced-snapshot", eis.enhancedSnapshotHandler)
	
	// Universal analysis endpoints
	eis.app.Post("/api/analyze/all-languages", eis.analyzeAllLanguagesHandler)
	eis.app.Get("/api/project-overview", eis.projectOverviewHandler)
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
		"language": language,
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
		"language": language,
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
		"message": "Multi-language analysis started",
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