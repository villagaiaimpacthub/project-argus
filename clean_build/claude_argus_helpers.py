#!/usr/bin/env python3
"""
Claude Code Helper Functions for Argus Integration

This module provides simple helper functions that Claude Code can use
to get project insights from Argus without dealing with the API details.
"""

import sys
import os
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from claude_argus_integration import ClaudeArgusIntegration

from typing import Dict, List, Optional, Any

# Global integration instance
_integration = None

def get_integration():
    """Get or create the global integration instance"""
    global _integration
    if _integration is None:
        _integration = ClaudeArgusIntegration()
    return _integration

def is_argus_available() -> bool:
    """Check if Argus is running and available"""
    return get_integration().argus.is_available()

def get_project_summary() -> Dict[str, Any]:
    """Get a quick project summary for Claude Code"""
    if not is_argus_available():
        return {"error": "Argus is not available"}
    
    context = get_integration().get_project_context()
    if "error" in context:
        return context
    
    return {
        "health_score": context.get("health", {}).get("health_score", 0),
        "total_files": context.get("structure", {}).get("total_files", 0),
        "project_type": context.get("structure", {}).get("project_type", "unknown"),
        "languages": len(context.get("languages", [])),
        "active_errors": len(context.get("active_errors", [])),
        "git_branch": context.get("git_status", {}).get("branch", "unknown"),
        "is_dirty": context.get("git_status", {}).get("is_dirty", False)
    }

def get_current_errors() -> List[Dict[str, Any]]:
    """Get current errors in the project"""
    if not is_argus_available():
        return [{"error": "Argus is not available"}]
    
    return get_integration().argus.get_active_errors()

def get_project_languages() -> List[Dict[str, Any]]:
    """Get all detected languages in the project"""
    if not is_argus_available():
        return [{"error": "Argus is not available"}]
    
    return get_integration().argus.get_detected_languages()

def search_project(query: str) -> List[Dict[str, Any]]:
    """Search for content across the project"""
    if not is_argus_available():
        return [{"error": "Argus is not available"}]
    
    return get_integration().argus.search_project(query)

def get_git_status() -> Dict[str, Any]:
    """Get current git repository status"""
    if not is_argus_available():
        return {"error": "Argus is not available"}
    
    return get_integration().argus.get_git_status()

def get_next_actions() -> List[str]:
    """Get suggested next actions based on project state"""
    if not is_argus_available():
        return ["⚠️ Argus is not available - start Argus to get project insights"]
    
    return get_integration().suggest_next_actions()

def analyze_code_quality() -> Dict[str, Any]:
    """Analyze overall code quality"""
    if not is_argus_available():
        return {"error": "Argus is not available"}
    
    return get_integration().analyze_code_quality()

def get_project_structure() -> Dict[str, Any]:
    """Get project file structure and organization"""
    if not is_argus_available():
        return {"error": "Argus is not available"}
    
    return get_integration().argus.get_project_structure()

def get_running_processes() -> List[Dict[str, Any]]:
    """Get currently running development processes"""
    if not is_argus_available():
        return [{"error": "Argus is not available"}]
    
    return get_integration().argus.get_running_processes()

def get_todos() -> List[Dict[str, Any]]:
    """Get TODO/FIXME items found in code"""
    if not is_argus_available():
        return [{"error": "Argus is not available"}]
    
    return get_integration().argus.get_todos()

def create_snapshot(name: str, description: str = "") -> Dict[str, Any]:
    """Create an investigation snapshot"""
    if not is_argus_available():
        return {"error": "Argus is not available"}
    
    return get_integration().argus.create_investigation_snapshot(name, description)

def add_finding(title: str, description: str, impact: str = "medium", category: str = "general") -> Dict[str, Any]:
    """Add a finding to the current investigation"""
    if not is_argus_available():
        return {"error": "Argus is not available"}
    
    return get_integration().argus.add_investigation_finding(title, description, impact, category)

def add_question(question: str, priority: str = "medium") -> Dict[str, Any]:
    """Add a question to the current investigation"""
    if not is_argus_available():
        return {"error": "Argus is not available"}
    
    return get_integration().argus.add_investigation_question(question, priority)

