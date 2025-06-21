#!/usr/bin/env python3
"""
Enhanced Code Suggestions using Argus Data

This module provides intelligent code suggestions based on real-time project analysis from Argus.
It combines static code analysis with dynamic project health metrics to provide contextual recommendations.
"""

import sys
import os
import re
import json
from pathlib import Path
from typing import Dict, List, Any, Optional, Tuple
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from claude_argus_helpers import *

class ArgusCodeSuggestions:
    """Enhanced code suggestions powered by Argus project intelligence"""
    
    def __init__(self):
        self.project_context = None
        self.refresh_context()
    
    def refresh_context(self):
        """Refresh project context from Argus"""
        if is_argus_available():
            self.project_context = get_integration().get_project_context()
        else:
            self.project_context = None
    
    def analyze_code_block(self, code: str, file_path: str = "", language: str = "") -> Dict[str, Any]:
        """Analyze a code block and provide suggestions"""
        if not language and file_path:
            language = self._detect_language_from_file(file_path)
        
        suggestions = {
            "quality_issues": self._find_quality_issues(code, language),
            "security_concerns": self._find_security_issues(code, language),
            "performance_hints": self._find_performance_issues(code, language),
            "style_improvements": self._find_style_issues(code, language),
            "project_specific": self._get_project_specific_suggestions(code, file_path),
            "argus_insights": self._get_argus_insights(code, file_path),
            "overall_score": 0
        }
        
        # Calculate overall score
        total_issues = sum(len(suggestions[key]) for key in ["quality_issues", "security_concerns", "performance_hints"])
        max_score = 100
        score_deduction = min(total_issues * 10, 80)  # Max 80 points deduction
        suggestions["overall_score"] = max_score - score_deduction
        
        return suggestions
    
    def suggest_improvements_for_file(self, file_path: str) -> Dict[str, Any]:
        """Suggest improvements for an entire file"""
        if not os.path.exists(file_path):
            return {"error": f"File {file_path} not found"}
        
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except Exception as e:
            return {"error": f"Could not read file: {e}"}
        
        language = self._detect_language_from_file(file_path)
        analysis = self.analyze_code_block(content, file_path, language)
        
        # Add file-specific suggestions
        analysis["file_suggestions"] = self._get_file_level_suggestions(file_path, content, language)
        analysis["argus_file_context"] = get_file_analysis(file_path) if is_argus_available() else {}
        
        return analysis
    
    def suggest_next_coding_steps(self, current_task: str = "") -> List[str]:
        """Suggest next coding steps based on project state"""
        suggestions = []
        
        if not is_argus_available():
            suggestions.append("🚀 Start Argus monitoring for enhanced project insights")
            return suggestions
        
        summary = get_project_summary()
        errors = get_current_errors()
        actions = get_next_actions()
        
        # Prioritize based on project health
        health_score = summary.get("health_score", 0)
        
        if health_score < 30:
            suggestions.append("🆘 Critical: Project health is very low - focus on error resolution")
            suggestions.extend([f"🔧 {action}" for action in actions[:2]])
        elif health_score < 70:
            suggestions.append("⚠️ Moderate issues detected - balance fixes with development")
            if errors:
                suggestions.append("🐛 Address critical errors first")
            suggestions.append("🔍 Consider refactoring problematic areas")
        else:
            suggestions.append("✅ Project health is good - safe to proceed with development")
            if current_task:
                suggestions.extend(self._get_task_specific_suggestions(current_task, summary))
        
        # Add git workflow suggestions
        if summary.get("is_dirty", False):
            suggestions.append("📝 Consider committing current progress")
        
        return suggestions
    
    def _detect_language_from_file(self, file_path: str) -> str:
        """Detect programming language from file extension"""
        ext = Path(file_path).suffix.lower()
        language_map = {
            '.py': 'python',
            '.js': 'javascript',
            '.ts': 'typescript',
            '.go': 'go',
            '.java': 'java',
            '.c': 'c',
            '.cpp': 'cpp',
            '.cs': 'csharp',
            '.php': 'php',
            '.rb': 'ruby',
            '.rs': 'rust',
            '.html': 'html',
            '.css': 'css',
            '.sql': 'sql',
            '.sh': 'bash'
        }
        return language_map.get(ext, 'unknown')
    
    def _find_quality_issues(self, code: str, language: str) -> List[str]:
        """Find code quality issues"""
        issues = []
        lines = code.split('\n')
        
        # Common quality issues across languages
        for i, line in enumerate(lines, 1):
            # Long lines
            if len(line) > 120:
                issues.append(f"Line {i}: Line too long ({len(line)} chars) - consider breaking it up")
            
            # TODO/FIXME comments
            if 'TODO' in line or 'FIXME' in line:
                issues.append(f"Line {i}: Address TODO/FIXME comment")
            
            # Debugging statements
            debug_patterns = ['console.log', 'print(', 'println(', 'fmt.Print', 'System.out.print']
            for pattern in debug_patterns:
                if pattern in line and not line.strip().startswith('//') and not line.strip().startswith('#'):
                    issues.append(f"Line {i}: Remove debugging statement ({pattern})")
        
        # Language-specific quality checks
        if language == 'python':
            issues.extend(self._find_python_quality_issues(code))
        elif language == 'go':
            issues.extend(self._find_go_quality_issues(code))
        elif language in ['javascript', 'typescript']:
            issues.extend(self._find_js_quality_issues(code))
        
        return issues
    
    def _find_security_issues(self, code: str, language: str) -> List[str]:
        """Find potential security issues"""
        issues = []
        
        # Common security patterns
        security_patterns = [
            ('password', 'Avoid hardcoded passwords'),
            ('api_key', 'Avoid hardcoded API keys'),
            ('secret', 'Avoid hardcoded secrets'),
            ('token', 'Be careful with token handling'),
            ('eval(', 'Avoid eval() - security risk'),
            ('exec(', 'Avoid exec() - security risk'),
            ('shell=True', 'shell=True can be dangerous'),
            ('innerHTML', 'innerHTML can lead to XSS'),
            ('document.write', 'document.write can lead to XSS')
        ]
        
        code_lower = code.lower()
        for pattern, message in security_patterns:
            if pattern in code_lower:
                issues.append(f"Security: {message}")
        
        return issues
    
    def _find_performance_issues(self, code: str, language: str) -> List[str]:
        """Find potential performance issues"""
        issues = []
        lines = code.split('\n')
        
        # Common performance anti-patterns
        for i, line in enumerate(lines, 1):
            # Nested loops (simplified detection)
            if 'for ' in line and any('for ' in lines[j] for j in range(max(0, i), min(len(lines), i + 5))):
                issues.append(f"Line {i}: Nested loops detected - consider optimization")
        
        # Language-specific performance checks
        if language == 'python':
            if '+=' in code and 'for ' in code:
                issues.append("Performance: Consider using list comprehension instead of += in loops")
        
        elif language == 'go':
            if 'append(' in code and 'for ' in code:
                issues.append("Performance: Pre-allocate slices when size is known")
        
        return issues
    
    def _find_style_issues(self, code: str, language: str) -> List[str]:
        """Find code style issues"""
        issues = []
        
        # Language-specific style checks
        if language == 'python':
            issues.extend(self._find_python_style_issues(code))
        elif language == 'go':
            issues.extend(self._find_go_style_issues(code))
        elif language in ['javascript', 'typescript']:
            issues.extend(self._find_js_style_issues(code))
        
        return issues
    
    def _find_python_quality_issues(self, code: str) -> List[str]:
        """Find Python-specific quality issues"""
        issues = []
        
        if 'except:' in code:
            issues.append("Python: Avoid bare except clauses - specify exception types")
        
        if 'import *' in code:
            issues.append("Python: Avoid wildcard imports - import specific items")
        
        return issues
    
    def _find_python_style_issues(self, code: str) -> List[str]:
        """Find Python-specific style issues"""
        issues = []
        lines = code.split('\n')
        
        for i, line in enumerate(lines, 1):
            # PEP 8 checks
            if line.endswith(' '):
                issues.append(f"Line {i}: Remove trailing whitespace")
            
            if '==' in line and ('True' in line or 'False' in line):
                issues.append(f"Line {i}: Use 'is True' or 'is False' instead of '== True/False'")
        
        return issues
    
    def _find_go_quality_issues(self, code: str) -> List[str]:
        """Find Go-specific quality issues"""
        issues = []
        
        if 'panic(' in code:
            issues.append("Go: Consider returning errors instead of using panic")
        
        if 'fmt.Print' in code and 'main(' not in code:
            issues.append("Go: Consider using structured logging instead of fmt.Print")
        
        return issues
    
    def _find_go_style_issues(self, code: str) -> List[str]:
        """Find Go-specific style issues"""
        issues = []
        lines = code.split('\n')
        
        for i, line in enumerate(lines, 1):
            # Go conventions
            if 'func ' in line and '(' in line and ')' in line:
                if not re.match(r'.*func\s+[A-Z][a-zA-Z]*\s*\(', line) and 'main(' not in line:
                    if not line.strip().startswith('//'):
                        issues.append(f"Line {i}: Exported functions should start with capital letter")
        
        return issues
    
    def _find_js_quality_issues(self, code: str) -> List[str]:
        """Find JavaScript/TypeScript quality issues"""
        issues = []
        
        if '==' in code and '===' not in code:
            issues.append("JS: Use strict equality (===) instead of loose equality (==)")
        
        if 'var ' in code:
            issues.append("JS: Consider using 'let' or 'const' instead of 'var'")
        
        return issues
    
    def _find_js_style_issues(self, code: str) -> List[str]:
        """Find JavaScript/TypeScript style issues"""
        issues = []
        
        if code.count(';') < code.count('\n') / 2:
            issues.append("JS: Consider using semicolons consistently")
        
        return issues
    
    def _get_project_specific_suggestions(self, code: str, file_path: str) -> List[str]:
        """Get project-specific suggestions based on Argus data"""
        suggestions = []
        
        if not self.project_context:
            return suggestions
        
        # Check if code matches project patterns
        project_type = self.project_context.get("structure", {}).get("project_type", "")
        
        if project_type == "go" and file_path.endswith('.go'):
            if 'http' in code.lower() and 'fiber' not in code.lower():
                suggestions.append("Project: Consider using Fiber framework (already in project)")
        
        return suggestions
    
    def _get_argus_insights(self, code: str, file_path: str) -> List[str]:
        """Get insights from Argus monitoring data"""
        insights = []
        
        if not is_argus_available():
            return insights
        
        # Get current project state
        errors = get_current_errors()
        summary = get_project_summary()
        
        # Check if code relates to current errors
        for error in errors[:3]:
            error_str = str(error).lower()
            if any(word in code.lower() for word in error_str.split() if len(word) > 3):
                insights.append(f"Argus: This code may relate to active error - verify carefully")
        
        # Health-based insights
        health_score = summary.get("health_score", 0)
        if health_score < 50:
            insights.append("Argus: Project health is low - ensure this code doesn't add complexity")
        
        return insights
    
    def _get_file_level_suggestions(self, file_path: str, content: str, language: str) -> List[str]:
        """Get file-level suggestions"""
        suggestions = []
        
        # File size check
        line_count = len(content.split('\n'))
        if line_count > 500:
            suggestions.append(f"File: Large file ({line_count} lines) - consider splitting")
        
        # Missing documentation
        if language in ['python', 'go', 'java'] and 'func ' in content:
            func_count = content.count('func ')
            doc_count = content.count('"""') + content.count('/*')
            if doc_count == 0 and func_count > 3:
                suggestions.append("File: Consider adding documentation for functions")
        
        return suggestions
    
    def _get_task_specific_suggestions(self, task: str, summary: Dict) -> List[str]:
        """Get suggestions specific to the current task"""
        suggestions = []
        task_lower = task.lower()
        
        if "test" in task_lower:
            suggestions.append("🧪 Focus on edge cases and error conditions")
            suggestions.append("📊 Ensure good test coverage")
        
        elif "api" in task_lower or "endpoint" in task_lower:
            suggestions.append("🔒 Implement proper authentication and validation")
            suggestions.append("📝 Document API endpoints")
        
        elif "database" in task_lower or "db" in task_lower:
            suggestions.append("🛡️ Use parameterized queries to prevent SQL injection")
            suggestions.append("🚀 Consider database indexes for performance")
        
        elif "ui" in task_lower or "frontend" in task_lower:
            suggestions.append("♿ Ensure accessibility compliance")
            suggestions.append("📱 Test responsive design")
        
        return suggestions

