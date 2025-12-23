# Agent Orchestration - Multi-Agent Coordination Patterns

## Purpose
Advanced coordination patterns for using multiple Claude Code agents together on complex tasks, enabling seamless collaboration between specialized agents for the ISX Pulse platform.

## Orchestration Principles

### 1. Primary-Agent-First Selection
Always identify the primary domain first, then coordinate specialized support:

```typescript
interface AgentTask {
  primaryAgent: string
  supportingAgents?: string[]
  coordination: AgentCoordination
  context: ISXContext
}

interface AgentCoordination {
  sequence: 'sequential' | 'parallel' | 'iterative'
  handoff: string[] // What data to pass between agents
  validation: ValidationRule[]
}
```

### 2. ISX Context Preservation
Ensure all agents understand the ISX Pulse context:

```typescript
interface ISXContext {
  architecture: 'pipeline-based-precalculation'
  storage: 'csv-based-with-indexing'
  market: 'iraqi-stock-exchange'
  timeframe: 'eod-only'
  licensing: 'hardware-locked'
  compliance: 'iraqi-financial-regulations'
}
```

## Common Orchestration Patterns

### Pattern 1: Feature Development Workflow

#### Scenario: Adding New Technical Indicator
```yaml
workflow: new-indicator-development
sequence: sequential
agents:
  primary: go-architect
  coordination:
    - agent: go-architect
      task: "Design indicator calculation architecture"
      output: "architecture-design.md"

    - agent: isx-data-specialist
      task: "Validate ISX market data compatibility"
      input: "architecture-design.md"
      output: "data-validation.md"

    - agent: technical-indicators-skill
      task: "Implement indicator calculation logic"
      input: ["architecture-design.md", "data-validation.md"]
      output: "indicator-implementation.go"

    - agent: test-architect
      task: "Create comprehensive test suite"
      input: "indicator-implementation.go"
      output: "test-suite.go"

    - agent: performance-profiler
      task: "Optimize calculation performance"
      input: ["indicator-implementation.go", "test-suite.go"]
      output: "optimized-implementation.go"

    - agent: security-auditor
      task: "Security review of implementation"
      input: "optimized-implementation.go"
      output: "security-review.md"

context:
  ssot: "ticker_indicators.csv"
  pipeline_stage: "indicators-stage"
  performance_requirement: "<1ms access time"
```

### Pattern 2: API Development Workflow

#### Scenario: New API Endpoint for Market Data
```yaml
workflow: api-endpoint-development
sequence: parallel-for-validation, sequential-for-implementation
agents:
  primary: api-contract-guardian
  coordination:
    parallel_validation:
      - agent: api-contract-guardian
        task: "Design API contract and schema"
        output: "api-contract.md"

      - agent: security-auditor
        task: "Define security requirements"
        output: "security-requirements.md"

      - agent: license-system-engineer
        task: "Define license requirements for endpoint"
        output: "license-requirements.md"

    sequential_implementation:
      - agent: go-architect
        task: "Implement backend handler"
        input: ["api-contract.md", "security-requirements.md", "license-requirements.md"]
        output: "handler.go"

      - agent: frontend-modernizer
        task: "Create frontend integration"
        input: ["api-contract.md", "handler.go"]
        output: "frontend-client.tsx"

      - agent: react-hydration-guardian
        task: "Ensure SSR/hydration compatibility"
        input: "frontend-client.tsx"
        output: "hydrated-client.tsx"

      - agent: test-architect
        task: "Create integration tests"
        input: ["handler.go", "hydrated-client.tsx"]
        output: "integration-tests.go"

context:
  build_rules: "strict-enforcement"
  error_format: "rfc-7807-problem-details"
  authentication: "license-based"
```

### Pattern 3: Performance Optimization Workflow

