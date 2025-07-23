# Task Template

Use this template when adding new tasks to DEVELOPMENT_TASKS.md

## Single Task Template
```markdown
- [EPIC-XXX] Task title here (Est: Xh) [STATE] [Assignee]
  - Brief description of what needs to be done
  - Key implementation details or requirements
  - **Files affected:** List of specific files that will be modified
  - **Testing approach:** Unit/Integration/E2E/Manual
  - **Documentation updates:** Which docs need updating
  - Dependencies: EPIC-YYY, EPIC-ZZZ (if any)
  - Actual time: Xh (filled when completed)
```

## Feature/Component Template
```markdown
### X.Y Feature/Component Name 🟡 P1
- [EPIC-XXX] First task in this feature (Est: Xh) [BACKLOG]
  - Implementation details
  - Specific requirements
  - Dependencies: None

- [EPIC-XXX] Second task in this feature (Est: Xh) [BACKLOG]
  - Implementation details
  - Specific requirements
  - Dependencies: EPIC-XXX (previous task)
```

## New Epic Template
```markdown
## X. Epic Name (EPIC)

### X.1 Feature Area 🟡 P1
- [EPIC-001] First task (Est: Xh) [BACKLOG]
  - Task details
  - Dependencies: None

### X.2 Another Feature Area 🟢 P2
- [EPIC-002] Another task (Est: Xh) [BACKLOG]
  - Task details
  - Dependencies: EPIC-001
```

## Priority Guidelines
- 🔴 **P0 - CRITICAL**: Production issues, security vulnerabilities, data loss risks
- 🟡 **P1 - HIGH**: Core features, significant bugs, customer-impacting issues
- 🟢 **P2 - MEDIUM**: Important improvements, minor bugs, nice-to-have features
- 🔵 **P3 - LOW**: Optimizations, technical debt, documentation

## State Guidelines
- **BACKLOG**: Default state for new tasks
- **READY**: Move here when task is fully defined and dependencies are met
- **IN_PROGRESS**: When you start working on it (add your name)
- **IN_REVIEW**: Code complete, awaiting review
- **DONE**: Fully tested and deployed
- **BLOCKED**: Document what's blocking
- **WONT_DO**: Document why it was cancelled

## Time Estimation Guidelines
- **Simple changes**: 1-2h (config updates, small fixes)
- **Standard features**: 3-5h (new endpoints, UI components)
- **Complex features**: 6-10h (integrations, major refactoring)
- **Large features**: 10-20h (new systems, architectural changes)
- Always round up and add buffer for testing

## Task ID Assignment
1. Check the last used number in the epic
2. Increment by 1
3. Ensure no duplicates
4. Format: [EPIC-XXX] where XXX is zero-padded (001, 002, etc.)

## Architecture Compliance
Before creating or starting any task, ensure:
- [ ] Task follows [Architecture Principles](../design/ARCHITECTURE_PRINCIPLES.md)
- [ ] All logic will be implemented in backend (Go)
- [ ] Frontend changes are display-only
- [ ] WebSocket usage is for status updates only
- [ ] Pipeline control uses exit codes, not WebSocket

## Example: Adding a New Task

### Before:
```markdown
### 2.1 Advanced Analytics 🟢 P2
- [DATA-001] Add moving averages to ticker charts (Est: 4h) [BACKLOG]
- [DATA-002] Volume analysis charts (Est: 3h) [BACKLOG]
```

### After (adding new task):
```markdown
### 2.1 Advanced Analytics 🟢 P2
- [DATA-001] Add moving averages to ticker charts (Est: 4h) [BACKLOG]
- [DATA-002] Volume analysis charts (Est: 3h) [BACKLOG]
- [DATA-003] Add Bollinger Bands indicator (Est: 3h) [BACKLOG]
  - Upper and lower bands based on standard deviation
  - Configurable period and deviation multiplier
  - Dependencies: DATA-001
```

## Task Review Checklist
When moving task to IN_REVIEW:
- [ ] Code follows architecture principles
- [ ] Unit tests included (if applicable)
- [ ] Documentation updated
- [ ] No hardcoded values
- [ ] Proper error handling with exit codes
- [ ] WebSocket messages for display only
- [ ] Backend controls all logic
- [ ] Files affected list is accurate
- [ ] Actual time tracked for estimation improvement
- [ ] No file conflicts with parallel work
- [ ] Testing approach executed successfully

## Commit Message Format
When updating task status, use this commit format:
```
task: [EPIC-XXX] Update status to IN_PROGRESS

- Started work on: Task title here
- Estimated time: Xh
- Dependencies: List any blocking tasks
```

## Common Mistakes to Avoid
1. **Putting logic in frontend** - All business logic must be in Go
2. **Using WebSocket for control** - Use exit codes for pipeline flow
3. **Frontend triggering stages** - Backend controls everything
4. **Missing documentation** - Update docs with code changes
5. **Skipping architecture review** - Check compliance before starting