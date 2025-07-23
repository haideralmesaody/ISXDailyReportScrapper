# Versioning & Release Management Guidelines

## 1. Overview

ISX Daily Reports Scraper follows Semantic Versioning 2.0.0 (SemVer) plus structured pre-release phases (alpha → beta → rc → stable) to deliver predictable upgrades and clear communication. ([semver.org](https://semver.org))

## 2. Public API Definition

For versioning decisions the public API consists of:
- Documented REST/HTTP endpoints (request/response & error schemas)
- WebSocket message payload schema
- CLI command/flags & exit codes
- Exported Go packages intended for external use
- Persisted data formats (database schema & export file formats)

Only changes to this surface determine whether a release is breaking.

## 3. Version Format

A normal version: `MAJOR.MINOR.PATCH` with optional prerelease (`-alpha.N`, `-beta.N`, `-rc.N`) and optional build metadata (`+dev.<date>`, `+ci.<run>`, `+commit.<sha>`); build metadata never affects precedence ordering.

## 4. Pre-Release Phases

- **Alpha** (0.1.0-alpha.1 ... 0.4.x): Rapid iteration; MINOR may introduce breaking changes; PATCH for fixes and small enhancements; every alpha build carries `-alpha.N`.
- **Beta** (0.5.0-beta.1 ... 0.9.x): Feature-complete; focus on stabilization, performance, and migration hardening; only necessary breaking changes with documented scripts.
- **Release Candidate** (x.y.z-rc.N): Feature freeze; only critical bug, security, or performance fixes; exit criteria defined (see §17).
- **Stable** (≥1.0.0): Strict SemVer—no breaking changes outside MAJOR; deprecations announced before removal.

## 5. Version Increment Policy

| Situation | < 1.0 (Alpha/Beta) | ≥ 1.0 (Stable) | Rationale |
|-----------|-------------------|----------------|-----------|
| Breaking API / data change | MINOR | MAJOR | Predictability pre-1.0 while acknowledging SemVer looseness for 0.y.z |
| Backward-compatible feature | PATCH (alpha) / MINOR (beta) | MINOR | Control scope inflation pre-RC |
| Bug / security / perf fix | PATCH | PATCH | Rapid delivery of corrections |
| Deprecation introduction | PATCH | MINOR | Advertise before removal |

A release script validates commits and enforces the correct bump.

## 6. Commit Message Convention

Use Conventional Commits: `<type>(scope?): <description>` with allowed types:
- `feat`: New feature
- `fix`: Bug fix
- `perf`: Performance improvement
- `refactor`: Code restructuring
- `docs`: Documentation
- `build`: Build system changes
- `ci`: CI configuration
- `test`: Test additions/changes
- `chore`: Maintenance tasks

`BREAKING CHANGE:` footer (or `!` after type) flags a breaking change and drives MAJOR (or MINOR <1.0).

## 7. Automated Version Derivation

CI parses commit history since last tag:
- Any breaking indicator → MAJOR (or MINOR <1.0)
- Presence of `feat` without breaking → MINOR (or PATCH in alpha if trivial & explicitly labeled)
- Otherwise PATCH when `fix|perf` present
- If no qualifying commits, no release is produced

## 8. Changelog Standard

Maintain `CHANGELOG.md` in reverse chronological order, one entry per version, ISO-8601 dates, grouping changes under:
- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Vulnerability fixes
- **Performance**: Performance improvements
- **Migration**: Required migration steps

Clearly state SemVer compliance. ([keepachangelog.com](https://keepachangelog.com))

## 9. Release Notes Template

Each release entry contains (when applicable):
- **Overview**: 1–2 sentence impact summary
- **Added/Changed/Deprecated/Removed/Fixed/Security/Performance**
- **Migration**: Step-by-step migration guide
- **Known Issues**: Outstanding problems
- **Upgrade Risks**: Potential issues during upgrade

## 10. Tagging Strategy

- All public releases use annotated tags: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
- Sign tags where possible
- Lightweight tags reserved for temporary internal experiments
- Tags are immutable—corrections require a new version

## 11. Branch & Flow Model

- `main` remains releasable
- Feature branches merge via PR with required tests & lint
- RC or stable hotfixes may branch from `release/X.Y`
- Security fixes branch from latest supported release then merge forward to prevent regressions

## 12. Deprecation Policy

- Deprecations announced in release notes under **Deprecated** with replacement guidance
- Remain available ≥1 MINOR (or until next MAJOR after 1.0)
- Runtime warnings optional for high-impact items
- Removal recorded under **Removed** with migration steps

## 13. Data & Migration Policy

- Each release declares a Data Format Version
- Migrations are idempotent, tested forward (old→new) and verified in CI
- From first beta onward, manual data resets are avoided—scripted migrations provided
- Rollback strategy documented for significant schema changes

## 14. Compatibility Matrix

| App Versions | Data Format | API Version | Notes |
|--------------|-------------|-------------|-------|
| 0.1.x–0.2.x | v1 | n/a or v1-alpha | Early experimentation; possible data resets |
| 0.5.x (beta) | v2 | v1-beta | Stable migrations promised |
| 1.0.x | v2 | v1 | Backward compatible within series |

Matrix updated whenever API or data format changes.

## 15. Build Metadata

Use build metadata only for non-tagged artifacts:
- `+dev.YYYYMMDD`: Development builds
- `+ci.<run>`: CI build number
- `+commit.<sha>`: Git commit hash

Avoid relying on metadata for dependency constraints since SemVer ignores it for ordering.

## 16. Release Automation Pipeline

Pipeline stages:
1. Lint
2. Test
3. Security scan
4. Build
5. Semantic version calculation (dry run)
6. Changelog generation
7. Tag & publish
8. Artifact attach (binaries, images)
9. Post-release notification

Tagging occurs only after successful artifact build.

## 17. RC Exit Criteria

Promote `-rc.N` to stable only when:
- Zero open P1 defects
- Performance benchmarks met
- Documentation & migration scripts complete
- Upgrade/downgrade smoke tests pass (including data migration)
- No unresolved security advisories

## 18. Security & Hotfix Releases

- Security issues patched via dedicated branch
- Publish PATCH release with clear **Security** section describing impact & mitigation
- Forward-merge to all active branches to prevent divergence

## 19. Release Cadence Targets

- **Alpha**: Weekly (≤10 working days between tags) to keep feedback loop tight
- **Beta**: Bi-weekly to monthly
- **Stable PATCH**: As needed (aim <14 days for critical fixes)
- **Stable MINOR**: Roughly quarterly
- **MAJOR**: When substantial breaking changes accumulate, not calendar-driven

## 20. Quality Gates

A release must pass:
- Automated tests (≥ defined coverage)
- Static analysis
- Performance baseline (no regressions beyond threshold)
- Schema diff (no unintended breaking change)
- Changelog completeness validation script

## 21. Commit & Changelog Tooling

Adopt commit linting & automated changelog generation leveraging Conventional Commits to reduce manual errors and enable semantic release tooling; changelog entries remain human-optimized per Keep a Changelog guidance.

### Recommended Tools:
- **commitlint**: Enforce commit conventions
- **semantic-release**: Automated version management
- **standard-version**: Changelog generation
- **git-cliff**: Rust-based changelog generator

## 22. Documentation Requirements per Release

Each tagged release updates:
- README version badge
- API docs (if surface changed)
- Migration guide (if needed)
- Release notes

Failing docs update blocks tagging in CI.

## 23. FAQ

**Q: When is 1.0.0 released?**
A: After API freeze, data format stability, RC criteria satisfied, and no critical defects outstanding.

**Q: How are unexpected breaking changes handled pre-1.0?**
A: Increment MINOR with explicit migration steps.

**Q: What if a fix materially changes behavior?**
A: Treat as feature (MINOR) unless acceptable as a PATCH by SemVer rules and risk assessment.

## 24. References

- [Semantic Versioning 2.0.0](https://semver.org)
- [Keep a Changelog](https://keepachangelog.com)
- [Conventional Commits](https://conventionalcommits.org)
- [Git Cliff](https://git-cliff.org)
- [Semantic Release](https://semantic-release.gitbook.io)

---

## ISX Project Specific Notes

### Current Version
- **Version**: 0.1.0-alpha.1
- **Stage**: Alpha development
- **Data Format**: v1
- **API Version**: v1-alpha (WebSocket only)

### Version File Locations
- `VERSION`: Root file containing current version
- `CHANGELOG.md`: In `docs/reference/`
- Version constants in code: `dev/internal/common/version.go` (to be created)

### Release Checklist
- [ ] Run tests: `go test ./...`
- [ ] Update VERSION file
- [ ] Update CHANGELOG.md
- [ ] Update version in code
- [ ] Build binaries: `.\build.bat`
- [ ] Create git tag
- [ ] Create GitHub release
- [ ] Attach binaries to release