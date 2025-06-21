package main

import (
	"fmt"
	"sync"
	"time"
)

// InvestigationTracker manages the active investigation state
type InvestigationTracker struct {
	currentSnapshot  *InvestigationSnapshot
	activeQuestions  map[string]*InvestigationQuestion
	activeHypotheses map[string]*Hypothesis
	activeFindings   map[string]*Finding
	activeBlockers   map[string]*Blocker
	sessionID        string
	mutex            sync.RWMutex
}

// NewInvestigationTracker creates a new investigation tracker
func NewInvestigationTracker() *InvestigationTracker {
	return &InvestigationTracker{
		activeQuestions:  make(map[string]*InvestigationQuestion),
		activeHypotheses: make(map[string]*Hypothesis),
		activeFindings:   make(map[string]*Finding),
		activeBlockers:   make(map[string]*Blocker),
		sessionID:        generateSessionID(),
	}
}

// SetCurrentSnapshot sets the active investigation snapshot
func (it *InvestigationTracker) SetCurrentSnapshot(snapshot *InvestigationSnapshot) {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	it.currentSnapshot = snapshot
	it.sessionID = snapshot.Collaboration.SessionID

	// Load investigation state into active maps
	for _, question := range snapshot.Investigation.Questions {
		it.activeQuestions[question.ID] = &question
	}

	for _, hypothesis := range snapshot.Investigation.Hypotheses {
		it.activeHypotheses[hypothesis.ID] = &hypothesis
	}

	for _, finding := range snapshot.Investigation.Findings {
		it.activeFindings[finding.ID] = &finding
	}

	for _, blocker := range snapshot.Investigation.Blockers {
		it.activeBlockers[blocker.ID] = &blocker
	}
}

// AddQuestion adds a new investigation question
func (it *InvestigationTracker) AddQuestion(question, priority, source string) *InvestigationQuestion {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	q := &InvestigationQuestion{
		ID:        generateQuestionID(),
		Question:  question,
		Priority:  priority,
		Status:    "open",
		CreatedAt: time.Now(),
		Source:    source,
	}

	it.activeQuestions[q.ID] = q

	// Update current snapshot if exists
	if it.currentSnapshot != nil {
		it.currentSnapshot.Investigation.Questions = append(it.currentSnapshot.Investigation.Questions, *q)
		it.currentSnapshot.UpdatedAt = time.Now()
	}

	return q
}

// UpdateQuestion updates an existing question
func (it *InvestigationTracker) UpdateQuestion(questionID, status, answer string) error {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	question, exists := it.activeQuestions[questionID]
	if !exists {
		return fmt.Errorf("question not found: %s", questionID)
	}

	if status != "" {
		question.Status = status
	}

	if answer != "" {
		question.Answer = answer
		now := time.Now()
		question.AnsweredAt = &now
	}

	// Update in current snapshot
	if it.currentSnapshot != nil {
		for i, q := range it.currentSnapshot.Investigation.Questions {
			if q.ID == questionID {
				it.currentSnapshot.Investigation.Questions[i] = *question
				break
			}
		}
		it.currentSnapshot.UpdatedAt = time.Now()
	}

	return nil
}

// AddHypothesis adds a new hypothesis
func (it *InvestigationTracker) AddHypothesis(statement string, confidence float64, evidence []string) *Hypothesis {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	h := &Hypothesis{
		ID:         generateHypothesisID(),
		Statement:  statement,
		Confidence: confidence,
		Evidence:   evidence,
		Status:     "active",
		CreatedAt:  time.Now(),
	}

	it.activeHypotheses[h.ID] = h

	// Update current snapshot
	if it.currentSnapshot != nil {
		it.currentSnapshot.Investigation.Hypotheses = append(it.currentSnapshot.Investigation.Hypotheses, *h)
		it.currentSnapshot.UpdatedAt = time.Now()
	}

	return h
}

