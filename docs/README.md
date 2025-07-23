# Documentation Directory

This directory contains all project documentation organized by purpose and audience.

## Directory Structure

```
docs/
├── README.md                    # This file - documentation index
├── user/                        # End-user documentation
│   ├── README.md               # User guide
│   └── INSTALLATION.md         # Installation instructions
├── developer/                   # Developer documentation
│   ├── README.md               # Developer guide overview
│   ├── ARCHITECTURE.md         # System architecture
│   ├── BUILD_GUIDE.md          # Build from source guide
│   ├── DEVELOPMENT_TASKS.md    # Current task list and priorities
│   ├── DEVELOPMENT_ROADMAP.md  # Version milestones and vision
│   ├── DEVELOPMENT_BEST_PRACTICES.md # Development, documentation, and testing standards
│   ├── TASK_TEMPLATE.md        # Template for new tasks
│   ├── FEATURE_TEMPLATE.md     # Template for feature development
│   ├── VERSIONING.md           # Semantic versioning guide
│   ├── PIPELINE_MANAGER_GUIDE.md # Pipeline manager documentation
│   ├── COMMUNICATION_STANDARDIZATION_PLAN.md # WebSocket standards
│   └── TEST_PLAN_WEBSOCKET_STANDARDIZATION.md # Testing plans
├── testing/                     # Testing documentation
│   ├── README.md               # Testing guide and standards
│   ├── PIPELINE_TEST_CHECKLIST.md # Pipeline testing guidelines
│   ├── integration/            # Integration test documents
│   │   └── INT_v0.3.0_Pipeline_Manager.md
│   ├── regression/             # Regression test documents
│   │   └── REG_v0.3.0_Pipeline_Orchestration.md
│   └── user-acceptance/        # User acceptance test scenarios
│       ├── UAT_v0.2.0_WebSocket_Progress_Tracking.md
│       ├── UAT_v0.3.0_Pipeline_Manager_Fix.md
│       ├── ISSUE_REPORT_v0.2.0.md
│       ├── QUICK_TEST_GUIDE_v0.2.0.md
│       └── UAT_CORRECTIONS_SUMMARY.md
├── specifications/              # Technical specifications
│   ├── DATA_SPECIFICATIONS.md  # Data format specifications
│   ├── COLUMN_NAME_MAPPING.md  # Column name transformations
│   ├── WEBSOCKET_MESSAGE_SPECS.md # WebSocket message specifications
│   ├── WEBSOCKET_MESSAGES_COMPLETE.md # Complete WebSocket message reference
│   ├── WEBSOCKET_MESSAGE_FRAMING.md # WebSocket message framing requirements
│   ├── WEBSOCKET_STATUS_PROTOCOL.md # WebSocket status protocol
│   ├── PIPELINE_STATUS_HANDLING.md # Pipeline status transitions
│   ├── PIPELINE_STAGE_INTERFACE.md # Pipeline stage interface specs
│   ├── PIPELINE_ARCHITECTURE.md # Pipeline architecture documentation
│   ├── BOM_HANDLING_GUIDE.md   # UTF-8 BOM handling guide
│   ├── LIQUIDITY_SCORING_SPECS.md # Liquidity scoring system specifications
│   └── STRATEGY_MODULE_SPECS.md # Trading strategy technical specifications
├── design/                      # Design documents
│   ├── ARCHITECTURE_PRINCIPLES.md # Core architectural principles
│   ├── REAL_TIME_UPDATES.md    # Real-time update system design
│   ├── LIQUIDITY_CALCULATION_APPROACH.md # Liquidity calculation methodology
│   └── STRATEGY_MODULE_DESIGN.md # Trading strategy module design
├── operations/                  # Operational documentation
│   └── TROUBLESHOOTING.md      # Common issues and solutions
└── reference/                   # Reference materials
    ├── CHANGELOG.md            # Version history
    ├── DATA_VALIDATION.md      # Data validation checklist
    └── FIXES_SUMMARY_2025_01_18.md # Historical fixes summary
```

