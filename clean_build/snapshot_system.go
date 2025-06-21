package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// InvestigationSnapshot represents a saved state of human-AI collaboration
type InvestigationSnapshot struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Version          int       `json:"version"`
	ParentSnapshotID *string   `json:"parent_snapshot_id,omitempty"`

	// Core project state
	ProjectState *EnhancedProjectSnapshot `json:"project_state"`

	// Investigation context
	Investigation InvestigationContext `json:"investigation"`

	// Human-AI collaboration state
	Collaboration CollaborationState `json:"collaboration"`

	// Metadata and tags
	Metadata SnapshotMetadata `json:"metadata"`
}

// InvestigationContext captures the current investigation state
type InvestigationContext struct {
	Focus        string                  `json:"focus"`         // What is being investigated
	Goals        []string                `json:"goals"`         // Investigation objectives
	CurrentPhase string                  `json:"current_phase"` // discovery, analysis, implementation, testing
	Questions    []InvestigationQuestion `json:"questions"`     // Open questions
	Hypotheses   []Hypothesis            `json:"hypotheses"`    // Current theories
	Findings     []Finding               `json:"findings"`      // Discovered insights
	NextSteps    []string                `json:"next_steps"`    // Planned actions
	Blockers     []Blocker               `json:"blockers"`      // Current obstacles
}

// InvestigationQuestion represents an open question in the investigation
type InvestigationQuestion struct {
	ID         string     `json:"id"`
	Question   string     `json:"question"`
	Priority   string     `json:"priority"` // high, medium, low
	Status     string     `json:"status"`   // open, investigating, answered
	CreatedAt  time.Time  `json:"created_at"`
	AnsweredAt *time.Time `json:"answered_at,omitempty"`
	Answer     string     `json:"answer,omitempty"`
	Source     string     `json:"source"` // human, ai, analysis
}

// Hypothesis represents a theory about the problem/solution
type Hypothesis struct {
	ID         string     `json:"id"`
	Statement  string     `json:"statement"`
	Confidence float64    `json:"confidence"` // 0.0 to 1.0
	Evidence   []string   `json:"evidence"`   // Supporting evidence
	Status     string     `json:"status"`     // active, confirmed, refuted
	CreatedAt  time.Time  `json:"created_at"`
	TestedAt   *time.Time `json:"tested_at,omitempty"`
}

// Finding represents a discovered insight or fact
type Finding struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Impact      string     `json:"impact"`   // high, medium, low
	Category    string     `json:"category"` // bug, performance, architecture, etc.
	Evidence    []Evidence `json:"evidence"` // Supporting evidence
	CreatedAt   time.Time  `json:"created_at"`
	Verified    bool       `json:"verified"`
}

// Evidence represents supporting data for findings
type Evidence struct {
	Type      string    `json:"type"`    // code, log, metric, observation
	Source    string    `json:"source"`  // file path, API endpoint, etc.
	Content   string    `json:"content"` // The actual evidence
	Timestamp time.Time `json:"timestamp"`
}