#### Scenario: Optimizing Indicator Calculation Pipeline
```yaml
workflow: performance-optimization
sequence: iterative
agents:
  primary: performance-profiler
  coordination:
    iterative_cycle:
      - agent: performance-profiler
        task: "Profile current performance"
        output: "performance-baseline.md"

      - agent: thread-safety-skill
        task: "Analyze concurrency bottlenecks"
        input: "performance-baseline.md"
        output: "concurrency-analysis.md"

      - agent: file-storage-optimizer
        task: "Optimize CSV storage and indexing"
        input: ["performance-baseline.md", "concurrency-analysis.md"]
        output: "storage-optimization.go"

      - agent: operation-orchestrator
        task: "Optimize pipeline execution"
        input: "storage-optimization.go"
        output: "pipeline-optimization.go"

      - agent: metrics-analyst
        task: "Define performance metrics and SLOs"
        input: "pipeline-optimization.go"
        output: "metrics-definition.md"

      - agent: observability-engineer
        task: "Implement monitoring and alerting"
        input: "metrics-definition.md"
        output: "monitoring-setup.go"

validation:
  - "sub-millisecond indicator access"
  - "100% CPU utilization efficiency"
  - "memory usage < 100MB"
  - "concurrent user support > 100"

context:
  target_performance: "sub-millisecond indicator access"
  current_baseline: "5ms average response time"
```

## Multi-Agent Coordination API

### Agent Selection Helper
```go
type AgentOrchestrator struct {
    agents map[string]Agent
    skills map[string]Skill
}

func (ao *AgentOrchestrator) SelectAgents(task Task) (*AgentPlan, error) {
    // Analyze task requirements
    requirements := ao.analyzeRequirements(task)

    // Select primary agent
    primaryAgent := ao.selectPrimaryAgent(requirements)

    // Select supporting agents
    supportingAgents := ao.selectSupportingAgents(requirements, primaryAgent)

    // Create coordination plan
    plan := &AgentPlan{
        PrimaryAgent: primaryAgent,
        SupportingAgents: supportingAgents,
        Sequence: ao.determineSequence(requirements),
        Context: ao.createISXContext(),
    }

    return plan, nil
}
```

### Context Passing Between Agents
```go
type AgentContext struct {
    TaskID        string
    PrimaryAgent  string
    ISXContext    ISXContext
    SharedData    map[string]interface{}
    HandoffData   map[string]HandoffData
    Validation    ValidationRules
}

type HandoffData struct {
    FromAgent    string
    ToAgent      string
    Data         interface{}
    Requirements []string
    Validation   []ValidationRule
}
```

## Real-World Orchestration Examples

### Example 1: Adding RSI Divergence Detection

**Task**: Add RSI divergence detection to the technical indicators system

```typescript
const rsiDivergenceWorkflow = {
  task: "implement-rsi-divergence-detection",
  primaryAgent: "technical-indicators-skill",
  coordination: [
    {
      agent: "technical-indicators-skill",
      task: "Design RSI divergence algorithm",
      context: "Use existing RSI calculations from ticker_indicators.csv",
      output: "divergence-algorithm.md"
    },
    {
      agent: "go-architect",
      task: "Integrate divergence detection into indicators stage",
      input: "divergence-algorithm.md",
      context: "Must maintain sub-millisecond performance",
      output: "divergence-integration.go"
    },
    {
      agent: "test-architect",
      task: "Create test cases for divergence scenarios",
      input: "divergence-integration.go",
      context: "Test bullish/bearish divergence patterns",
      output: "divergence-tests.go"
    },
    {
      agent: "performance-profiler",
      task: "Benchmark performance impact",
      input: ["divergence-integration.go", "divergence-tests.go"],
      context: "Must not impact existing indicator performance",
      output: "performance-benchmark.md"
    },
    {
      agent: "frontend-modernizer",
      task: "Add divergence visualization to charts",
      input: "divergence-algorithm.md",
      context: "Use TradingView Lightweight Charts",
      output: "divergence-chart-component.tsx"
    }
  ],
  validation: [
    "Performance < 0.1ms additional latency",
    "Accuracy > 95% for divergence detection",
    "UI responsiveness maintained"
  ]
}
```