## Documentation Guide

### For Users
Start with `user/README.md` for general usage, then refer to specific guides as needed.

### For Developers
- New contributors: Start with `developer/CONTRIBUTING.md`
- Architecture overview: `developer/ARCHITECTURE.md`
- Data formats: `specifications/DATA_SPECIFICATIONS.md`

### For System Administrators
- Deployment: `operations/DEPLOYMENT.md`
- Monitoring: `operations/MONITORING.md`

## Document Purposes

### User Documentation (`user/`)
**Purpose**: Help end-users install, configure, and use the application
- **README.md**: Quick start guide and feature overview
- **INSTALLATION.md**: Detailed installation steps for different platforms
- **LICENSE_ACTIVATION.md**: How to activate and manage licenses

### Developer Documentation (`developer/`)
**Purpose**: Enable developers to understand, build, and contribute to the project
- **ARCHITECTURE.md**: System design, component interactions, data flow
- **API_REFERENCE.md**: Detailed API endpoints, parameters, and responses
- **CONTRIBUTING.md**: Code style, PR process, testing requirements
- **BUILD_INSTRUCTIONS.md**: Development environment setup, build process
- **DEVELOPMENT_BEST_PRACTICES.md**: Comprehensive guide for documentation updates, testing requirements, and development standards

### Specifications (`specifications/`)
**Purpose**: Define exact technical requirements and formats
- **DATA_SPECIFICATIONS.md**: CSV/JSON formats, field definitions, data types
- **COLUMN_NAME_MAPPING.md**: How column names transform through the pipeline
- **WEBSOCKET_MESSAGE_SPECS.md**: Real-time messaging protocol specification
- **WEBSOCKET_MESSAGE_FRAMING.md**: Requirements for proper WebSocket message framing
- **PIPELINE_STATUS_HANDLING.md**: How pipeline stages transition and update in the UI
- **BOM_HANDLING_GUIDE.md**: Guidelines for handling UTF-8 BOM in CSV files
- **LIQUIDITY_SCORING_SPECS.md**: Technical specifications for the liquidity scoring system
- **STRATEGY_MODULE_SPECS.md**: Technical specifications for trading strategy module

### Design Documents (`design/`)
**Purpose**: Document system design decisions and future plans
- **REAL_TIME_UPDATES.md**: WebSocket implementation, file watching strategy
- **FUTURE_FEATURES.md**: Roadmap and planned enhancements
- **ARCHITECTURE_PRINCIPLES.md**: Three-layer architecture rules and guidelines
- **FRONTEND_ARCHITECTURE.md**: Frontend modularization and component system design
- **COMPONENT_NAMING_STANDARDS.md**: Standardized naming conventions for UI components
- **LIQUIDITY_CALCULATION_APPROACH.md**: Comprehensive methodology for hybrid liquidity scoring
- **STRATEGY_MODULE_DESIGN.md**: Architecture and design of the trading strategy subsystem

### Testing (`testing/`)
**Purpose**: Comprehensive testing documentation for all stakeholders
- **README.md**: Testing standards, conventions, and best practices
- **user-acceptance/**: End-user validation scenarios for new features
  - Includes step-by-step test cases for external users
  - Structured feedback collection forms
  - Version-specific test scenarios

### Operations (`operations/`)
**Purpose**: Guide deployment, maintenance, and troubleshooting
- **DEPLOYMENT.md**: Production deployment steps, configuration
- **MONITORING.md**: Health checks, logs, performance metrics
- **TROUBLESHOOTING.md**: Common problems and solutions

### Reference (`reference/`)
**Purpose**: Quick reference materials and historical information
- **CHANGELOG.md**: Version history, breaking changes, migration guides
- **DATA_VALIDATION.md**: Checklist for validating data integrity
- **GLOSSARY.md**: Definition of terms, acronyms, concepts

## Maintenance

- Keep documentation close to code - update docs with code changes
- Use relative links between documents
- Include examples where possible
- Date documents and note version compatibility
- Archive outdated documents rather than deleting