# Secure Development Workflow - Quick Reference

## ⚠️ MANDATORY: Follow this workflow for ALL development

### 🚀 Quick Start Commands

#### 1. Start New Feature
```bash
.\scripts\new-feature.bat my-feature-name
```

#### 2. Before EVERY Commit
```bash
.\scripts\sanitize-credentials.bat
git add .
git commit -m "feat: Your message"
```

#### 3. After Merge
```bash
git checkout main && git pull
.\scripts\restore-credentials.bat
```

## 📋 Complete Workflow

### Step 1: Create Feature Branch
```bash
git checkout main
git pull origin main
git checkout -b feature/task-name
```

### Step 2: Develop & Test
- Work normally - your credentials.json will be used
- Test thoroughly before committing

### Step 3: Sanitize Before Commit
```bash
# CRITICAL: Always run this!
.\scripts\sanitize-credentials.bat
```

### Step 4: Commit & Push
```bash
git add .
git commit -m "feat: Description"
git push origin feature/task-name
```

### Step 5: Create Pull Request
- Go to GitHub
- Create PR to main branch
- Wait for review/approval

### Step 6: After Merge
```bash
git checkout main
git pull origin main
.\scripts\restore-credentials.bat
git branch -d feature/task-name
```

## 🔴 Emergency: Committed Credentials?

### If NOT pushed yet:
```bash
git reset --soft HEAD~1
.\scripts\sanitize-credentials.bat
git add .
git commit -m "feat: Your message (sanitized)"
```

### If already pushed:
1. **STOP** - Don't push more commits
2. Contact team immediately
3. Rotate the exposed credentials
4. Follow credential rotation procedure

## 📁 Files That Need Sanitization
- `credentials.json` (never commit)
- `sheets-config.json` (never commit)
- `dev/internal/license/manager.go`
- `internal/license/license.go`

## 🛠️ Helper Scripts
All in `scripts/` directory:
- `sanitize-credentials.bat` - Removes credentials before commit
- `restore-credentials.bat` - Restores after merge
- `check-credentials.bat` - Verifies no exposed secrets
- `new-feature.bat` - Creates new feature branch

## ✅ Pre-commit Hook
A pre-commit hook is installed to catch credentials.
If commit fails with credential warning:
1. Run `.\scripts\sanitize-credentials.bat`
2. Try commit again

## 📚 Full Documentation
See `docs/developer/SECURE_DEVELOPMENT_WORKFLOW.md` for detailed guide.

---
**Remember**: NEVER commit real credentials. When in doubt, sanitize!