# Contributing to ISX Pulse

Thank you for your interest in contributing to the ISX Pulse project! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and professional environment for all contributors.

## Getting Started

1. **Fork the repository** and clone your fork locally
2. **Set up your development environment**:
   - Install Go 1.21+
   - Install Node.js 18+
   - Copy `config/credentials.json.example` to `credentials.json`
   - Copy `.env.example` to `.env` and configure

## Development Workflow

### 1. Create a Feature Branch
```bash
git checkout -b feature/your-feature-name
```

### 2. Follow Project Standards

#### Build Rules (MANDATORY)
- **NEVER** build in the `api/` or `web/` directories
- **ALWAYS** use `./build.bat` from project root
- See `BUILD_RULES.md` for details

#### Code Standards
- **Go**: Follow CLAUDE.md standards (Chi v5, slog, RFC 7807 errors)
- **TypeScript**: Strict mode, no `any` types
- **React**: Use hooks, Shadcn/ui components, handle hydration properly

### 3. Write Tests
- Minimum 80% coverage for new code
- 90% for critical paths (licensing, operations)
- Run tests with: `./build.bat -target=test`

### 4. Update Documentation
- Update relevant README files
- Add entries to CHANGELOG.md
- Document new features/APIs

## Commit Guidelines

### Commit Message Format
```
type(scope): description

- Detail 1
- Detail 2

Fixes #issue
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Build/tooling changes

### Examples
```
feat(operations): add retry logic for failed steps
fix(websocket): handle connection drops gracefully
docs(api): update endpoint documentation
```

## Pull Request Process

1. **Ensure all tests pass**:
   ```bash
   ./build.bat -target=test
   ```

2. **Update documentation**:
   - Update CHANGELOG.md
   - Update relevant README files
   - Add/update API documentation

3. **Create Pull Request**:
   - Use a clear, descriptive title
   - Reference related issues
   - Provide detailed description
   - Include test plan

### PR Template
```markdown
## Summary
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added new tests
- [ ] Updated existing tests

## Checklist
- [ ] Code follows CLAUDE.md standards
- [ ] Tests achieve >80% coverage
- [ ] Documentation updated
- [ ] No sensitive data in commits
- [ ] Used ./build.bat for builds
```

## Architecture Guidelines

### Backend (Go)
- Use Chi v5 router exclusively
- Implement structured logging with slog
- Follow RFC 7807 for error responses
- Use dependency injection
- Pass context.Context as first parameter

### Frontend (Next.js/React)
- Use TypeScript strict mode
- Implement proper hydration handling
- Use Shadcn/ui components
- Follow component architecture patterns

### Security
- Never commit credentials or secrets
- Use encrypted configuration
- Implement proper input validation
- Follow OWASP guidelines

## Testing Guidelines

### Unit Tests
- Table-driven tests for Go
- Mock external dependencies
- Test error conditions
- Use meaningful test names

### Integration Tests
- Test API endpoints
- Verify WebSocket functionality
- Test license activation flows

### Frontend Tests
- Component tests with React Testing Library
- E2E tests with Playwright
- Accessibility tests (WCAG 2.1 AA)

## Code Review Process

Reviewers will check:
1. **Functionality**: Does it work as intended?
2. **Tests**: Adequate coverage and quality?
3. **Documentation**: Is it clear and complete?
4. **Standards**: Follows CLAUDE.md guidelines?
5. **Security**: No vulnerabilities introduced?
6. **Performance**: No degradation?

## Using AI Assistants

When using Claude Code or other AI assistants:
1. Reference `.claude/agents/` for specialized help
2. Always verify generated code
3. Ensure AI follows BUILD_RULES.md
4. Review security implications

## Common Tasks

### Adding a New API Endpoint
1. Define contracts in `api/pkg/contracts/`
2. Implement handler in `api/internal/transport/http/`
3. Add service logic in `api/internal/services/`
4. Write comprehensive tests
5. Update API documentation

### Adding a Frontend Component
1. Create component in `web/components/`
2. Use TypeScript interfaces
3. Handle hydration properly
4. Add component tests
5. Update Storybook (if applicable)

### Modifying Operations
1. Update operation types in `api/internal/operations/`
2. Implement stage logic
3. Add WebSocket updates
4. Write integration tests
5. Update operation documentation

## Getting Help

- Check existing issues and discussions
- Review documentation in `docs/`
- Consult CLAUDE.md for standards
- Use appropriate `.claude/agents/` for guidance

## License

By contributing, you agree that your contributions will be licensed under the project's license.

## Questions?

If you have questions, please:
1. Check the documentation
2. Search existing issues
3. Create a new issue with the question label

Thank you for contributing to ISX Pulse!