def analyze_file(file_path: str) -> None:
    """Analyze a file and print suggestions"""
    analyzer = ArgusCodeSuggestions()
    result = analyzer.suggest_improvements_for_file(file_path)
    
    if "error" in result:
        print(f"Error: {result['error']}")
        return
    
    print(f"📄 Analysis for: {file_path}")
    print(f"🎯 Overall Score: {result['overall_score']}/100")
    print()
    
    for category, issues in result.items():
        if isinstance(issues, list) and issues and category != "argus_file_context":
            print(f"🔍 {category.replace('_', ' ').title()}:")
            for issue in issues:
                print(f"   • {issue}")
            print()

def analyze_code(code: str, language: str = "") -> None:
    """Analyze a code snippet and print suggestions"""
    analyzer = ArgusCodeSuggestions()
    result = analyzer.analyze_code_block(code, language=language)
    
    print(f"🎯 Code Quality Score: {result['overall_score']}/100")
    print()
    
    for category, issues in result.items():
        if isinstance(issues, list) and issues:
            print(f"🔍 {category.replace('_', ' ').title()}:")
            for issue in issues:
                print(f"   • {issue}")
            print()

def main():
    """Command line interface"""
    if len(sys.argv) < 2:
        print("Argus Code Suggestions")
        print("Usage:")
        print("  python3 argus_code_suggestions.py file <path>     - Analyze file")
        print("  python3 argus_code_suggestions.py code <code>     - Analyze code snippet")
        print("  python3 argus_code_suggestions.py steps [task]    - Get next coding steps")
        return
    
    command = sys.argv[1]
    
    if command == "file" and len(sys.argv) > 2:
        file_path = sys.argv[2]
        analyze_file(file_path)
    
    elif command == "code" and len(sys.argv) > 2:
        code = " ".join(sys.argv[2:])
        analyze_code(code)
    
    elif command == "steps":
        task = " ".join(sys.argv[2:]) if len(sys.argv) > 2 else ""
        analyzer = ArgusCodeSuggestions()
        steps = analyzer.suggest_next_coding_steps(task)
        
        print("🚀 Suggested Next Steps:")
        for i, step in enumerate(steps, 1):
            print(f"   {i}. {step}")
    
    else:
        print(f"Unknown command: {command}")

if __name__ == "__main__":
    main()