// Blocker represents an obstacle in the investigation
type Blocker struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Severity    string     `json:"severity"` // critical, high, medium, low
	Type        string     `json:"type"`     // technical, environmental, knowledge
	Status      string     `json:"status"`   // open, investigating, resolved
	CreatedAt   time.Time  `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	Resolution  string     `json:"resolution,omitempty"`
}

// CollaborationState tracks human-AI interaction patterns
type CollaborationState struct {
	SessionID         string                 `json:"session_id"`
	InteractionCount  int                    `json:"interaction_count"`
	LastInteraction   time.Time              `json:"last_interaction"`
	AIRecommendations []AIRecommendation     `json:"ai_recommendations"`
	HumanDecisions    []HumanDecision        `json:"human_decisions"`
	SharedContext     map[string]interface{} `json:"shared_context"`
	CommunicationLog  []CommunicationEntry   `json:"communication_log"`
	ToolUsage         map[string]int         `json:"tool_usage"` // Track which tools are used most
}

// AIRecommendation represents AI suggestions
type AIRecommendation struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"` // investigation, action, tool, analysis
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Confidence  float64    `json:"confidence"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"` // pending, accepted, rejected, modified
	CreatedAt   time.Time  `json:"created_at"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
	Response    string     `json:"response,omitempty"`
}

// HumanDecision represents human choices and direction
type HumanDecision struct {
	ID        string    `json:"id"`
	Decision  string    `json:"decision"`
	Rationale string    `json:"rationale"`
	Impact    string    `json:"impact"`   // high, medium, low
	Category  string    `json:"category"` // direction, tool-choice, approach
	CreatedAt time.Time `json:"created_at"`
	Outcome   string    `json:"outcome,omitempty"`
}

// CommunicationEntry represents a message in human-AI dialogue
type CommunicationEntry struct {
	ID         string    `json:"id"`
	Source     string    `json:"source"` // human, ai
	Type       string    `json:"type"`   // question, answer, observation, instruction
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	References []string  `json:"references,omitempty"` // Referenced files, functions, etc.
}

// SnapshotMetadata contains additional snapshot information
type SnapshotMetadata struct {
	Tags             []string          `json:"tags"`
	ProjectPath      string            `json:"project_path"`
	GitCommit        string            `json:"git_commit,omitempty"`
	GitBranch        string            `json:"git_branch,omitempty"`
	Environment      map[string]string `json:"environment"` // OS, versions, etc.
	Size             int64             `json:"size"`        // Snapshot size in bytes
	Checksum         string            `json:"checksum"`
	RelatedSnapshots []string          `json:"related_snapshots"` // Related snapshot IDs
	ExportFormats    []string          `json:"export_formats"`    // Available export formats
}

// SnapshotManager handles snapshot operations
type SnapshotManager struct {
	storageDir string
	snapshots  map[string]*InvestigationSnapshot
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(storageDir string) *SnapshotManager {
	return &SnapshotManager{
		storageDir: storageDir,
		snapshots:  make(map[string]*InvestigationSnapshot),
	}
}

// CreateSnapshot creates a new investigation snapshot
func (sm *SnapshotManager) CreateSnapshot(name, description, workspace string, projectState *EnhancedProjectSnapshot) (*InvestigationSnapshot, error) {
	snapshot := &InvestigationSnapshot{
		ID:           generateSnapshotID(),
		Name:         name,
		Description:  description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Version:      1,
		ProjectState: projectState,
		Investigation: InvestigationContext{
			Focus:        "Initial investigation",
			Goals:        []string{},
			CurrentPhase: "discovery",
			Questions:    []InvestigationQuestion{},
			Hypotheses:   []Hypothesis{},
			Findings:     []Finding{},
			NextSteps:    []string{},
			Blockers:     []Blocker{},
		},
		Collaboration: CollaborationState{
			SessionID:         generateSessionID(),
			InteractionCount:  0,
			LastInteraction:   time.Now(),
			AIRecommendations: []AIRecommendation{},
			HumanDecisions:    []HumanDecision{},
			SharedContext:     make(map[string]interface{}),
			CommunicationLog:  []CommunicationEntry{},
			ToolUsage:         make(map[string]int),
		},
		Metadata: SnapshotMetadata{
			Tags:             []string{"initial"},
			ProjectPath:      workspace,
			Environment:      getEnvironmentInfo(),
			RelatedSnapshots: []string{},
			ExportFormats:    []string{"json", "markdown", "html"},
		},
	}

	// Calculate size and checksum
	if err := sm.calculateMetadata(snapshot); err != nil {
		return nil, fmt.Errorf("failed to calculate metadata: %v", err)
	}

	// Store snapshot
	sm.snapshots[snapshot.ID] = snapshot

	return snapshot, nil
}

// SaveSnapshot persists a snapshot to storage
func (sm *SnapshotManager) SaveSnapshot(snapshot *InvestigationSnapshot) error {
	filename := filepath.Join(sm.storageDir, fmt.Sprintf("snapshot-%s.json", snapshot.ID))

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %v", err)
	}

	// Update metadata before saving
	snapshot.UpdatedAt = time.Now()
	if err := sm.calculateMetadata(snapshot); err != nil {
		return fmt.Errorf("failed to update metadata: %v", err)
	}

	return writeFile(filename, data)
}

// LoadSnapshot loads a snapshot from storage
func (sm *SnapshotManager) LoadSnapshot(snapshotID string) (*InvestigationSnapshot, error) {
	// Check memory cache first
	if snapshot, exists := sm.snapshots[snapshotID]; exists {
		return snapshot, nil
	}

	// Load from disk
	filename := filepath.Join(sm.storageDir, fmt.Sprintf("snapshot-%s.json", snapshotID))
	data, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot file: %v", err)
	}

	var snapshot InvestigationSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %v", err)
	}

	// Cache in memory
	sm.snapshots[snapshotID] = &snapshot

	return &snapshot, nil
}

// ListSnapshots returns all available snapshots
func (sm *SnapshotManager) ListSnapshots() ([]*InvestigationSnapshot, error) {
	// First, scan directory for snapshot files
	if err := sm.loadSnapshotsFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load snapshots from disk: %v", err)
	}

	snapshots := make([]*InvestigationSnapshot, 0, len(sm.snapshots))
	for _, snapshot := range sm.snapshots {
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

// loadSnapshotsFromDisk scans the storage directory and loads all snapshots
func (sm *SnapshotManager) loadSnapshotsFromDisk() error {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(sm.storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %v", err)
	}

	files, err := ioutil.ReadDir(sm.storageDir)
	if err != nil {
		return fmt.Errorf("failed to read storage directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			// Extract snapshot ID from filename
			if len(file.Name()) > 9 && file.Name()[:9] == "snapshot-" {
				snapshotID := file.Name()[9 : len(file.Name())-5] // Remove "snapshot-" prefix and ".json" suffix

				// Load snapshot if not already in memory
				if _, exists := sm.snapshots[snapshotID]; !exists {
					if _, err := sm.LoadSnapshot(snapshotID); err != nil {
						// Log error but continue loading other snapshots
						fmt.Printf("Warning: failed to load snapshot %s: %v\n", snapshotID, err)
					}
				}
			}
		}
	}

	return nil
}

// DeleteSnapshot removes a snapshot
func (sm *SnapshotManager) DeleteSnapshot(snapshotID string) error {
	filename := filepath.Join(sm.storageDir, fmt.Sprintf("snapshot-%s.json", snapshotID))

	if err := deleteFile(filename); err != nil {
		return fmt.Errorf("failed to delete snapshot file: %v", err)
	}

	delete(sm.snapshots, snapshotID)
	return nil
}

// CreateChildSnapshot creates a new snapshot based on an existing one
func (sm *SnapshotManager) CreateChildSnapshot(parentID, name, description string) (*InvestigationSnapshot, error) {
	parent, err := sm.LoadSnapshot(parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load parent snapshot: %v", err)
	}

	child := &InvestigationSnapshot{
		ID:               generateSnapshotID(),
		Name:             name,
		Description:      description,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Version:          parent.Version + 1,
		ParentSnapshotID: &parentID,
		ProjectState:     parent.ProjectState,  // Copy project state
		Investigation:    parent.Investigation, // Copy investigation context
		Collaboration:    parent.Collaboration, // Copy collaboration state
		Metadata:         parent.Metadata,      // Copy metadata
	}

	// Update collaboration session
	child.Collaboration.SessionID = generateSessionID()
	child.Collaboration.InteractionCount = 0
	child.Collaboration.LastInteraction = time.Now()

	// Update metadata
	child.Metadata.Tags = append([]string{fmt.Sprintf("v%d", child.Version)}, child.Metadata.Tags...)
	child.Metadata.RelatedSnapshots = append(child.Metadata.RelatedSnapshots, parentID)

	if err := sm.calculateMetadata(child); err != nil {
		return nil, fmt.Errorf("failed to calculate metadata: %v", err)
	}

	sm.snapshots[child.ID] = child
	return child, nil
}

// Helper functions

func generateSnapshotID() string {
	return fmt.Sprintf("snap_%d", time.Now().UnixNano())
}

func generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}

func getEnvironmentInfo() map[string]string {
	return map[string]string{
		"os":         "linux", // This would be detected dynamically
		"go_version": "1.21",  // This would be detected dynamically
		"timestamp":  time.Now().Format(time.RFC3339),
	}
}

func (sm *SnapshotManager) calculateMetadata(snapshot *InvestigationSnapshot) error {
	// Calculate actual size and checksum
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	snapshot.Metadata.Size = int64(len(data))

	// Calculate SHA-256 checksum
	hash := sha256.Sum256(data)
	snapshot.Metadata.Checksum = "sha256:" + hex.EncodeToString(hash[:])

	return nil
}

// File operations
func writeFile(filename string, data []byte) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	return ioutil.WriteFile(filename, data, 0644)
}

func readFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func deleteFile(filename string) error {
	return os.Remove(filename)
}
