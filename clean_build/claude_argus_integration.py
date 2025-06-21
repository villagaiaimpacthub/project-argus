#!/usr/bin/env python3
"""
Claude Code <-> Argus Integration Bridge

This script provides functions for Claude Code to query Argus for project insights.
It acts as a bridge between Claude Code's analysis capabilities and Argus's
real-time project monitoring data.
"""

import json
import requests
import sys
import os
from typing import Dict, List, Optional, Any
from datetime import datetime

class ArgusClient:
    """Client for communicating with Argus API"""
    
    def __init__(self, base_url: str = "http://localhost:3002"):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'User-Agent': 'Claude-Code-Integration/1.0'
        })
    
    def is_available(self) -> bool:
        """Check if Argus is running and accessible"""
        try:
            response = self.session.get(f"{self.base_url}/health", timeout=5)
            return response.status_code == 200
        except:
            return False
    
    def get_project_health(self) -> Dict[str, Any]:
        """Get overall project health metrics"""
        try:
            response = self.session.get(f"{self.base_url}/health")
            response.raise_for_status()
            data = response.json()
            
            return {
                "health_score": data.get("score", 0),
                "error_count": data.get("error_count", 0),
                "warning_count": data.get("warning_count", 0),
                "technical_debt": data.get("technical_debt", "unknown"),
                "last_check": data.get("last_health_check"),
                "total_files": 0,  # Not in health endpoint
                "total_size": 0    # Not in health endpoint
            }
        except Exception as e:
            return {"error": str(e)}
    
    def get_detected_languages(self) -> List[Dict[str, Any]]:
        """Get all detected programming languages in the project"""
        try:
            response = self.session.get(f"{self.base_url}/api/languages")
            response.raise_for_status()
            data = response.json()
            
            languages = []
            for lang in data.get("languages", []):
                languages.append({
                    "name": lang["language"]["name"],
                    "file_count": lang["file_count"],
                    "line_count": lang["line_count"],
                    "frameworks": [f["name"] for f in lang.get("frameworks", [])],
                    "has_tests": lang.get("has_tests", False),
                    "has_linting": lang.get("has_linting", False),
                    "package_manager": lang.get("package_manager", "unknown")
                })
            
            return languages
        except Exception as e:
            return [{"error": str(e)}]
    
    def get_active_errors(self) -> List[Dict[str, Any]]:
        """Get current errors and warnings in the project"""
        try:
            response = self.session.get(f"{self.base_url}/errors")
            response.raise_for_status()
            data = response.json()
            # Handle both array and object responses
            if isinstance(data, list):
                return data
            return data.get("errors", [])
        except Exception as e:
            return [{"error": str(e)}]
    
    def get_project_structure(self) -> Dict[str, Any]:
        """Get project file structure and organization"""
        try:
            response = self.session.get(f"{self.base_url}/structure")
            response.raise_for_status()
            data = response.json()
            
            return {
                "root_path": data.get("root_path", ""),
                "total_files": len(data.get("files", [])),
                "file_types": self._analyze_file_types(data.get("files", [])),
                "directories": len(data.get("directories", [])),
                "project_type": data.get("project_type", "unknown"),
                "main_files": data.get("main_files", []),
                "config_files": data.get("config_files", [])
            }
        except Exception as e:
            return {"error": str(e)}
    
    def get_recent_changes(self) -> List[Dict[str, Any]]:
        """Get recent file changes and modifications"""
        try:
            response = self.session.get(f"{self.base_url}/changes")
            response.raise_for_status()
            data = response.json()
            # Handle both array and object responses
            if isinstance(data, list):
                return data
            return data.get("changes", [])
        except Exception as e:
            return [{"error": str(e)}]
    
    def get_git_status(self) -> Dict[str, Any]:
        """Get current git repository status"""
        try:
            response = self.session.get(f"{self.base_url}/git")
            response.raise_for_status()
            return response.json()
        except Exception as e:
            return {"error": str(e)}
    
    def get_dependencies(self) -> List[Dict[str, Any]]:
        """Get project dependencies and their status"""
        try:
            response = self.session.get(f"{self.base_url}/dependencies")
            response.raise_for_status()
            return response.json().get("dependencies", [])
        except Exception as e:
            return [{"error": str(e)}]
    
    def get_todos(self) -> List[Dict[str, Any]]:
        """Get TODO/FIXME items found in code"""
        try:
            response = self.session.get(f"{self.base_url}/todos")
            response.raise_for_status()
            return response.json().get("todos", [])
        except Exception as e:
            return [{"error": str(e)}]
    
    def get_running_processes(self) -> List[Dict[str, Any]]:
        """Get currently running development processes"""
        try:
            response = self.session.get(f"{self.base_url}/processes")
            response.raise_for_status()
            return response.json().get("processes", [])
        except Exception as e:
            return [{"error": str(e)}]
    
    def search_project(self, query: str) -> List[Dict[str, Any]]:
        """Search across project files for specific content"""
        try:
            response = self.session.get(f"{self.base_url}/search", params={"q": query})
            response.raise_for_status()
            return response.json().get("results", [])
        except Exception as e:
            return [{"error": str(e)}]
    
    def get_project_topology(self) -> Dict[str, Any]:
        """Get project topology and service relationships"""
        try:
            response = self.session.get(f"{self.base_url}/api/topology")
            response.raise_for_status()
            return response.json()
        except Exception as e:
            return {"error": str(e)}
    
    def create_investigation_snapshot(self, name: str, description: str = "") -> Dict[str, Any]:
        """Create a new investigation snapshot"""
        try:
            payload = {"name": name, "description": description}
            response = self.session.post(f"{self.base_url}/api/snapshots", json=payload)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            return {"error": str(e)}
    
    def get_investigation_state(self) -> Dict[str, Any]:
        """Get current investigation tracking state"""
        try:
            response = self.session.get(f"{self.base_url}/api/investigation/state")
            response.raise_for_status()
            return response.json()
        except Exception as e:
            return {"error": str(e)}
    
    def add_investigation_question(self, question: str, priority: str = "medium") -> Dict[str, Any]:
        """Add a question to the current investigation"""
        try:
            payload = {
                "question": question,
                "priority": priority,
                "source": "claude-code"
            }
            response = self.session.post(f"{self.base_url}/api/investigation/questions", json=payload)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            return {"error": str(e)}
    
    def add_investigation_finding(self, title: str, description: str, impact: str = "medium", category: str = "general") -> Dict[str, Any]:
        """Add a finding to the current investigation"""
        try:
            payload = {
                "title": title,
                "description": description,
                "impact": impact,
                "category": category
            }
            response = self.session.post(f"{self.base_url}/api/investigation/findings", json=payload)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            return {"error": str(e)}
    
    def _analyze_file_types(self, files: List[Dict]) -> Dict[str, int]:
        """Analyze file types in the project"""
        file_types = {}
        for file_info in files:
            ext = os.path.splitext(file_info.get("path", ""))[-1].lower()
            if ext:
                file_types[ext] = file_types.get(ext, 0) + 1
        return file_types

