#!/usr/bin/env python3
"""
Claude Code Plugin for Argus Integration

This script provides Claude Code with enhanced project awareness through Argus.
It can be called by Claude Code to get project insights and suggestions.
"""

import sys
import os
import json
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from claude_argus_helpers import *
from typing import Dict, List, Any

def analyze_current_context() -> Dict[str, Any]:
    """Analyze the current project context for Claude Code"""
    if not is_argus_available():
        return {
            "status": "unavailable",
            "message": "Argus is not running. Start Argus for enhanced project insights.",
            "suggestions": ["Start Argus with 'go run .' in the project directory"]
        }
    
    summary = get_project_summary()
    errors = get_current_errors()
    actions = get_next_actions()
    quality = analyze_code_quality()
    
    return {
        "status": "available",
        "summary": summary,
        "errors": errors[:5],  # Top 5 errors
        "next_actions": actions,
        "code_quality": quality,
        "context_string": get_claude_context(),
        "recommendations": generate_recommendations(summary, errors, quality)
    }

def generate_recommendations(summary: Dict, errors: List, quality: Dict) -> List[str]:
    """Generate specific recommendations for Claude Code"""
    recommendations = []
    
    # Health-based recommendations
    health_score = summary.get("health_score", 0)
    if health_score < 50:
        recommendations.append("🚨 Project health is critical - focus on fixing errors first")
    elif health_score < 80:
        recommendations.append("⚠️ Project health needs attention - consider refactoring")
    else:
        recommendations.append("✅ Project health is good - safe to add new features")
    
    # Error-based recommendations
    error_count = len(errors)
    if error_count > 10:
        recommendations.append("🐛 High error count - prioritize bug fixes over new features")
    elif error_count > 0:
        recommendations.append(f"🔧 Fix {error_count} active error(s) before proceeding")
    
    # Git-based recommendations
    if summary.get("is_dirty", False):
        recommendations.append("💾 Consider committing current changes")
    
    # Language-specific recommendations
    project_type = summary.get("project_type", "")
    if project_type == "go":
        recommendations.append("🔍 Use 'go fmt', 'go vet', and 'go mod tidy' for Go best practices")
    elif project_type == "javascript" or project_type == "typescript":
        recommendations.append("🔍 Consider using ESLint, Prettier, and TypeScript checks")
    elif project_type == "python":
        recommendations.append("🔍 Consider using Black, Flake8, and mypy for Python")
    
    return recommendations

def get_file_context(file_path: str) -> Dict[str, Any]:
    """Get context for a specific file"""
    if not is_argus_available():
        return {"error": "Argus not available"}
    
    return get_file_analysis(file_path)

def suggest_code_improvements(code_snippet: str, file_path: str = "") -> List[str]:
    """Suggest improvements for a code snippet using Argus data"""
    suggestions = []
    
    if not is_argus_available():
        return ["Consider starting Argus for enhanced code analysis"]
    
    # Get project context
    summary = get_project_summary()
    errors = get_current_errors()
    
    # File-specific analysis if path provided
    if file_path:
        file_context = get_file_context(file_path)
        if file_context.get("related_errors"):
            suggestions.extend(file_context["suggestions"])
    
    # General code improvement suggestions based on project state
    health_score = summary.get("health_score", 0)
    if health_score < 70:
        suggestions.append("Focus on error handling and defensive programming")
    
    # Check for common patterns
    lines = code_snippet.split('\n')
    if any('TODO' in line or 'FIXME' in line for line in lines):
        suggestions.append("Address TODO/FIXME comments in this code")
    
    # Language-specific suggestions
    if file_path.endswith('.go'):
        if 'fmt.Print' in code_snippet:
            suggestions.append("Consider using structured logging instead of fmt.Print")
        if 'panic(' in code_snippet:
            suggestions.append("Consider returning errors instead of using panic")
    
    return suggestions if suggestions else ["Code looks good! Consider adding tests if not present."]

def create_investigation_context(task_description: str) -> Dict[str, Any]:
    """Create an investigation context for a new task"""
    if not is_argus_available():
        return {"error": "Argus not available"}
    
    # Create a snapshot for this investigation
    snapshot_name = f"Investigation: {task_description[:50]}..."
    snapshot = create_snapshot(snapshot_name, f"Starting investigation for: {task_description}")
    
    # Add initial question
    question = f"How should I approach: {task_description}?"
    add_question(question, "high")
    
    # Get current project state
    context = analyze_current_context()
    
    return {
        "snapshot": snapshot,
        "task": task_description,
        "project_context": context,
        "initial_recommendations": generate_task_recommendations(task_description, context)
    }

def generate_task_recommendations(task: str, context: Dict) -> List[str]:
    """Generate recommendations for a specific task"""
    recommendations = []
    summary = context.get("summary", {})
    
    # Task-specific analysis
    task_lower = task.lower()
    
    if "bug" in task_lower or "fix" in task_lower or "error" in task_lower:
        recommendations.append("🐛 This is a bug fix - check current errors first")
        if summary.get("active_errors", 0) > 0:
            recommendations.append("📋 Review active errors in the project")
    
    if "feature" in task_lower or "add" in task_lower or "implement" in task_lower:
        health_score = summary.get("health_score", 0)
        if health_score < 70:
            recommendations.append("⚠️ Consider fixing existing issues before adding features")
        else:
            recommendations.append("✅ Project health is good for new features")
    
    if "refactor" in task_lower or "improve" in task_lower:
        recommendations.append("🔧 This is a refactoring task - ensure tests exist first")
        recommendations.append("📸 Create snapshots before and after refactoring")
    
    if "test" in task_lower:
        recommendations.append("🧪 Focus on edge cases and error conditions")
        recommendations.append("📊 Check current test coverage if available")
    
    return recommendations

def main():
    """Command line interface"""
    if len(sys.argv) < 2:
        print("Claude Code Plugin for Argus")
        print("Commands:")
        print("  analyze           - Analyze current project context")
        print("  file <path>       - Get context for specific file")
        print("  suggest <code>    - Suggest improvements for code")
        print("  investigate <task> - Start investigation for task")
        return
    
    command = sys.argv[1]
    
    if command == "analyze":
        result = analyze_current_context()
        print(json.dumps(result, indent=2))
    
    elif command == "file" and len(sys.argv) > 2:
        file_path = sys.argv[2]
        result = get_file_context(file_path)
        print(json.dumps(result, indent=2))
    
    elif command == "suggest" and len(sys.argv) > 2:
        code = " ".join(sys.argv[2:])
        suggestions = suggest_code_improvements(code)
        for suggestion in suggestions:
            print(suggestion)
    
    elif command == "investigate" and len(sys.argv) > 2:
        task = " ".join(sys.argv[2:])
        result = create_investigation_context(task)
        print(json.dumps(result, indent=2))
    
    else:
        print(f"Unknown command: {command}")

if __name__ == "__main__":
    main()