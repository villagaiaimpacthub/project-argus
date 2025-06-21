package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SnapshotExporter handles different export formats
type SnapshotExporter struct {
	snapshotManager *SnapshotManager
}

// NewSnapshotExporter creates a new snapshot exporter
func NewSnapshotExporter(snapshotManager *SnapshotManager) *SnapshotExporter {
	return &SnapshotExporter{
		snapshotManager: snapshotManager,
	}
}

// ExportFormat represents different export formats
type ExportFormat string

const (
	FormatJSON     ExportFormat = "json"
	FormatMarkdown ExportFormat = "markdown"
	FormatHTML     ExportFormat = "html"
	FormatZIP      ExportFormat = "zip"
	FormatPDF      ExportFormat = "pdf"
)

// ExportOptions configures export behavior
type ExportOptions struct {
	Format               ExportFormat `json:"format"`
	IncludeProjectState  bool         `json:"include_project_state"`
	IncludeFiles         bool         `json:"include_files"`
	IncludeTimeline      bool         `json:"include_timeline"`
	IncludeCollaboration bool         `json:"include_collaboration"`
	CompressOutput       bool         `json:"compress_output"`
}

// ExportSnapshot exports a snapshot in the specified format
func (se *SnapshotExporter) ExportSnapshot(snapshotID string, options ExportOptions) ([]byte, string, error) {
	snapshot, err := se.snapshotManager.LoadSnapshot(snapshotID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load snapshot: %v", err)
	}

	switch options.Format {
	case FormatJSON:
		return se.exportJSON(snapshot, options)
	case FormatMarkdown:
		return se.exportMarkdown(snapshot, options)
	case FormatHTML:
		return se.exportHTML(snapshot, options)
	case FormatZIP:
		return se.exportZIP(snapshot, options)
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", options.Format)
	}
}