class ClaudeArgusIntegration:
    """Main integration class for Claude Code + Argus"""
    
    def __init__(self):
        self.argus = ArgusClient()
    
    def get_project_context(self) -> Dict[str, Any]:
        """Get comprehensive project context for Claude Code"""
        if not self.argus.is_available():
            return {"error": "Argus is not available. Please start Argus first."}
        
        context = {
            "timestamp": datetime.now().isoformat(),
            "health": self.argus.get_project_health(),
            "languages": self.argus.get_detected_languages(),
            "structure": self.argus.get_project_structure(),
            "git_status": self.argus.get_git_status(),
            "active_errors": self.argus.get_active_errors()[:10],  # Limit to 10 most recent
            "recent_changes": self.argus.get_recent_changes()[:5],  # Limit to 5 most recent
            "dependencies": len(self.argus.get_dependencies()),
            "todos": len(self.argus.get_todos()),
            "running_processes": len(self.argus.get_running_processes())
        }
        
        return context
    
    def analyze_code_quality(self) -> Dict[str, Any]:
        """Analyze code quality using Argus data"""
        health = self.argus.get_project_health()
        errors = self.argus.get_active_errors()
        todos = self.argus.get_todos()
        
        analysis = {
            "overall_score": health.get("health_score", 0),
            "critical_issues": len([e for e in errors if e.get("severity") == "error"]),
            "warnings": len([e for e in errors if e.get("severity") == "warning"]),
            "technical_debt": health.get("technical_debt", "unknown"),
            "todo_count": len(todos),
            "recommendations": []
        }
        
        # Generate recommendations
        if analysis["critical_issues"] > 0:
            analysis["recommendations"].append("Fix critical errors before proceeding with new features")
        
        if analysis["overall_score"] < 80:
            analysis["recommendations"].append("Consider refactoring to improve code quality")
        
        if analysis["todo_count"] > 20:
            analysis["recommendations"].append("Address TODO items to reduce technical debt")
        
        return analysis
    
    def suggest_next_actions(self) -> List[str]:
        """Suggest next actions based on project state"""
        context = self.get_project_context()
        suggestions = []
        
        # Based on errors
        if context.get("active_errors"):
            suggestions.append(f"🐛 Fix {len(context['active_errors'])} active errors")
        
        # Based on git status
        git_status = context.get("git_status", {})
        if git_status.get("is_dirty"):
            suggestions.append("💾 Commit pending changes")
        
        # Based on health
        health = context.get("health", {})
        if health.get("health_score", 0) < 90:
            suggestions.append("🔧 Improve code quality (current score: {}/100)".format(health.get("health_score", 0)))
        
        # Based on dependencies
        if context.get("dependencies", 0) > 0:
            suggestions.append("📦 Review dependencies for updates")
        
        if not suggestions:
            suggestions.append("✅ Project looks good! Consider adding tests or documentation")
        
        return suggestions

def main():
    """Main function for command-line usage"""
    integration = ClaudeArgusIntegration()
    
    if len(sys.argv) < 2:
        print("Usage: python claude-argus-integration.py <command>")
        print("Commands:")
        print("  context     - Get project context")
        print("  health      - Get project health")
        print("  languages   - Get detected languages")
        print("  errors      - Get active errors")
        print("  quality     - Analyze code quality")
        print("  suggestions - Get next action suggestions")
        print("  search <query> - Search project")
        return
    
    command = sys.argv[1]
    
    if command == "context":
        print(json.dumps(integration.get_project_context(), indent=2))
    elif command == "health":
        print(json.dumps(integration.argus.get_project_health(), indent=2))
    elif command == "languages":
        print(json.dumps(integration.argus.get_detected_languages(), indent=2))
    elif command == "errors":
        print(json.dumps(integration.argus.get_active_errors(), indent=2))
    elif command == "quality":
        print(json.dumps(integration.analyze_code_quality(), indent=2))
    elif command == "suggestions":
        suggestions = integration.suggest_next_actions()
        for suggestion in suggestions:
            print(suggestion)
    elif command == "search" and len(sys.argv) > 2:
        query = " ".join(sys.argv[2:])
        results = integration.argus.search_project(query)
        print(json.dumps(results, indent=2))
    else:
        print(f"Unknown command: {command}")

if __name__ == "__main__":
    main()