// UpdateHypothesis updates an existing hypothesis
func (it *InvestigationTracker) UpdateHypothesis(hypothesisID, status string, confidence float64) error {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	hypothesis, exists := it.activeHypotheses[hypothesisID]
	if !exists {
		return fmt.Errorf("hypothesis not found: %s", hypothesisID)
	}

	if status != "" {
		hypothesis.Status = status
		if status == "confirmed" || status == "refuted" {
			now := time.Now()
			hypothesis.TestedAt = &now
		}
	}

	if confidence >= 0 && confidence <= 1 {
		hypothesis.Confidence = confidence
	}

	// Update in current snapshot
	if it.currentSnapshot != nil {
		for i, h := range it.currentSnapshot.Investigation.Hypotheses {
			if h.ID == hypothesisID {
				it.currentSnapshot.Investigation.Hypotheses[i] = *hypothesis
				break
			}
		}
		it.currentSnapshot.UpdatedAt = time.Now()
	}

	return nil
}

// AddFinding adds a new finding
func (it *InvestigationTracker) AddFinding(title, description, impact, category string, evidence []Evidence) *Finding {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	f := &Finding{
		ID:          generateFindingID(),
		Title:       title,
		Description: description,
		Impact:      impact,
		Category:    category,
		Evidence:    evidence,
		CreatedAt:   time.Now(),
		Verified:    false,
	}

	it.activeFindings[f.ID] = f

	// Update current snapshot
	if it.currentSnapshot != nil {
		it.currentSnapshot.Investigation.Findings = append(it.currentSnapshot.Investigation.Findings, *f)
		it.currentSnapshot.UpdatedAt = time.Now()
	}

	return f
}

// AddBlocker adds a new blocker
func (it *InvestigationTracker) AddBlocker(title, description, severity, blockerType string) *Blocker {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	b := &Blocker{
		ID:          generateBlockerID(),
		Title:       title,
		Description: description,
		Severity:    severity,
		Type:        blockerType,
		Status:      "open",
		CreatedAt:   time.Now(),
	}

	it.activeBlockers[b.ID] = b

	// Update current snapshot
	if it.currentSnapshot != nil {
		it.currentSnapshot.Investigation.Blockers = append(it.currentSnapshot.Investigation.Blockers, *b)
		it.currentSnapshot.UpdatedAt = time.Now()
	}

	return b
}

// ResolveBlocker resolves an existing blocker
func (it *InvestigationTracker) ResolveBlocker(blockerID, resolution string) error {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	blocker, exists := it.activeBlockers[blockerID]
	if !exists {
		return fmt.Errorf("blocker not found: %s", blockerID)
	}

	blocker.Status = "resolved"
	blocker.Resolution = resolution
	now := time.Now()
	blocker.ResolvedAt = &now

	// Update in current snapshot
	if it.currentSnapshot != nil {
		for i, b := range it.currentSnapshot.Investigation.Blockers {
			if b.ID == blockerID {
				it.currentSnapshot.Investigation.Blockers[i] = *blocker
				break
			}
		}
		it.currentSnapshot.UpdatedAt = time.Now()
	}

	return nil
}

// AddCommunication logs a communication entry
func (it *InvestigationTracker) AddCommunication(source, messageType, content string, references []string) {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	entry := CommunicationEntry{
		ID:         generateCommunicationID(),
		Source:     source,
		Type:       messageType,
		Content:    content,
		Timestamp:  time.Now(),
		References: references,
	}

	// Update current snapshot
	if it.currentSnapshot != nil {
		it.currentSnapshot.Collaboration.CommunicationLog = append(
			it.currentSnapshot.Collaboration.CommunicationLog, entry)
		it.currentSnapshot.Collaboration.InteractionCount++
		it.currentSnapshot.Collaboration.LastInteraction = time.Now()
		it.currentSnapshot.UpdatedAt = time.Now()
	}
}

