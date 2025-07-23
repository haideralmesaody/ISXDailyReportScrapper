# Development Best Practices

## Overview
This document outlines the best practices for developing, documenting, and testing the ISX Daily Reports Scrapper project. All developers should follow these guidelines to ensure consistency and maintainability.

## Documentation Update Process

### When to Update Documentation
Documentation MUST be updated whenever:
1. New features are added
2. APIs or interfaces change
3. Configuration options are added/modified
4. File formats or structures change
5. New packages or modules are created
6. Bug fixes affect user-facing behavior
7. Dependencies are added or updated

### Documentation Checklist
Before marking any development task as DONE, verify:
- [ ] User documentation reflects new features
- [ ] Technical specifications are updated
- [ ] CHANGELOG.md includes the changes with proper version
- [ ] CLAUDE.md is updated for AI assistant context
- [ ] Code comments explain complex logic
- [ ] README files exist in new directories
- [ ] API documentation matches implementation
- [ ] Examples and code snippets are tested
- [ ] User Acceptance Test (UAT) document created in `docs/testing/user-acceptance/`
- [ ] Test plan includes clear success criteria

### Required Documentation Updates by Change Type

#### New Package/Module
1. Update project structure in:
   - CLAUDE.md
   - Main README.md
   - docs/user/README.md (if user-facing)
2. Create package README with:
   - Purpose and overview
   - Usage examples
   - API reference
   - Dependencies
3. Update technical specifications if applicable

#### New Feature
1. Update user documentation:
   - docs/user/README.md with usage instructions
   - Add screenshots if UI changes
2. Add to CHANGELOG.md:
   - Under [Unreleased] or new version
   - Follow Keep a Changelog format
3. Update relevant specifications
4. Add to CLAUDE.md recent updates section
5. Create or update feature documentation

#### API/Interface Changes
1. Update API documentation
2. Update all usage examples
3. Note breaking changes prominently in CHANGELOG
4. Update integration guides
5. Version the API if necessary

#### Bug Fixes
1. Add to CHANGELOG.md under "Fixed"
2. Update troubleshooting guides if relevant
3. Document any workarounds
4. Add test cases to prevent regression

### Documentation Standards

#### Writing Style
1. Use clear, concise language
2. Write in present tense
3. Use active voice
4. Avoid jargon without explanation
5. Include context for decisions

#### Code Examples
1. Test all examples before committing
2. Use realistic scenarios
3. Include error handling
4. Show both basic and advanced usage
5. Comment complex parts

#### Formatting
1. Use Markdown for all documentation
2. Follow consistent heading hierarchy
3. Use code blocks with language hints
4. Include tables for structured data
5. Add diagrams where helpful

## Testing Requirements

### User Acceptance Test (UAT) Creation
For every user-facing feature, create a UAT document:

#### When UAT is Required
- New features visible to end users
- Changes to existing workflows
- Performance improvements users should validate
- UI/UX modifications
- Any change affecting user experience

#### UAT Document Location
Create in: `docs/testing/user-acceptance/UAT_v{version}_{feature}.md`

#### UAT Must Include
1. **Executive Summary** - Non-technical feature explanation
2. **Prerequisites** - Clear setup requirements
3. **Step-by-Step Scenarios** - Detailed instructions with expected results
4. **Success Criteria** - Unambiguous pass/fail conditions
5. **Feedback Form** - Structured way to collect user input
6. **Submission Instructions** - How to report results

### Test Plan Creation
For every significant feature or change, create a test plan that includes:

#### Test Plan Template
```markdown
## Test Plan: [Feature Name]

### Objective
Brief description of what is being tested and why.

### Prerequisites
- List all setup requirements
- Required test data
- System state needed

### Test Cases

#### Test Case 1: [Name]
**Objective**: What this test verifies
**Steps**:
1. Step-by-step instructions
2. Include exact commands or clicks
3. Specify expected inputs

**Expected Results**:
- Specific outcomes to verify
- Include both positive and negative cases
- Error messages if applicable

**Actual Results**: (Fill during testing)
- [ ] Pass
- [ ] Fail (describe issue)

### Success Criteria
- All test cases pass
- Performance meets requirements
- No regression in existing features
```

### Types of Testing

#### Unit Testing
- Test individual functions/methods
- Mock external dependencies
- Aim for >80% code coverage
- Use table-driven tests for multiple scenarios

#### Integration Testing
- Test component interactions
- Verify data flow between modules
- Test with real dependencies
- Check error propagation

#### End-to-End Testing
- Test complete user workflows
- Use production-like data
- Verify all pipeline stages
- Check UI updates

#### Performance Testing
- Measure processing times
- Monitor memory usage
- Test with large datasets
- Compare against baselines

### Test Documentation
1. Document test results with:
   - Date and version tested
   - Environment details
   - Actual vs expected results
   - Screenshots for UI tests
2. Create regression test suite
3. Maintain test data sets
4. Document known issues

## Code Review Process

### Before Submitting
1. Self-review against standards
2. Run all relevant tests
3. Update documentation
4. Check for sensitive data
5. Verify build succeeds

