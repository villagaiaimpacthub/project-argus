package main

import (
	"fmt"
	"sort"
	"time"
)

// SnapshotVersion tracks snapshot lineage and relationships
type SnapshotVersion struct {
	SnapshotID   string                 `json:"snapshot_id"`
	Version      int                    `json:"version"`
	ParentID     *string                `json:"parent_id,omitempty"`
	Children     []string               `json:"children"`
	CreatedAt    time.Time              `json:"created_at"`
	CreatedBy    string                 `json:"created_by"`
	ChangesSince *SnapshotDiff          `json:"changes_since,omitempty"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SnapshotDiff represents changes between snapshots
type SnapshotDiff struct {
	QuestionsAdded    int `json:"questions_added"`
	QuestionsAnswered int `json:"questions_answered"`
	HypothesesAdded   int `json:"hypotheses_added"`
	HypothesesTested  int `json:"hypotheses_tested"`
	FindingsAdded     int `json:"findings_added"`
	BlockersAdded     int `json:"blockers_added"`
	BlockersResolved  int `json:"blockers_resolved"`
	FilesChanged      int `json:"files_changed"`
	LinesChanged      int `json:"lines_changed"`
}

// SnapshotGraph represents the version graph of snapshots
type SnapshotGraph struct {
	versions map[string]*SnapshotVersion
	roots    []string // Root snapshots (no parent)
}

// NewSnapshotGraph creates a new snapshot graph
func NewSnapshotGraph() *SnapshotGraph {
	return &SnapshotGraph{
		versions: make(map[string]*SnapshotVersion),
		roots:    []string{},
	}
}

// AddSnapshot adds a snapshot to the version graph
func (sg *SnapshotGraph) AddSnapshot(snapshot *InvestigationSnapshot) {
	version := &SnapshotVersion{
		SnapshotID: snapshot.ID,
		Version:    snapshot.Version,
		ParentID:   snapshot.ParentSnapshotID,
		Children:   []string{},
		CreatedAt:  snapshot.CreatedAt,
		CreatedBy:  "human", // This could be enhanced to track actual user
		Tags:       snapshot.Metadata.Tags,
		Metadata: map[string]interface{}{
			"size":         snapshot.Metadata.Size,
			"checksum":     snapshot.Metadata.Checksum,
			"project_path": snapshot.Metadata.ProjectPath,
		},
	}

	// Calculate diff if this has a parent
	if snapshot.ParentSnapshotID != nil {
		if parent, exists := sg.versions[*snapshot.ParentSnapshotID]; exists {
			version.ChangesSince = sg.calculateDiff(*snapshot.ParentSnapshotID, snapshot.ID)
			// Add this as a child to the parent
			parent.Children = append(parent.Children, snapshot.ID)
		}
	}

	sg.versions[snapshot.ID] = version

	// Update roots list
	if snapshot.ParentSnapshotID == nil {
		sg.roots = append(sg.roots, snapshot.ID)
	}
}

// calculateDiff calculates the differences between two snapshots
func (sg *SnapshotGraph) calculateDiff(parentID, childID string) *SnapshotDiff {
	// This is a simplified version - in practice, you would load both snapshots
	// and compare their investigation contexts
	return &SnapshotDiff{
		QuestionsAdded:    1, // Mock data
		QuestionsAnswered: 0,
		HypothesesAdded:   0,
		HypothesesTested:  0,
		FindingsAdded:     1,
		BlockersAdded:     0,
		BlockersResolved:  0,
		FilesChanged:      5,
		LinesChanged:      120,
	}
}

// GetLineage returns the full lineage (path from root to snapshot)
func (sg *SnapshotGraph) GetLineage(snapshotID string) ([]*SnapshotVersion, error) {
	version, exists := sg.versions[snapshotID]
	if !exists {
		return nil, fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	var lineage []*SnapshotVersion
	current := version

	// Build lineage by following parent pointers
	for current != nil {
		lineage = append([]*SnapshotVersion{current}, lineage...) // Prepend

		if current.ParentID == nil {
			break
		}

		parent, exists := sg.versions[*current.ParentID]
		if !exists {
			break
		}
		current = parent
	}

	return lineage, nil
}

// GetDescendants returns all descendants of a snapshot
func (sg *SnapshotGraph) GetDescendants(snapshotID string) ([]*SnapshotVersion, error) {
	version, exists := sg.versions[snapshotID]
	if !exists {
		return nil, fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	var descendants []*SnapshotVersion
	sg.collectDescendants(version, &descendants)

	return descendants, nil
}

// collectDescendants recursively collects all descendants
func (sg *SnapshotGraph) collectDescendants(version *SnapshotVersion, descendants *[]*SnapshotVersion) {
	for _, childID := range version.Children {
		if child, exists := sg.versions[childID]; exists {
			*descendants = append(*descendants, child)
			sg.collectDescendants(child, descendants)
		}
	}
}

// GetBranches returns all branch points (snapshots with multiple children)
func (sg *SnapshotGraph) GetBranches() []*SnapshotVersion {
	var branches []*SnapshotVersion

	for _, version := range sg.versions {
		if len(version.Children) > 1 {
			branches = append(branches, version)
		}
	}

	return branches
}

// GetTimeline returns snapshots sorted by creation time
func (sg *SnapshotGraph) GetTimeline() []*SnapshotVersion {
	versions := make([]*SnapshotVersion, 0, len(sg.versions))
	for _, version := range sg.versions {
		versions = append(versions, version)
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].CreatedAt.Before(versions[j].CreatedAt)
	})

	return versions
}

// GetStats returns statistics about the version graph
func (sg *SnapshotGraph) GetStats() map[string]interface{} {
	branches := sg.GetBranches()

	// Calculate average version depth
	totalDepth := 0
	maxDepth := 0

	for _, version := range sg.versions {
		lineage, _ := sg.GetLineage(version.SnapshotID)
		depth := len(lineage)
		totalDepth += depth
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	avgDepth := 0.0
	if len(sg.versions) > 0 {
		avgDepth = float64(totalDepth) / float64(len(sg.versions))
	}

	return map[string]interface{}{
		"total_snapshots": len(sg.versions),
		"root_snapshots":  len(sg.roots),
		"branch_points":   len(branches),
		"max_depth":       maxDepth,
		"average_depth":   avgDepth,
		"total_changes":   sg.getTotalChanges(),
	}
}

// getTotalChanges calculates total changes across all snapshots
func (sg *SnapshotGraph) getTotalChanges() map[string]int {
	totals := map[string]int{
		"questions_added":    0,
		"questions_answered": 0,
		"hypotheses_added":   0,
		"hypotheses_tested":  0,
		"findings_added":     0,
		"blockers_added":     0,
		"blockers_resolved":  0,
	}

	for _, version := range sg.versions {
		if version.ChangesSince != nil {
			totals["questions_added"] += version.ChangesSince.QuestionsAdded
			totals["questions_answered"] += version.ChangesSince.QuestionsAnswered
			totals["hypotheses_added"] += version.ChangesSince.HypothesesAdded
			totals["hypotheses_tested"] += version.ChangesSince.HypothesesTested
			totals["findings_added"] += version.ChangesSince.FindingsAdded
			totals["blockers_added"] += version.ChangesSince.BlockersAdded
			totals["blockers_resolved"] += version.ChangesSince.BlockersResolved
		}
	}

	return totals
}

// VersionManager handles snapshot versioning operations
type VersionManager struct {
	graph           *SnapshotGraph
	snapshotManager *SnapshotManager
}

// NewVersionManager creates a new version manager
func NewVersionManager(snapshotManager *SnapshotManager) *VersionManager {
	return &VersionManager{
		graph:           NewSnapshotGraph(),
		snapshotManager: snapshotManager,
	}
}

// RegisterSnapshot registers a snapshot in the version graph
func (vm *VersionManager) RegisterSnapshot(snapshot *InvestigationSnapshot) {
	vm.graph.AddSnapshot(snapshot)
}

// GetVersionInfo returns version information for a snapshot
func (vm *VersionManager) GetVersionInfo(snapshotID string) (map[string]interface{}, error) {
	lineage, err := vm.graph.GetLineage(snapshotID)
	if err != nil {
		return nil, err
	}

	descendants, err := vm.graph.GetDescendants(snapshotID)
	if err != nil {
		return nil, err
	}

	version, exists := vm.graph.versions[snapshotID]
	if !exists {
		return nil, fmt.Errorf("version not found: %s", snapshotID)
	}

	return map[string]interface{}{
		"version":     version,
		"lineage":     lineage,
		"descendants": descendants,
		"is_branch":   len(version.Children) > 1,
		"is_leaf":     len(version.Children) == 0,
		"depth":       len(lineage),
	}, nil
}

// GetAllVersions returns the complete version graph
func (vm *VersionManager) GetAllVersions() map[string]interface{} {
	return map[string]interface{}{
		"graph":    vm.graph.versions,
		"roots":    vm.graph.roots,
		"branches": vm.graph.GetBranches(),
		"timeline": vm.graph.GetTimeline(),
		"stats":    vm.graph.GetStats(),
	}
}
