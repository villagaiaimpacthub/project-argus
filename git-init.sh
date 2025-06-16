#!/bin/bash

# Initialize git repository for Claude Code / Project Argus

echo "Initializing git repository for Claude Code / Project Argus..."

# Remove any existing .git directory
if [ -d ".git" ]; then
    echo "Removing existing .git directory..."
    rm -rf .git
fi

# Initialize git
git init --initial-branch=main 2>/dev/null || git init -b main

# Set git config that works with WSL
git config core.filemode false
git config core.autocrlf false
git config core.eol lf

echo "Git repository initialized successfully!"