### Review Checklist
- [ ] Code follows project style guide
- [ ] Tests are included and pass
- [ ] Documentation is updated
- [ ] No hardcoded values
- [ ] Error handling is appropriate
- [ ] Performance impact considered
- [ ] Security implications reviewed

## Version Control

### Commit Messages
Follow conventional commits format:
```
type(scope): subject

body (optional)

footer (optional)
```

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation only
- style: Code style changes
- refactor: Code restructuring
- test: Test additions/changes
- chore: Build/tool changes

### Branch Strategy
1. main: Production-ready code
2. develop: Integration branch
3. feature/*: New features
4. fix/*: Bug fixes
5. docs/*: Documentation updates

## Release Process

### Pre-Release Checklist
1. All tests pass
2. Documentation is complete
3. CHANGELOG.md is updated
4. Version numbers are bumped
5. Build artifacts are created

### Release Documentation
1. Create release notes from CHANGELOG
2. Include upgrade instructions
3. List breaking changes prominently
4. Add migration guides if needed

## Architecture Patterns

**IMPORTANT**: All development must follow the three-layer architecture defined in [Architecture Principles](../design/ARCHITECTURE_PRINCIPLES.md).

### Key Rules:
1. **All logic in backend (Go)** - No business logic in frontend
2. **Frontend for display only** - HTML/JS only shows status and collects input
3. **WebSocket for status only** - Never use for control flow
4. **Exit codes determine success** - Not WebSocket messages

### Pipeline Control Pattern:
```go
// CORRECT: Backend controls pipeline
func runPipeline() {
    // Execute stage
    if err := exec.Command("scraper.exe").Run(); err != nil {
        handleError(err)
        return
    }
    
    // Continue to next stage based on exit code
    if err := exec.Command("process.exe").Run(); err != nil {
        handleError(err)
        return
    }
}

// WRONG: Using WebSocket for control
if websocketMsg.Status == "completed" {
    startNextStage() // Never do this!
}
```

## WebSocket Communication Best Practices

When developing executables that communicate via WebSocket:

1. **Always flush stdout before exiting**:
   ```go
   // At the end of main() or before any return/exit
   os.Stdout.Sync()
   ```

2. **Use standardized stage names**:
   ```go
   // CORRECT stage names:
   calc := progress.NewEnhancedCalculator("scraping", totalItems, metricsManager)    // scraper.exe
   calc := progress.NewEnhancedCalculator("processing", totalItems, metricsManager)  // process.exe
   calc := progress.NewEnhancedCalculator("indices", totalItems, metricsManager)     // indexcsv.exe
   ```

3. **Send status updates for UI only**:
   ```go
   // Status messages are for display, not control
   sendStatus("scraping", "completed", "Download finished")
   
   // Progress updates for user feedback
   sendProgress(calc, "Processing...", details)
   ```

4. **Exit codes control pipeline**:
   ```go
   func main() {
       if err := doWork(); err != nil {
           fmt.Printf("Error: %v\n", err)
           os.Exit(1) // Non-zero = failure
       }
       os.Exit(0) // Zero = success
   }
   ```

5. **Handle errors properly**:
   - Send structured error messages with recovery hints
   - Use appropriate error codes
   - Return non-zero exit code for failures

6. **Pipeline transitions in backend only**:
   - Web application controls all stage transitions
   - Based on process exit codes, not WebSocket
   - Executables only report status for display
   - Frontend never controls pipeline flow

## Development Environment

### Required Tools
- Go 1.23+
- Git
- Chrome (for testing)
- Text editor with Go support
- Markdown preview tool

### Recommended Setup
1. Enable Go modules
2. Configure auto-formatting
3. Set up linting tools
4. Install spell checker
5. Use consistent line endings

## Security Considerations

### Code Security
1. Never commit secrets or credentials
2. Use environment variables for config
3. Validate all user inputs
4. Handle errors gracefully
5. Log security events

### Documentation Security
1. Sanitize examples
2. Don't expose internal details
3. Mark security-sensitive sections
4. Review before publishing
5. Version control access

## Continuous Improvement

### Metrics to Track
1. Documentation coverage
2. Test coverage
3. Build success rate
4. Issue resolution time
5. User feedback

### Regular Reviews
1. Monthly documentation audit
2. Quarterly process review
3. Annual best practices update
4. Continuous user feedback
5. Team retrospectives

## Quick Reference

### Essential Files to Update
- [ ] CHANGELOG.md - Version history
- [ ] CLAUDE.md - AI context
- [ ] README.md - Project overview
- [ ] docs/user/README.md - User guide
- [ ] Relevant specifications
- [ ] Test documentation
- [ ] UAT document in docs/testing/user-acceptance/

### Documentation Locations
```
docs/
├── user/                  # End-user guides
├── developer/            # Developer docs
├── specifications/       # Technical specs
├── design/              # Architecture
├── operations/          # Deployment
└── reference/           # References
```

### Help and Resources
- Keep a Changelog: https://keepachangelog.com
- Semantic Versioning: https://semver.org
- Markdown Guide: https://www.markdownguide.org
- Go Documentation: https://go.dev/doc

---

Remember: Good documentation is as important as good code. It ensures the project remains maintainable and accessible to all stakeholders.