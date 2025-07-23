# Developer Documentation

Welcome to the ISX Daily Reports Scrapper developer documentation. This guide will help you understand the codebase, contribute effectively, and implement new features.

## Quick Links

- [Architecture Overview](ARCHITECTURE.md) - System design and components
- [Build Guide](BUILD_GUIDE.md) - Comprehensive build instructions for all environments
- [Development Tasks](DEVELOPMENT_TASKS.md) - Current task list and priorities
- [Development Roadmap](DEVELOPMENT_ROADMAP.md) - Version milestones and vision
- [Versioning Guide](VERSIONING.md) - Semantic versioning and release process
- [Development Best Practices](DEVELOPMENT_BEST_PRACTICES.md) - Documentation, testing, and development standards

## Getting Started

### 1. Development Setup

```bash
# Clone the repository
git clone https://github.com/haideralmesaody/ISXDailyReportScrapper.git
cd ISXDailyReportScrapper

# Install Go (1.19 or later)
# https://golang.org/dl/

# Build the project (see BUILD_GUIDE.md for detailed instructions)
.\build.bat
```

### 2. Understanding the Codebase

The project follows a modular architecture:

```
dev/
├── cmd/               # Command-line applications
│   ├── indexcsv/      # Index data extractor
│   ├── process/       # Data processor
│   └── web-licensed/  # Web application
├── internal/          # Internal packages
│   ├── common/        # Shared utilities
│   ├── websocket/     # Real-time communication
│   └── ...           # Other domain packages
└── web/              # Frontend assets
```

### 3. Task Management Workflow

Before starting any work:

1. Check [DEVELOPMENT_TASKS.md](DEVELOPMENT_TASKS.md) for available tasks
2. Select a task marked as [READY]
3. Update task status to [IN_PROGRESS]
4. Create a feature branch: `git checkout -b epic-xxx-description`
5. Follow conventional commits (see [VERSIONING.md](VERSIONING.md))

## Key Features

### Real-time Updates
- WebSocket-based communication
- File system watching for data changes
- Automatic UI synchronization

### Data Processing Pipeline
- Excel to CSV conversion
- Historical data management
- Forward-filling for missing data
- Market indices extraction

### Liquidity Scoring (Planned)
- Hybrid liquidity calculation system
- No market cap requirements
- Handles non-trading days intelligently
- See [Liquidity Calculation Approach](../design/LIQUIDITY_CALCULATION_APPROACH.md)

### Trading Strategy Module (Planned)
- Rule-based strategy definition
- Comprehensive backtesting engine
- Liquidity-aware position sizing
- Parameter optimization (grid search, walk-forward)
- Daily trading recommendations
- See [Strategy Module Design](../design/STRATEGY_MODULE_DESIGN.md)

## Development Guidelines

### Code Style
- Follow Go conventions and idioms
- Use existing utilities from `internal/common`
- Maintain backward compatibility
- Write self-documenting code

### Security
- Never log sensitive data (license keys, credentials)
- Use ISX_DEBUG environment variable for debug logging
- All paths should be relative to executable
- Sanitize user inputs

### Testing
```bash
# Run tests
go test ./...

# Run with debug logging
set ISX_DEBUG=true
web-licensed.exe
```

### Data Validation
When working with data:
1. Review [DATA_SPECIFICATIONS.md](../specifications/DATA_SPECIFICATIONS.md)
2. Follow [DATA_VALIDATION_CHECKLIST.md](../specifications/DATA_VALIDATION_CHECKLIST.md)
3. Ensure consistency across components
4. Handle edge cases (BOM, null values, etc.)

## API Development

### Current Endpoints
- `/api/license/*` - License management
- `/api/admin/*` - System administration
- `/ws` - WebSocket connection
- `/health` - Health check
- `/api/version` - Version information

### Adding New Endpoints
1. Define in `web-application.go`
2. Follow RESTful conventions
3. Implement proper error handling
4. Document in API specifications
5. Add to Postman collection (if exists)

## Frontend Development

### Technology Stack
- Vanilla JavaScript (no framework)
- Bootstrap 5 for UI
- Chart.js for visualizations
- WebSocket for real-time updates

### Key Components
- `DataUpdateManager` - Handles real-time data synchronization
- `WebSocketManager` - Manages WebSocket connection
- Pipeline status indicators
- Interactive charts

## Common Tasks

### Adding a New Data Field
1. Update data structures in processor
2. Modify CSV export logic
3. Update frontend to display
4. Add to data specifications
5. Test end-to-end flow

### Implementing a New Chart
1. Add HTML container in `index.html`
2. Create chart initialization function
3. Connect to data update system
4. Handle resize events
5. Add loading states

### Creating a New WebSocket Message Type
1. Define message structure in specifications
2. Implement sender in Go backend
3. Add handler in frontend
4. Update documentation
5. Maintain backward compatibility

## Debugging Tips

### Backend Issues
```bash
# Enable debug logging
set ISX_DEBUG=true

# Check logs
tail -f release/logs/web-licensed.log

# Verify file paths
dir release/data/
```

### Frontend Issues
- Use browser DevTools Console
- Check Network tab for API calls
- Monitor WebSocket messages
- Verify data format matches specs

### Data Issues
- Check for UTF-8 BOM in CSV files
- Verify date formats (YYYY-MM-DD)
- Ensure boolean values are "true"/"false"
- Check for null vs empty arrays

## Contributing

### Before Submitting
1. Run tests and ensure they pass
2. Update documentation if needed
3. Follow commit conventions
4. Update task status in DEVELOPMENT_TASKS.md
5. Create pull request with clear description

### Code Review Checklist
- [ ] Code follows project style
- [ ] Tests are included/updated
- [ ] Documentation is updated
- [ ] No sensitive data in logs
- [ ] Backward compatibility maintained
- [ ] Task status updated

## Resources

### Internal Documentation
- [Architecture](ARCHITECTURE.md)
- [Communication Plan](COMMUNICATION_STANDARDIZATION_PLAN.md)
- [Task Template](TASK_TEMPLATE.md)

### Specifications
- [Data Formats](../specifications/DATA_SPECIFICATIONS.md)
- [WebSocket Protocol](../specifications/WEBSOCKET_MESSAGE_SPECS.md)
- [Liquidity Scoring](../specifications/LIQUIDITY_SCORING_SPECS.md)
- [Strategy Module](../specifications/STRATEGY_MODULE_SPECS.md)

### External Resources
- [Go Documentation](https://golang.org/doc/)
- [WebSocket RFC](https://tools.ietf.org/html/rfc6455)
- [ISX Website](http://www.isx-iq.net/)

## Getting Help

- Check existing documentation first
- Search closed issues on GitHub
- Ask in developer discussions
- Contact maintainers for licensing issues

---

*Last updated: 2025-01-19 | Version: 0.1.0-alpha.1*