// exportJSON exports as JSON
func (se *SnapshotExporter) exportJSON(snapshot *InvestigationSnapshot, options ExportOptions) ([]byte, string, error) {
	exportData := map[string]interface{}{
		"snapshot_id":   snapshot.ID,
		"name":          snapshot.Name,
		"description":   snapshot.Description,
		"created_at":    snapshot.CreatedAt,
		"updated_at":    snapshot.UpdatedAt,
		"version":       snapshot.Version,
		"investigation": snapshot.Investigation,
		"metadata":      snapshot.Metadata,
	}

	if options.IncludeProjectState {
		exportData["project_state"] = snapshot.ProjectState
	}

	if options.IncludeCollaboration {
		exportData["collaboration"] = snapshot.Collaboration
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	filename := fmt.Sprintf("snapshot_%s_%s.json", snapshot.ID, time.Now().Format("20060102_150405"))
	return data, filename, nil
}

// exportMarkdown exports as Markdown
func (se *SnapshotExporter) exportMarkdown(snapshot *InvestigationSnapshot, options ExportOptions) ([]byte, string, error) {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# Investigation Snapshot: %s\n\n", snapshot.Name))
	md.WriteString(fmt.Sprintf("**Created:** %s  \n", snapshot.CreatedAt.Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("**Updated:** %s  \n", snapshot.UpdatedAt.Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("**Version:** %d  \n", snapshot.Version))
	md.WriteString(fmt.Sprintf("**Description:** %s\n\n", snapshot.Description))

	// Investigation Context
	md.WriteString("## 🔍 Investigation Context\n\n")
	md.WriteString(fmt.Sprintf("**Focus:** %s  \n", snapshot.Investigation.Focus))
	md.WriteString(fmt.Sprintf("**Current Phase:** %s  \n", snapshot.Investigation.CurrentPhase))

	if len(snapshot.Investigation.Goals) > 0 {
		md.WriteString("\n### Goals\n")
		for i, goal := range snapshot.Investigation.Goals {
			md.WriteString(fmt.Sprintf("%d. %s\n", i+1, goal))
		}
	}

	// Questions
	if len(snapshot.Investigation.Questions) > 0 {
		md.WriteString("\n## ❓ Questions\n\n")
		for _, q := range snapshot.Investigation.Questions {
			status := "🔓"
			if q.Status == "answered" {
				status = "✅"
			} else if q.Status == "investigating" {
				status = "🔍"
			}

			md.WriteString(fmt.Sprintf("### %s %s\n", status, q.Question))
			md.WriteString(fmt.Sprintf("- **Priority:** %s\n", q.Priority))
			md.WriteString(fmt.Sprintf("- **Source:** %s\n", q.Source))
			md.WriteString(fmt.Sprintf("- **Created:** %s\n", q.CreatedAt.Format("2006-01-02 15:04")))

			if q.Answer != "" {
				md.WriteString(fmt.Sprintf("- **Answer:** %s\n", q.Answer))
			}
			md.WriteString("\n")
		}
	}

	// Hypotheses
	if len(snapshot.Investigation.Hypotheses) > 0 {
		md.WriteString("\n## 💡 Hypotheses\n\n")
		for _, h := range snapshot.Investigation.Hypotheses {
			status := "🔬"
			if h.Status == "confirmed" {
				status = "✅"
			} else if h.Status == "refuted" {
				status = "❌"
			}

			md.WriteString(fmt.Sprintf("### %s %s\n", status, h.Statement))
			md.WriteString(fmt.Sprintf("- **Confidence:** %.2f\n", h.Confidence))
			md.WriteString(fmt.Sprintf("- **Status:** %s\n", h.Status))
			md.WriteString(fmt.Sprintf("- **Created:** %s\n", h.CreatedAt.Format("2006-01-02 15:04")))

			if len(h.Evidence) > 0 {
				md.WriteString("- **Evidence:**\n")
				for _, evidence := range h.Evidence {
					md.WriteString(fmt.Sprintf("  - %s\n", evidence))
				}
			}
			md.WriteString("\n")
		}
	}

	// Findings
	if len(snapshot.Investigation.Findings) > 0 {
		md.WriteString("\n## 🎯 Findings\n\n")
		for _, f := range snapshot.Investigation.Findings {
			verified := "⏳"
			if f.Verified {
				verified = "✅"
			}

			md.WriteString(fmt.Sprintf("### %s %s\n", verified, f.Title))
			md.WriteString(fmt.Sprintf("**Description:** %s\n\n", f.Description))
			md.WriteString(fmt.Sprintf("- **Impact:** %s\n", f.Impact))
			md.WriteString(fmt.Sprintf("- **Category:** %s\n", f.Category))
			md.WriteString(fmt.Sprintf("- **Created:** %s\n", f.CreatedAt.Format("2006-01-02 15:04")))
			md.WriteString("\n")
		}
	}

	// Blockers
	if len(snapshot.Investigation.Blockers) > 0 {
		md.WriteString("\n## 🚧 Blockers\n\n")
		for _, b := range snapshot.Investigation.Blockers {
			status := "🔴"
			if b.Status == "resolved" {
				status = "✅"
			} else if b.Status == "investigating" {
				status = "🔍"
			}

			md.WriteString(fmt.Sprintf("### %s %s\n", status, b.Title))
			md.WriteString(fmt.Sprintf("**Description:** %s\n\n", b.Description))
			md.WriteString(fmt.Sprintf("- **Severity:** %s\n", b.Severity))
			md.WriteString(fmt.Sprintf("- **Type:** %s\n", b.Type))
			md.WriteString(fmt.Sprintf("- **Created:** %s\n", b.CreatedAt.Format("2006-01-02 15:04")))

			if b.Resolution != "" {
				md.WriteString(fmt.Sprintf("- **Resolution:** %s\n", b.Resolution))
			}
			md.WriteString("\n")
		}
	}

	// Next Steps
	if len(snapshot.Investigation.NextSteps) > 0 {
		md.WriteString("\n## 📋 Next Steps\n\n")
		for i, step := range snapshot.Investigation.NextSteps {
			md.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		md.WriteString("\n")
	}

	// Collaboration Summary
	if options.IncludeCollaboration {
		md.WriteString("\n## 🤝 Collaboration Summary\n\n")
		md.WriteString(fmt.Sprintf("- **Session ID:** %s\n", snapshot.Collaboration.SessionID))
		md.WriteString(fmt.Sprintf("- **Interactions:** %d\n", snapshot.Collaboration.InteractionCount))
		md.WriteString(fmt.Sprintf("- **Last Interaction:** %s\n", snapshot.Collaboration.LastInteraction.Format("2006-01-02 15:04")))
		md.WriteString(fmt.Sprintf("- **AI Recommendations:** %d\n", len(snapshot.Collaboration.AIRecommendations)))
		md.WriteString(fmt.Sprintf("- **Human Decisions:** %d\n", len(snapshot.Collaboration.HumanDecisions)))
	}

	// Project State Summary
	if options.IncludeProjectState && snapshot.ProjectState != nil {
		md.WriteString("\n## 📊 Project State\n\n")
		md.WriteString(fmt.Sprintf("- **Health Score:** %d/100\n", snapshot.ProjectState.Health.Score))
		md.WriteString(fmt.Sprintf("- **Total Files:** %d\n", len(snapshot.ProjectState.Structure.Files)))
		md.WriteString(fmt.Sprintf("- **Active Errors:** %d\n", len(snapshot.ProjectState.ActiveErrors)))
		md.WriteString(fmt.Sprintf("- **Languages:** %d\n", len(snapshot.ProjectState.Languages)))
	}

	md.WriteString("\n---\n")
	md.WriteString(fmt.Sprintf("*Generated by Project Argus on %s*\n", time.Now().Format("2006-01-02 15:04:05")))

	filename := fmt.Sprintf("snapshot_%s_%s.md", snapshot.ID, time.Now().Format("20060102_150405"))
	return []byte(md.String()), filename, nil
}

// exportHTML exports as HTML
func (se *SnapshotExporter) exportHTML(snapshot *InvestigationSnapshot, options ExportOptions) ([]byte, string, error) {
	// Convert markdown to HTML (simplified version)
	markdownData, _, err := se.exportMarkdown(snapshot, options)
	if err != nil {
		return nil, "", err
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Investigation Snapshot: %s</title>
    <style>
        body { font-family: 'Segoe UI', sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; line-height: 1.6; }
        h1, h2, h3 { color: #2c3e50; }
        h1 { border-bottom: 3px solid #3498db; padding-bottom: 10px; }
        h2 { border-bottom: 1px solid #ecf0f1; padding-bottom: 5px; margin-top: 30px; }
        .meta { background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .status-open { color: #f39c12; }
        .status-resolved { color: #27ae60; }
        .status-error { color: #e74c3c; }
        pre { background: #f4f4f4; padding: 15px; border-radius: 5px; overflow-x: auto; }
        blockquote { border-left: 4px solid #3498db; margin: 0; padding-left: 15px; color: #7f8c8d; }
    </style>
</head>
<body>
    <div class="content">
        %s
    </div>
</body>
</html>`, snapshot.Name, string(markdownData))

	filename := fmt.Sprintf("snapshot_%s_%s.html", snapshot.ID, time.Now().Format("20060102_150405"))
	return []byte(html), filename, nil
}

// exportZIP exports as ZIP archive
func (se *SnapshotExporter) exportZIP(snapshot *InvestigationSnapshot, options ExportOptions) ([]byte, string, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add JSON export
	jsonData, jsonFilename, err := se.exportJSON(snapshot, options)
	if err != nil {
		return nil, "", err
	}

	jsonFile, err := zipWriter.Create(jsonFilename)
	if err != nil {
		return nil, "", err
	}
	_, err = jsonFile.Write(jsonData)
	if err != nil {
		return nil, "", err
	}

	// Add Markdown export
	mdData, mdFilename, err := se.exportMarkdown(snapshot, options)
	if err != nil {
		return nil, "", err
	}

	mdFile, err := zipWriter.Create(mdFilename)
	if err != nil {
		return nil, "", err
	}
	_, err = mdFile.Write(mdData)
	if err != nil {
		return nil, "", err
	}

	// Add HTML export
	htmlData, htmlFilename, err := se.exportHTML(snapshot, options)
	if err != nil {
		return nil, "", err
	}

	htmlFile, err := zipWriter.Create(htmlFilename)
	if err != nil {
		return nil, "", err
	}
	_, err = htmlFile.Write(htmlData)
	if err != nil {
		return nil, "", err
	}

	// Add metadata file
	metadataFile, err := zipWriter.Create("metadata.json")
	if err != nil {
		return nil, "", err
	}

	metadataJSON, _ := json.MarshalIndent(map[string]interface{}{
		"export_timestamp": time.Now(),
		"export_options":   options,
		"snapshot_info": map[string]interface{}{
			"id":         snapshot.ID,
			"name":       snapshot.Name,
			"version":    snapshot.Version,
			"created_at": snapshot.CreatedAt,
			"updated_at": snapshot.UpdatedAt,
		},
	}, "", "  ")

	_, err = metadataFile.Write(metadataJSON)
	if err != nil {
		return nil, "", err
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("snapshot_%s_%s.zip", snapshot.ID, time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}

// ShareableLink represents a shareable snapshot link
type ShareableLink struct {
	ID          string            `json:"id"`
	SnapshotID  string            `json:"snapshot_id"`
	Token       string            `json:"token"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	Permissions map[string]bool   `json:"permissions"`
	CreatedAt   time.Time         `json:"created_at"`
	AccessCount int               `json:"access_count"`
	MaxAccesses *int              `json:"max_accesses,omitempty"`
	Metadata    map[string]string `json:"metadata"`
}

// SnapshotSharer handles sharing functionality
type SnapshotSharer struct {
	links map[string]*ShareableLink
}

// NewSnapshotSharer creates a new snapshot sharer
func NewSnapshotSharer() *SnapshotSharer {
	return &SnapshotSharer{
		links: make(map[string]*ShareableLink),
	}
}

// CreateShareableLink creates a shareable link for a snapshot
func (ss *SnapshotSharer) CreateShareableLink(snapshotID string, options map[string]interface{}) (*ShareableLink, error) {
	link := &ShareableLink{
		ID:         generateShareLinkID(),
		SnapshotID: snapshotID,
		Token:      generateShareToken(),
		Permissions: map[string]bool{
			"read":     true,
			"download": true,
			"fork":     false,
		},
		CreatedAt:   time.Now(),
		AccessCount: 0,
		Metadata:    make(map[string]string),
	}

	// Apply options
	if expiresIn, ok := options["expires_in"].(int); ok && expiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Hour)
		link.ExpiresAt = &expiresAt
	}

	if maxAccesses, ok := options["max_accesses"].(int); ok && maxAccesses > 0 {
		link.MaxAccesses = &maxAccesses
	}

	if permissions, ok := options["permissions"].(map[string]bool); ok {
		for k, v := range permissions {
			link.Permissions[k] = v
		}
	}

	ss.links[link.ID] = link
	return link, nil
}

// GetShareableLink retrieves a shareable link by ID
func (ss *SnapshotSharer) GetShareableLink(linkID string) (*ShareableLink, error) {
	link, exists := ss.links[linkID]
	if !exists {
		return nil, fmt.Errorf("shareable link not found: %s", linkID)
	}

	// Check if link has expired
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return nil, fmt.Errorf("shareable link has expired")
	}

	// Check if max accesses reached
	if link.MaxAccesses != nil && link.AccessCount >= *link.MaxAccesses {
		return nil, fmt.Errorf("shareable link access limit reached")
	}

	return link, nil
}

// AccessShareableLink records an access to a shareable link
func (ss *SnapshotSharer) AccessShareableLink(linkID string) error {
	link, err := ss.GetShareableLink(linkID)
	if err != nil {
		return err
	}

	link.AccessCount++
	return nil
}

// Helper functions for ID generation
func generateShareLinkID() string {
	return fmt.Sprintf("link_%d", time.Now().UnixNano())
}

func generateShareToken() string {
	return fmt.Sprintf("tok_%d", time.Now().UnixNano())
}
