# Feature Development Template

## Feature Overview
**Name**: [Feature Name]  
**Epic**: [EPIC-XXX]  
**Priority**: [P0/P1/P2/P3]  
**Estimated Time**: [X hours/days]  

**Description**:
[Brief description of what the feature does and why it's needed]

## Architecture Compliance Checklist

Before starting development, confirm:
- [ ] I have read [Architecture Principles](../design/ARCHITECTURE_PRINCIPLES.md)
- [ ] I understand the three-layer architecture (Frontend, Backend, WebSocket)
- [ ] I know that all logic must be in backend Go code
- [ ] I understand WebSocket is for status updates only

## Design

### User Interface Changes
- [ ] New UI elements needed? (describe)
- [ ] Changes to existing UI? (describe)
- [ ] All UI changes are display-only (no logic)

### Backend Logic
- [ ] New Go packages/modules needed?
- [ ] Changes to existing backend logic?
- [ ] New pipeline stages?
- [ ] External dependencies?

### Data Flow
```
User Action → Frontend → HTTP API → Backend Processing → Status Updates via WebSocket
```

Specific flow for this feature:
1. [Step 1]
2. [Step 2]
3. ...

### API Endpoints
- [ ] New endpoints needed?
  - `POST /api/[endpoint]` - [Description]
  - `GET /api/[endpoint]` - [Description]

### WebSocket Messages
- [ ] New message types needed?
  - Type: `[message_type]`
  - Purpose: Status updates only
  - Data structure:
    ```json
    {
      "type": "[message_type]",
      "data": {
        // fields
      }
    }
    ```

## Implementation Plan

### Phase 1: Backend Implementation
1. [ ] Create/modify Go modules
2. [ ] Implement business logic
3. [ ] Add error handling
4. [ ] Use proper exit codes
5. [ ] Send status updates via stdout

### Phase 2: Web Application Integration
1. [ ] Add HTTP endpoint handlers
2. [ ] Execute processes with proper error checking
3. [ ] Broadcast status updates via WebSocket
4. [ ] Control pipeline flow based on exit codes

### Phase 3: Frontend Updates
1. [ ] Add UI elements for user input
2. [ ] Display status updates from WebSocket
3. [ ] No business logic in JavaScript
4. [ ] Handle connection/reconnection

### Phase 4: Testing
1. [ ] Unit tests for backend logic
2. [ ] Integration tests for pipeline
3. [ ] Manual testing of UI
4. [ ] Error scenario testing

## Code Structure

### Backend Structure
```
dev/
├── cmd/
│   └── [newfeature]/      # If new executable
│       └── main.go
├── internal/
│   └── [feature]/         # New internal package
│       ├── handler.go     # Business logic
│       └── types.go       # Data structures
└── web-application.go     # Integration point
```

### Frontend Structure
```
dev/web/
├── index.html            # UI elements
└── static/
    └── [feature].js      # Display logic only (if needed)
```

## Error Handling

### Backend Errors
- Use appropriate exit codes:
  - 0 = Success
  - 1 = General failure
  - 2+ = Specific errors (document)

### Status Communication
- Send clear error messages via WebSocket
- Include recovery hints
- Mark errors as recoverable/non-recoverable

## Testing Strategy

### Unit Tests
```go
func TestFeatureName(t *testing.T) {
    // Test individual functions
}
```

### Integration Tests
```go
func TestFeaturePipeline(t *testing.T) {
    // Test complete flow
}
```

### Manual Test Cases
1. [ ] Happy path scenario
2. [ ] Error scenarios
3. [ ] Edge cases
4. [ ] Performance under load

## Documentation Updates

### Required Documentation
- [ ] Update user guide (docs/user/README.md)
- [ ] Update technical specs if applicable
- [ ] Add to CHANGELOG.md
- [ ] Update CLAUDE.md if significant
- [ ] Create UAT document if user-facing

### Code Documentation
- [ ] Add package comments
- [ ] Document exported functions
- [ ] Include usage examples

## Security Considerations
- [ ] No hardcoded credentials
- [ ] Validate all inputs
- [ ] Log security events
- [ ] Handle sensitive data properly

## Performance Considerations
- [ ] Expected data volumes
- [ ] Processing time estimates
- [ ] Memory usage concerns
- [ ] Concurrent execution needs

## Deployment Notes
- [ ] Configuration changes needed?
- [ ] Database migrations?
- [ ] Backward compatibility maintained?
- [ ] Version bump required?

## Rollback Plan
- [ ] How to disable feature if issues?
- [ ] Data cleanup needed?
- [ ] Revert procedure documented?

## Success Criteria
- [ ] Feature works as designed
- [ ] All tests pass
- [ ] Documentation complete
- [ ] No regression in existing features
- [ ] Performance acceptable

## Review Checklist

### Self-Review
- [ ] Code follows architecture principles
- [ ] All logic in backend
- [ ] WebSocket for status only
- [ ] Proper error handling
- [ ] Tests included
- [ ] Documentation updated

### Peer Review
- [ ] Architecture compliance verified
- [ ] Code quality acceptable
- [ ] Tests adequate
- [ ] Documentation clear
- [ ] Security concerns addressed

## Notes
[Any additional notes, assumptions, or concerns]

---

**Remember**: 
- Backend controls everything
- Frontend displays only
- WebSocket communicates status only
- Exit codes determine success/failure