// AddAIRecommendation records an AI recommendation
func (it *InvestigationTracker) AddAIRecommendation(recType, title, description string, confidence float64, priority string) *AIRecommendation {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	rec := &AIRecommendation{
		ID:          generateRecommendationID(),
		Type:        recType,
		Title:       title,
		Description: description,
		Confidence:  confidence,
		Priority:    priority,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	// Update current snapshot
	if it.currentSnapshot != nil {
		it.currentSnapshot.Collaboration.AIRecommendations = append(
			it.currentSnapshot.Collaboration.AIRecommendations, *rec)
		it.currentSnapshot.UpdatedAt = time.Now()
	}

	return rec
}

// AddHumanDecision records a human decision
func (it *InvestigationTracker) AddHumanDecision(decision, rationale, impact, category string) *HumanDecision {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	dec := &HumanDecision{
		ID:        generateDecisionID(),
		Decision:  decision,
		Rationale: rationale,
		Impact:    impact,
		Category:  category,
		CreatedAt: time.Now(),
	}

	// Update current snapshot
	if it.currentSnapshot != nil {
		it.currentSnapshot.Collaboration.HumanDecisions = append(
			it.currentSnapshot.Collaboration.HumanDecisions, *dec)
		it.currentSnapshot.UpdatedAt = time.Now()
	}

	return dec
}

// GetCurrentState returns the current investigation state
func (it *InvestigationTracker) GetCurrentState() map[string]interface{} {
	it.mutex.RLock()
	defer it.mutex.RUnlock()

	// Convert maps to slices for JSON serialization
	questions := make([]InvestigationQuestion, 0, len(it.activeQuestions))
	for _, q := range it.activeQuestions {
		questions = append(questions, *q)
	}

	hypotheses := make([]Hypothesis, 0, len(it.activeHypotheses))
	for _, h := range it.activeHypotheses {
		hypotheses = append(hypotheses, *h)
	}

	findings := make([]Finding, 0, len(it.activeFindings))
	for _, f := range it.activeFindings {
		findings = append(findings, *f)
	}

	blockers := make([]Blocker, 0, len(it.activeBlockers))
	for _, b := range it.activeBlockers {
		blockers = append(blockers, *b)
	}

	return map[string]interface{}{
		"session_id":        it.sessionID,
		"questions":         questions,
		"hypotheses":        hypotheses,
		"findings":          findings,
		"blockers":          blockers,
		"question_stats":    it.getQuestionStats(),
		"blocker_stats":     it.getBlockerStats(),
		"verified_findings": it.countVerified(),
	}
}

// Helper functions for counting states
func (it *InvestigationTracker) getQuestionStats() map[string]int {
	stats := make(map[string]int)
	for _, q := range it.activeQuestions {
		stats[q.Status]++
	}
	return stats
}

func (it *InvestigationTracker) getBlockerStats() map[string]int {
	stats := make(map[string]int)
	for _, b := range it.activeBlockers {
		stats[b.Status]++
	}
	return stats
}

func (it *InvestigationTracker) countVerified() int {
	count := 0
	for _, finding := range it.activeFindings {
		if finding.Verified {
			count++
		}
	}
	return count
}

// ID generation functions
func generateQuestionID() string {
	return fmt.Sprintf("q_%d", time.Now().UnixNano())
}

func generateHypothesisID() string {
	return fmt.Sprintf("h_%d", time.Now().UnixNano())
}

func generateFindingID() string {
	return fmt.Sprintf("f_%d", time.Now().UnixNano())
}

func generateBlockerID() string {
	return fmt.Sprintf("b_%d", time.Now().UnixNano())
}

func generateCommunicationID() string {
	return fmt.Sprintf("c_%d", time.Now().UnixNano())
}

func generateRecommendationID() string {
	return fmt.Sprintf("r_%d", time.Now().UnixNano())
}

func generateDecisionID() string {
	return fmt.Sprintf("d_%d", time.Now().UnixNano())
}
