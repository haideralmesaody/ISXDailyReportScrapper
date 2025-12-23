## Summary
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that causes existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring

## Related Issues
Fixes #(issue number)

## Testing
- [ ] Tests pass locally (`./build.bat -target=test`)
- [ ] Added new tests for changes
- [ ] Updated existing tests as needed
- [ ] Coverage remains above 80% (90% for critical paths)

## Checklist
- [ ] Code follows CLAUDE.md standards
- [ ] Used Chi v5 router (no Gin/Echo)
- [ ] Used slog for logging (no fmt.Println)
- [ ] Errors follow RFC 7807 format
- [ ] Context passed as first parameter
- [ ] No builds in api/ or web/ directories
- [ ] Documentation updated (README, API docs)
- [ ] No sensitive data in commits
- [ ] Security implications reviewed

## Screenshots (if applicable)
Add screenshots for UI changes

## Additional Notes
Any other information that reviewers should know