### Example 2: Enhancing License Security

**Task**: Enhance license system with hardware fingerprinting improvements

```typescript
const licenseSecurityWorkflow = {
  task: "enhance-license-security",
  primaryAgent: "license-system-engineer",
  coordination: [
    {
      agent: "license-system-engineer",
      task: "Design enhanced hardware fingerprinting",
      context: "Improve device binding accuracy",
      output: "enhanced-fingerprinting.md"
    },
    {
      agent: "security-auditor",
      task: "Security review of fingerprinting approach",
      input: "enhanced-fingerprinting.md",
      context: "OWASP compliance and anti-tampering",
      output: "security-review.md"
    },
    {
      agent: "go-architect",
      task: "Implement secure fingerprinting module",
      input: ["enhanced-fingerprinting.md", "security-review.md"],
      context: "Cross-platform compatibility",
      output: "fingerprinting-module.go"
    },
    {
      agent: "test-architect",
      task: "Create security test scenarios",
      input: "fingerprinting-module.go",
      context: "Test bypass attempts and edge cases",
      output: "security-tests.go"
    },
    {
      agent: "compliance-regulator",
      task: "Validate Iraqi regulatory compliance",
      input: "security-tests.go",
      context: "Financial software regulations",
      output: "compliance-report.md"
    }
  ],
  validation: [
    "Bypass resistance > 99.9%",
    "False positive rate < 0.1%",
    "Cross-platform support maintained"
  ]
}
```

## Error Handling in Multi-Agent Workflows

### Agent Failure Recovery
```go
type AgentFailure struct {
    Agent    string
    Task     string
    Error    error
    Context  AgentContext
    Recovery RecoveryStrategy
}

type RecoveryStrategy struct {
    Type        string // "retry", "fallback", "skip", "abort"
    MaxRetries  int
    FallbackAgent string
    ContinueOnError bool
}

func (ao *AgentOrchestrator) HandleAgentFailure(failure AgentFailure) error {
    switch failure.Recovery.Type {
    case "retry":
        return ao.retryAgentTask(failure)
    case "fallback":
        return ao.useFallbackAgent(failure)
    case "skip":
        return ao.skipAgentTask(failure)
    case "abort":
        return fmt.Errorf("workflow aborted: %w", failure.Error)
    }
    return nil
}
```

### Workflow State Management
```go
type WorkflowState struct {
    ID           string
    Status       string // "running", "completed", "failed", "paused"
    CurrentAgent string
    CompletedAgents []string
    Context      AgentContext
    Results      map[string]interface{}
    Errors       []AgentFailure
    StartTime    time.Time
    EndTime      time.Time
}

func (ws *WorkflowState) UpdateAgentResult(agent string, result interface{}) {
    ws.Results[agent] = result
    ws.CompletedAgents = append(ws.CompletedAgents, agent)
}

func (ws *WorkflowState) IsCompleted() bool {
    return ws.Status == "completed" || ws.Status == "failed"
}
```

## Best Practices for Agent Orchestration

### DO ✅
1. **Always identify primary domain first** - Choose main agent based on core task
2. **Preserve ISX context throughout** - Ensure all agents understand platform specifics
3. **Validate handoff data** - Ensure data integrity between agent transitions
4. **Plan for failures** - Have fallback strategies for agent failures
5. **Measure orchestration overhead** - Track time spent in coordination
6. **Document agent decisions** - Keep record of why specific agents were chosen

### DON'T ❌
1. **Don't over-coordinate** - Keep workflows simple and direct
2. **Don't lose context** - Maintain ISX-specific knowledge throughout workflow
3. **Don't ignore performance** - Multi-agent workflows should be efficient
4. **Don't create circular dependencies** - Avoid agent dependency loops
5. **Don't skip validation** - Validate each agent's output before handoff
6. **Don't forget error handling** - Plan for agent failures and recovery

This orchestration system enables complex multi-agent collaborations while maintaining the specialized expertise and ISX Pulse context awareness required for financial platform development.