# Secure Development Workflow

## Overview
This document outlines the mandatory workflow for all development to ensure credentials never get committed to the repository.

## The Golden Rules
1. **NEVER work directly on main branch**
2. **ALWAYS create a feature branch for new work**
3. **ALWAYS sanitize credentials before committing**
4. **ALWAYS test locally before pushing**
5. **NEVER commit credentials.json or any file with secrets**

## Step-by-Step Workflow

### 1. Starting a New Task

```bash
# Ensure you're on main and it's up to date
git checkout main
git pull origin main

# Create a new feature branch
git checkout -b feature/task-name
# Or for bugs:
git checkout -b fix/bug-name
```

### 2. Development Phase

Work on your task normally. Your local credentials will continue to work:
- `credentials.json` is loaded automatically
- The app functions normally during development

### 3. Before Committing - Sanitize Credentials

**CRITICAL**: Before any commit, run the sanitization script:

```bash
# Windows
.\scripts\sanitize-credentials.bat

# Git Bash
./scripts/sanitize-credentials.sh
```

This script will:
1. Backup your credential files
2. Replace credentials with placeholders
3. Stage the sanitized files

### 4. Commit Your Changes

```bash
# Add your changes (credentials are already sanitized)
git add .

# Commit with descriptive message
git commit -m "feat: Add new feature X

- Implemented feature X
- Added tests for feature X
- Updated documentation

Credentials sanitized for security"
```

### 5. Push to GitHub

```bash
# Push your feature branch
git push origin feature/task-name
```

### 6. Create Pull Request

1. Go to GitHub
2. Create a Pull Request from your branch to main
3. Ensure all checks pass
4. Request review if needed

### 7. After Merge - Restore Credentials

Once your PR is merged:

```bash
# Switch back to main
git checkout main
git pull origin main

# Restore your credentials
.\scripts\restore-credentials.bat  # Windows
./scripts/restore-credentials.sh    # Git Bash

# Delete the feature branch
git branch -d feature/task-name
```

## Quick Reference Commands

### Start New Task
```bash
git checkout main && git pull
git checkout -b feature/new-task
```

### Before Commit
```bash
.\scripts\sanitize-credentials.bat  # Always run this!
git add .
git commit -m "feat: Your message"
```

### After Merge
```bash
git checkout main && git pull
.\scripts\restore-credentials.bat
git branch -d feature/new-task
```

## Credential Files to Watch

These files should NEVER be committed with real credentials:
- `credentials.json`
- `sheets-config.json` 
- `dev/internal/license/manager.go` (line ~110-125)
- `internal/license/license.go` (line ~537-549)
- Any file matching pattern: `*credentials*`, `*secret*`, `*key*`

## Emergency: If You Accidentally Commit Credentials

**DO NOT PUSH!** If you haven't pushed yet:

```bash
# Undo the last commit but keep changes
git reset --soft HEAD~1

# Run sanitization
.\scripts\sanitize-credentials.bat

# Recommit
git add .
git commit -m "feat: Your message (sanitized)"
```

If you already pushed, contact the team immediately to:
1. Rotate the compromised credentials
2. Clean the repository history

## Automated Checks

The following automated checks are in place:
1. Pre-commit hooks check for credential patterns
2. GitHub push protection blocks known secrets
3. Pull request checks verify no credentials

## Script Locations

All helper scripts are in the `scripts/` directory:
- `sanitize-credentials.bat/.sh` - Sanitizes credentials before commit
- `restore-credentials.bat/.sh` - Restores credentials after merge
- `check-credentials.bat/.sh` - Checks if credentials are exposed
- `new-feature.bat/.sh` - Creates new feature branch with setup

## Tips for Success

1. **Set up aliases** for common commands:
   ```bash
   git config --global alias.sanitize "!./scripts/sanitize-credentials.sh"
   git config --global alias.restore "!./scripts/restore-credentials.sh"
   ```

2. **Use the pre-commit hook** - It will warn you about credentials

3. **Test locally** - Always ensure your changes work with credentials before sanitizing

4. **Keep backups** - The scripts create backups, but keep your own too

## Questions?

If you're unsure about any step, ask before proceeding. It's better to be safe than to expose credentials.