def get_file_analysis(file_path: str) -> Dict[str, Any]:
    """Analyze a specific file using Argus data"""
    if not is_argus_available():
        return {"error": "Argus is not available"}
    
    # Search for the file
    search_results = search_project(os.path.basename(file_path))
    
    # Get errors that might be related to this file
    errors = get_current_errors()
    file_errors = [e for e in errors if file_path in str(e)]
    
    # Get project context for additional insights
    context = get_integration().get_project_context()
    
    return {
        "file_path": file_path,
        "found_in_search": len(search_results) > 0,
        "search_results": search_results[:3],  # Top 3 results
        "related_errors": file_errors,
        "project_health": context.get("health", {}).get("health_score", 0),
        "suggestions": _get_file_suggestions(file_path, file_errors, context)
    }

def _get_file_suggestions(file_path: str, errors: List[Dict], context: Dict) -> List[str]:
    """Generate suggestions for a specific file"""
    suggestions = []
    
    # Error-based suggestions
    if errors:
        suggestions.append(f"🐛 Fix {len(errors)} error(s) in this file")
    
    # File type suggestions
    ext = os.path.splitext(file_path)[-1].lower()
    if ext == '.go':
        suggestions.append("🔍 Run 'go fmt' and 'go vet' for Go best practices")
    elif ext in ['.js', '.ts']:
        suggestions.append("🔍 Consider running ESLint or Prettier")
    elif ext == '.py':
        suggestions.append("🔍 Consider running Black or Flake8")
    
    # Health-based suggestions
    health_score = context.get("health", {}).get("health_score", 0)
    if health_score < 80:
        suggestions.append("🔧 Consider refactoring to improve code quality")
    
    return suggestions

def get_claude_context() -> str:
    """Get a formatted context string for Claude Code"""
    if not is_argus_available():
        return "Argus monitoring is not available. Consider starting Argus for enhanced project insights."
    
    summary = get_project_summary()
    if "error" in summary:
        return f"Argus error: {summary['error']}"
    
    context_parts = [
        f"📊 Project Health: {summary['health_score']}/100",
        f"📁 Files: {summary['total_files']} ({summary['project_type']} project)",
        f"🌐 Languages: {summary['languages']}",
        f"🔴 Errors: {summary['active_errors']}",
        f"🌿 Git: {summary['git_branch']}" + (" (dirty)" if summary['is_dirty'] else " (clean)")
    ]
    
    return " | ".join(context_parts)

def main():
    """Command-line interface for testing"""
    if len(sys.argv) < 2:
        print("Available commands:")
        print("  summary     - Get project summary")
        print("  context     - Get Claude context string")
        print("  errors      - Get current errors")
        print("  languages   - Get detected languages")
        print("  structure   - Get project structure")
        print("  actions     - Get suggested next actions")
        print("  quality     - Analyze code quality")
        print("  git         - Get git status")
        print("  search <query> - Search project")
        print("  file <path> - Analyze specific file")
        return
    
    command = sys.argv[1]
    
    if command == "summary":
        import json
        print(json.dumps(get_project_summary(), indent=2))
    elif command == "context":
        print(get_claude_context())
    elif command == "errors":
        import json
        print(json.dumps(get_current_errors(), indent=2))
    elif command == "languages":
        import json
        print(json.dumps(get_project_languages(), indent=2))
    elif command == "structure":
        import json
        print(json.dumps(get_project_structure(), indent=2))
    elif command == "actions":
        actions = get_next_actions()
        for action in actions:
            print(action)
    elif command == "quality":
        import json
        print(json.dumps(analyze_code_quality(), indent=2))
    elif command == "git":
        import json
        print(json.dumps(get_git_status(), indent=2))
    elif command == "search" and len(sys.argv) > 2:
        import json
        query = " ".join(sys.argv[2:])
        print(json.dumps(search_project(query), indent=2))
    elif command == "file" and len(sys.argv) > 2:
        import json
        file_path = sys.argv[2]
        print(json.dumps(get_file_analysis(file_path), indent=2))
    else:
        print(f"Unknown command: {command}")

if __name__ == "__main__":
    main()