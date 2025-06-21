# Argus Technical Implementation Specification

## 🎯 Current State Analysis

### Existing Argus Capabilities (What We Have)
- **Solid Go Fiber backend** with comprehensive project monitoring
- **Rich data collection**: File structure, Git status, errors, TODOs, dependencies, processes
- **Multiple interfaces**: REST API, CLI tool (`claude-query.sh`), static dashboard, WebSocket test interface
- **Real-time capabilities**: WebSocket support for live error/process streaming (partially implemented)
- **Multi-language support**: Error detection for Go, JS/TS, Python, Java, etc.
- **Project intelligence**: Health scoring, dependency tracking, change monitoring

### Current Limitations (What's Missing)
- **No AI-specific features**: No intent detection, learning, or self-correction capabilities
- **Limited WebSocket implementation**: Process monitoring exists but process management incomplete
- **Static monitoring**: Reactive rather than proactive - watches but doesn't act
- **No feedback loop**: No learning from AI actions and outcomes
- **No checkpoint system**: No rollback/recovery mechanisms
- **Missing real-time process control**: Can monitor but limited process management
- **No multi-agent coordination**: No conflict detection between multiple AI agents
- **Limited error resolution**: Detection but no suggested fixes

## 🏗️ Core Technical Components to Build

### 1. Mind-Map Visualization Engine
```javascript
// Frontend: Interactive service topology
const ServiceTopology = {
  nodes: [
    { id: 'vox', name: 'VOX', type: 'service', status: 'healthy', position: {x: 100, y: 100} },
    { id: 'task', name: 'TASK', type: 'service', status: 'alpha', position: {x: 300, y: 100} },
    { id: 'crm', name: 'CRM', type: 'service', status: 'development', position: {x: 200, y: 200} }
  ],
  edges: [
    { from: 'vox', to: 'task', type: 'planned_integration', status: 'future' },
    { from: 'crm', to: 'fund', type: 'due_diligence_flow', status: 'priority' }
  ],
  layers: {
    human: { annotations: [], selections: [], snapshots: [] },
    ai: { traces: [], investigations: [], snapshots: [] }
  }
};
```

### 2. Snapshot System Architecture
```go
// Backend: Snapshot management
type VisualSnapshot struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`        // "human_exploration", "ai_investigation"
    Creator     string                 `json:"creator"`     // "human", "claude-instance-1"
    Timestamp   time.Time             `json:"timestamp"`
    Title       string                 `json:"title"`       // "Payment flow analysis"
    Description string                 `json:"description"` // User/AI provided context
    State       SnapshotState         `json:"state"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type SnapshotState struct {
    SelectedNodes    []string            `json:"selected_nodes"`
    HighlightedPaths [][]string         `json:"highlighted_paths"`
    Annotations      []Annotation       `json:"annotations"`
    ViewerState      ViewerState        `json:"viewer_state"`
    SystemState      SystemState        `json:"system_state"`  // What services were running, etc.
}

type Annotation struct {
    NodeID   string `json:"node_id"`
    Text     string `json:"text"`
    Type     string `json:"type"`    // "concern", "insight", "question"
    Position Point  `json:"position"`
}
```

### 3. AI-Argus Communication Protocol
```go
// WebSocket message types for Claude Code integration
type AIMessage struct {
    Type      string      `json:"type"`
    Timestamp time.Time   `json:"timestamp"`
    Data      interface{} `json:"data"`
}

// Message types:
// - "ai_intent": Claude announces what it's about to do
// - "ai_investigation": Claude is exploring specific services/relationships
// - "ai_snapshot": Claude saves investigation state
// - "ai_query": Claude asks Argus for specific information
// - "human_flag": Human interrupts AI process with concern
// - "human_snapshot": Human shares exploration findings

type AIIntent struct {
    Action      string   `json:"action"`       // "debug_auth", "integrate_services", "optimize_database"
    Targets     []string `json:"targets"`      // Service IDs being worked on
    Context     string   `json:"context"`      // Free-form description
    Risk        string   `json:"risk"`         // "low", "medium", "high"
    Estimated   string   `json:"estimated"`    // "2 minutes", "10 minutes"
}
```

### 4. Service Discovery and Relationship Mapping
```go
// Auto-detect HIVE components regardless of tech stack
type ServiceDiscovery struct {
    DetectedServices []DetectedService `json:"detected_services"`
    Relationships    []ServiceRelation `json:"relationships"`
    HealthStatus     []ServiceHealth   `json:"health_status"`
}

type DetectedService struct {
    ID           string            `json:"id"`
    Name         string            `json:"name"`
    Type         string            `json:"type"`         // "api", "frontend", "database", "worker"
    Technology   string            `json:"technology"`   // "go", "react", "python", "supabase"
    Port         int               `json:"port"`
    Endpoints    []APIEndpoint     `json:"endpoints"`
    Dependencies []string          `json:"dependencies"`
    ConfigFiles  []string          `json:"config_files"`
    Status       string            `json:"status"`       // "running", "stopped", "error", "unknown"
}

type ServiceRelation struct {
    From         string `json:"from"`
    To           string `json:"to"`
    Type         string `json:"type"`         // "api_call", "database_query", "file_share", "planned"
    Protocol     string `json:"protocol"`     // "http", "grpc", "direct_db", "file_system"
    Frequency    string `json:"frequency"`    // "high", "medium", "low", "unknown"
    Health       string `json:"health"`       // "healthy", "degraded", "broken", "untested"
    LastChecked  time.Time `json:"last_checked"`
}
```

