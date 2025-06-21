#!/usr/bin/env python3
"""
Quick Argus Context for Claude Code

A simple command that Claude Code can run to get instant project context.
Usage: python3 argus_context.py [--json|--brief|--detailed]
"""

import sys
import os
import json
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from claude_argus_helpers import *

def get_brief_context():
    """Get a brief one-line context"""
    return get_claude_context()

def get_detailed_context():
    """Get detailed context with actionable insights"""
    if not is_argus_available():
        return "Argus monitoring unavailable. Start with 'go run .' for enhanced insights."
    
    summary = get_project_summary()
    actions = get_next_actions()
    errors = get_current_errors()
    
    lines = []
    lines.append(f"🎯 Project: {summary['project_type'].upper()} | Health: {summary['health_score']}/100")
    lines.append(f"📊 Status: {summary['total_files']} files, {summary['active_errors']} errors, {summary['languages']} languages")
    lines.append(f"🌿 Git: {summary['git_branch']}" + (" (uncommitted changes)" if summary['is_dirty'] else " (clean)"))
    
    if errors:
        lines.append(f"🔴 Current Errors:")
        for i, error in enumerate(errors[:3]):
            error_text = str(error).replace("{'error': '", "").replace("'}", "").replace("error", "")[:80]
            lines.append(f"   {i+1}. {error_text}")
        if len(errors) > 3:
            lines.append(f"   ... and {len(errors) - 3} more")
    
    if actions:
        lines.append(f"💡 Suggested Actions:")
        for i, action in enumerate(actions[:3]):
            lines.append(f"   {i+1}. {action}")
    
    return "\n".join(lines)

def get_json_context():
    """Get context as JSON"""
    if not is_argus_available():
        return json.dumps({
            "available": False,
            "message": "Argus not running"
        }, indent=2)
    
    return json.dumps({
        "available": True,
        "summary": get_project_summary(),
        "context": get_claude_context(),
        "actions": get_next_actions()[:5],
        "errors": get_current_errors()[:5]
    }, indent=2)

def main():
    """Main CLI interface"""
    if len(sys.argv) > 1:
        arg = sys.argv[1]
        if arg == "--json":
            print(get_json_context())
        elif arg == "--brief":
            print(get_brief_context())
        elif arg == "--detailed":
            print(get_detailed_context())
        elif arg == "--help":
            print("Argus Context for Claude Code")
            print("Usage: python3 argus_context.py [--json|--brief|--detailed|--help]")
            print("  --json      Output as JSON")
            print("  --brief     One-line summary (default)")
            print("  --detailed  Detailed context with actions")
            print("  --help      Show this help")
        else:
            print(get_brief_context())
    else:
        # Default: brief context
        print(get_brief_context())

if __name__ == "__main__":
    main()