### 5. Database Schema Intelligence
```go
// Understand shared Supabase database relationships
type DatabaseIntelligence struct {
    Tables       []TableInfo       `json:"tables"`
    Relationships []TableRelation  `json:"relationships"`
    Ownership    []ServiceOwnership `json:"ownership"`
    Migrations   []Migration       `json:"migrations"`
}

type TableInfo struct {
    Name         string      `json:"name"`
    Schema       string      `json:"schema"`
    Columns      []Column    `json:"columns"`
    Indexes      []Index     `json:"indexes"`
    UsedBy       []string    `json:"used_by"`        // Which services access this table
    RowCount     int64       `json:"row_count"`
    LastModified time.Time   `json:"last_modified"`
}

type ServiceOwnership struct {
    ServiceID    string   `json:"service_id"`
    OwnedTables  []string `json:"owned_tables"`     // Tables this service creates/manages
    AccessTables []string `json:"access_tables"`    // Tables this service reads from
    WriteLevel   string   `json:"write_level"`      // "owner", "contributor", "read_only"
}
```

## 🔌 Integration Points

### Claude Code Integration
```bash
# New claude-query.sh commands for AI collaboration
./claude-query.sh ai-intent "debugging auth timeout" --risk medium --targets auth,database
./claude-query.sh snapshot save "payment-flow-investigation" --type ai
./claude-query.sh snapshot load "user-exploration-1" --merge-with "ai-analysis-2"
./claude-query.sh service-map --highlight-path auth,payment,billing
./claude-query.sh flag-concern "Redis connection looks unstable" --ai-process current
```

### WebSocket Endpoints (New)
```go
// Add to existing WebSocket infrastructure
app.Get("/ws/ai", websocket.New(aiCollaborationHandler))
app.Get("/ws/mindmap", websocket.New(mindMapSyncHandler))
app.Get("/ws/snapshots", websocket.New(snapshotSyncHandler))
```

### REST API Extensions
```go
// New endpoints for mind-map and collaboration
app.Get("/api/topology", getServiceTopology)
app.Post("/api/snapshots", createSnapshot)
app.Get("/api/snapshots/:id", getSnapshot)
app.Post("/api/snapshots/compare", compareSnapshots)
app.Get("/api/ai/intent", getCurrentAIIntent)
app.Post("/api/ai/intent", setAIIntent)
app.Post("/api/human/flag", flagAIConcern)
```

## 💾 Data Storage Strategy

### Extend Existing ProjectSnapshot
```go
type EnhancedProjectSnapshot struct {
    // Existing fields
    Timestamp       time.Time              `json:"timestamp"`
    Structure       *ProjectStructure      `json:"structure"`
    RecentChanges   []FileChange           `json:"recent_changes"`
    GitStatus       *GitStatus             `json:"git_status"`
    
    // New HIVE-specific fields
    ServiceTopology *ServiceTopology       `json:"service_topology"`
    DatabaseSchema  *DatabaseIntelligence  `json:"database_schema"`
    Integrations    []ServiceIntegration   `json:"integrations"`
    AIActivity      *AIActivityState       `json:"ai_activity"`
    HumanSnapshots  []VisualSnapshot       `json:"human_snapshots"`
    AISnapshots     []VisualSnapshot       `json:"ai_snapshots"`
}
```

## 🚀 Implementation Phases

### Phase 1: Foundation (2-3 weeks)
1. **Service Discovery**: Detect running HIVE components
2. **Basic Mind-Map**: Static visualization of detected services
3. **Snapshot API**: Save/load visual exploration states
4. **Database Schema Detection**: Map Supabase table relationships

### Phase 2: Collaboration (3-4 weeks)
1. **Dual-Layer Interface**: Human + AI views
2. **WebSocket Sync**: Real-time mind-map updates
3. **AI Intent Integration**: Claude Code announces actions
4. **Snapshot Comparison**: Side-by-side analysis

### Phase 3: Intelligence (4-5 weeks)
1. **Change Impact Prediction**: Cross-service effect analysis
2. **Integration Pathway Suggestions**: How to connect services
3. **Performance Impact Tracking**: Before/after metrics
4. **Learning System**: Pattern recognition from outcomes

## 🧪 Testing Strategy

### Development Testing
- **Mock HIVE services** for testing service discovery
- **Simulated AI interactions** for collaboration workflow
- **Database fixtures** for schema intelligence testing
- **Load testing** for real-time mind-map updates

### Integration Testing
- **Real HIVE component detection** as they become available
- **Claude Code integration** with actual development workflows
- **Multi-developer scenarios** with concurrent usage
- **Performance testing** with complex service topologies

## 📊 Monitoring and Metrics

### Development Speed Metrics
- Time from "start development" to "all services connected"
- Cross-service debugging time reduction
- Integration bug discovery speed

### System Understanding Metrics
- Developer onboarding time for multi-service features
- Accuracy of change impact predictions
- Snapshot usage patterns and effectiveness

### AI Collaboration Metrics
- Human intervention frequency in AI processes
- Snapshot comparison usage
- Successful problem resolution from collaboration

---

*This technical specification provides the implementation roadmap for transforming Argus into an AI development companion specifically designed for HIVE's multi-service architecture evolution.*