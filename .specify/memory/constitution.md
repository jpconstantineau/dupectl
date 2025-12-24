# DupeCTL Constitution

<!--
SYNC IMPACT REPORT
==================
Version: 0.0.0 → 1.0.0
Rationale: Initial constitution ratification establishing foundational governance

Modified Principles: N/A (initial creation)
Added Sections:
  - All 12 core principles
  - Development Standards section
  - Operational Standards section
  - Governance section

Templates Requiring Updates:
  ✅ plan-template.md - Constitution Check section validated
  ✅ spec-template.md - Requirements alignment validated
  ✅ tasks-template.md - Task categorization aligned with principles

Follow-up TODOs: None
-->

## Core Principles

### I. Clean Code & Code Quality (NON-NEGOTIABLE)

All code MUST adhere to clean code principles:
- Functions MUST have a single, clear responsibility (Single Responsibility Principle)
- Names MUST be descriptive and reveal intent - no abbreviations except industry-standard (e.g., ID, HTTP, API)
- Functions MUST be small (<50 lines preferred, >100 lines requires justification)
- Code MUST be self-documenting - comments explain "why", not "what"
- Cyclomatic complexity MUST be kept low (≤10 per function, >15 requires refactoring)
- No magic numbers - use named constants with clear meaning
- Error handling MUST be explicit - no silent failures
- Code MUST pass Go linting (golangci-lint) with zero warnings

**Rationale**: Clean code reduces cognitive load, enables faster onboarding, reduces bugs, and improves long-term maintainability. Code is read 10x more than written.

### II. Clean Architecture & Separation of Concerns

The system MUST follow clean architecture principles:
- Clear separation between domain logic, application logic, and infrastructure
- Dependencies MUST point inward - domain has zero external dependencies
- Domain entities in `pkg/entities/` - pure business logic
- Application services in `pkg/` - orchestrate domain logic
- Infrastructure in `pkg/datastore/`, `pkg/api/`, `pkg/apiclient/` - external interfaces
- CLI commands in `cmd/` - thin adapter layer calling application services
- No circular dependencies between packages
- Interfaces define contracts at package boundaries
- Business logic MUST be testable without external dependencies (database, filesystem, network)

**Rationale**: Clean architecture ensures testability, flexibility, and long-term adaptability. It prevents tight coupling that makes changes expensive and risky.

### III. Test-First Development (NON-NEGOTIABLE)

Testing discipline MUST be strictly followed:
- **Test Coverage**: Minimum 70% code coverage for new code, 80% target for critical paths
- **Test Pyramid**: Unit tests (70%) > Integration tests (20%) > End-to-end tests (10%)
- **Test Types Required**:
  - Unit tests: Test individual functions/methods in isolation using mocks
  - Integration tests: Test package interactions and database operations
  - Contract tests: Verify API endpoint contracts match specifications
  - End-to-end tests: Test complete user workflows for critical features
- Tests MUST be written before or alongside implementation (TDD encouraged)
- All tests MUST pass before merge - zero tolerance for failing tests
- Tests MUST be fast - unit test suite completes in <30 seconds
- Tests MUST be deterministic - no flaky tests allowed
- Use table-driven tests for comprehensive scenario coverage (Go idiom)
- Mock external dependencies in unit tests - no network/database calls
- Use Go's standard testing package - avoid heavy testing frameworks unless justified

**Rationale**: Comprehensive testing prevents regressions, enables confident refactoring, serves as living documentation, and reduces production bugs by 70%+.

### IV. User Experience Consistency

CLI user experience MUST be consistent and predictable:
- All commands follow consistent verb-noun pattern (e.g., `get host`, `add policy`)
- Output formats MUST be consistent across commands (human-readable, JSON via --json flag)
- Error messages MUST be actionable - tell user what went wrong AND how to fix it
- Progress indication required for long-running operations (>2 seconds)
- Confirmation prompts for destructive operations (delete, apply policies)
- Help text (`--help`) MUST be comprehensive with examples
- Exit codes follow conventions: 0=success, 1=error, 2=usage error
- Color coding for clarity (green=success, red=error, yellow=warning) - but MUST work without colors
- Support both interactive and non-interactive modes for automation
- All dates/times in ISO 8601 format
- Human-readable output includes units (bytes → KB/MB/GB, durations in readable format)

**Rationale**: Consistent UX reduces user frustration, minimizes support burden, and enables efficient scripting/automation.

### V. Performance Requirements

The system MUST meet performance standards:
- File hash calculation: Process at least 100 MB/sec on modern hardware (SSD)
- Folder scans: Process at least 1000 files/sec for metadata collection
- API response time: <100ms p95 for GET requests, <500ms p95 for POST/PUT
- Database queries: Use indexes effectively, <50ms for common queries
- Memory usage: Stay under 500MB for typical workloads (<100k files)
- Scan resumption: Resume within 5 seconds of restart
- Graceful shutdown: Complete in <10 seconds, persist state before exit
- No memory leaks - long-running processes must have stable memory footprint
- Concurrent scanning of multiple folders using goroutines (leverage Go concurrency)
- Batch database operations - insert files in transactions of 1000 records
- Profile performance bottlenecks with pprof before optimization

**Rationale**: Performance directly impacts user satisfaction and system scalability. Users abandon slow tools. Defined thresholds enable objective measurement.

### VI. Minimal Dependencies & Standard Library First

Dependencies MUST be justified and minimal:
- **Prefer Go standard library** - it's comprehensive, stable, and well-tested
- External dependencies require justification in documentation
- Current approved dependencies:
  - `spf13/cobra` - CLI framework (industry standard, well-maintained)
  - `spf13/viper` - Configuration management (Cobra companion, configuration flexibility)
  - `modernc.org/sqlite` - Embedded SQLite database (pure Go, no CGo dependencies)
  - `golang-jwt/jwt` - JWT authentication (security-critical, peer-reviewed)
  - `google/uuid` - UUID generation (standard implementation)
- New dependencies MUST meet criteria:
  - Actively maintained (commit in last 6 months)
  - Well-documented with examples
  - Large user base (>1000 stars or enterprise adoption)
  - Compatible license (MIT, Apache 2.0, BSD)
  - Solves problem not adequately addressed by stdlib
- Avoid frameworks - prefer libraries for specific needs
- No external dependencies for core domain logic
- Vendor dependencies for build reproducibility

**Rationale**: Dependencies are liabilities - each adds security risks, maintenance burden, and potential breakage. Standard library is stable, performant, and has no external risk.

### VII. Maintainability & Code Longevity

Code MUST be maintainable for multi-year lifecycle:
- Package organization follows Go conventions: `pkg/` for libraries, `cmd/` for commands
- Public APIs (exported functions/types) MUST be documented with GoDoc comments
- Complex algorithms MUST include explanatory comments with big-O complexity
- Configuration via files, environment variables, or flags (12-factor compatibility)
- No hardcoded paths or system-specific assumptions
- Use `filepath` package for cross-platform path handling
- All database queries use parameterized statements (SQL injection prevention)
- Secrets never in code or version control - use environment variables or secure storage
- Regular dependency updates (quarterly security audits)
- Deprecation process: Mark deprecated, provide alternative, remove after 2 major versions
- Keep dependencies up-to-date within semantic versioning constraints

**Rationale**: Maintainability enables sustainable development velocity. Unmaintainable code accumulates technical debt that eventually paralyzes progress.

### VIII. Upgradability (User & Developer)

The system MUST support smooth upgrades:
- **User Upgradability**:
  - Single binary distribution - drop-in replacement
  - Automatic configuration migration on startup if schema changed
  - Database schema versioning with automatic migrations
  - Migration code tested with pre/post validation
  - Rollback capability if migration fails
  - Clear upgrade instructions in release notes
  - Backward compatible configuration (old configs continue working)
  - Breaking changes require major version bump and migration guide
- **Developer Upgradability**:
  - Go version specified in go.mod - developers use same version
  - Developer setup documented in README.md with exact steps
  - Build process reproducible across platforms
  - No undocumented environment requirements
  - Development dependencies managed by standard Go tooling

**Rationale**: Users must upgrade easily or they won't upgrade (security risk). Developers must onboard quickly or productivity suffers.

### IX. Observability for Users

Users MUST be able to understand system behavior:
- Structured logging to stderr (errors, warnings, info)
- Log levels configurable: ERROR, WARN, INFO, DEBUG
- Include context in logs: operation, file paths, error details, timestamps
- Progress bars for long-running operations with ETA when possible
- Status commands: `get agent` shows scan status, `get duplicates` shows analysis results
- Verbose mode (`--verbose` flag) for detailed operation logging
- Dry-run mode for destructive operations (`--dry-run` shows what would happen)
- Health check endpoint for API server
- Metrics exposed: files scanned, duplicates found, scan duration, errors encountered
- Debug logging includes stack traces on panic recovery
- Operation audit trail in database for critical operations (policy applications, deletions)

**Rationale**: Observability builds user trust and enables self-service troubleshooting. Users can diagnose issues without developer intervention.

### X. Cross-Platform Portability

Code MUST run on Windows, Linux, and macOS without modification:
- Use `filepath.Join()` and `filepath.Separator` for paths - never hardcode `/` or `\`
- Use `os.PathSeparator`, `os.PathListSeparator` for system-specific separators
- File permissions handled with `os.FileMode` - respect platform differences
- Windows vs Unix user/group ID handling - store as strings, handle absence gracefully
- Timestamps use `time.Time` with UTC normalization
- Case sensitivity - assume case-insensitive filesystems exist (Windows, macOS)
- Line endings normalized (`\r\n` vs `\n`) in text processing
- Test on all three platforms before release
- Build for all platforms: `GOOS=windows`, `GOOS=linux`, `GOOS=darwin`
- CI/CD tests on Linux (primary), validation on Windows and macOS
- No platform-specific build tags unless functionality truly differs

**Rationale**: Users run diverse platforms. Platform-specific code fragments user base and increases maintenance. Go's cross-platform stdlib makes portability straightforward.

### XI. Graceful Shutdown & Interruption Handling

The system MUST handle interruption gracefully:
- Catch SIGINT (Ctrl+C), SIGTERM signals using `signal.Notify()`
- On interrupt during scan: Save current progress to database immediately
- Persist scan state: last processed folder, last processed file, files processed count
- Graceful shutdown sequence:
  1. Stop accepting new work
  2. Complete current file hash calculation
  3. Commit database transaction
  4. Write checkpoint with resume information
  5. Close database connections cleanly
  6. Exit with status code 130 (interrupted)
- On startup: Check for incomplete scans, offer to resume
- Resume capability: Continue from last checkpoint without re-scanning
- Progress visible on resume: "Resuming scan at file 5,432 of 10,000"
- Timeout protection: Force exit after 10 seconds if graceful shutdown stalls
- No corrupted data after hard kill (database transactions provide atomicity)
- Long operations use context with cancellation for clean termination

**Rationale**: User may need to interrupt long scans. Data loss or corruption is unacceptable. Resume capability respects user time.

### XII. Twelve-Factor CLI Application Principles

Adapt 12-factor app principles for CLI application:

1. **Codebase**: Single codebase in version control, multiple deployments (user machines)
2. **Dependencies**: Explicitly declared in go.mod, vendored for reproducibility
3. **Config**: All configuration via environment variables or config file, no hardcoded values
4. **Backing Services**: Database (SQLite) treated as attached resource, path configurable
5. **Build/Release/Run**: Strict separation - `go build` (build), GitHub releases (release), user execution (run)
6. **Processes**: CLI commands are stateless processes, state in database only
7. **Port Binding**: Server mode binds to configured port, self-contained HTTP server
8. **Concurrency**: Scale via goroutines (multiple folder scans), not multiple processes
9. **Disposability**: Fast startup (<1s), graceful shutdown (<10s), robust against crashes
10. **Dev/Prod Parity**: Same Go binary runs in dev and prod, same database engine (SQLite)
11. **Logs**: Treat logs as event streams to stderr, never manage log files
12. **Admin Processes**: Maintenance tasks as regular commands (e.g., `dupectl admin migrate`)

**Rationale**: 12-factor principles provide battle-tested patterns for robust, maintainable applications. CLI adaptation ensures consistency with modern software practices.

## Development Standards

### Code Review Requirements

- All changes require code review before merge (no direct commits to main)
- Reviewer checklist:
  - Constitution compliance (all principles followed)
  - Test coverage adequate
  - Error handling comprehensive
  - Documentation updated
  - No new linter warnings
  - Performance implications considered
  - Security implications reviewed
- PR description includes: motivation, approach, testing performed
- Breaking changes highlighted and documented

### Version Control Practices

- Semantic versioning: MAJOR.MINOR.PATCH
  - MAJOR: Breaking API changes, incompatible database schema
  - MINOR: New features, backward compatible additions
  - PATCH: Bug fixes, performance improvements, documentation
- Conventional commits: `type(scope): description`
  - Types: feat, fix, docs, refactor, test, chore, perf
  - Breaking changes marked with `BREAKING CHANGE:` in footer
- Feature branches: `###-feature-name` pattern
- Main branch always stable and deployable
- Git tags for releases: `v1.2.3`

### Documentation Requirements

- README.md: Project overview, quick start, installation, basic usage
- Each package has package-level documentation (doc.go or package comment)
- Complex functions have usage examples in GoDoc
- Design documents in `docs/design/` for architecture decisions
- API documentation generated from code (OpenAPI/Swagger for HTTP API)
- User guide for common workflows
- Migration guides for breaking changes
- Changelog maintained (release notes)

## Operational Standards

### Build & Release Automation

- CI/CD pipeline automates entire release process
- Build process:
  1. Run linters (golangci-lint)
  2. Run tests with coverage (>70% gate)
  3. Build binaries for Windows, Linux, macOS (amd64, arm64)
  4. Generate checksums (SHA256)
  5. Create GitHub release with binaries and release notes
- Release triggers: Git tags matching `v*.*.*`
- Automated checks: no manual release steps
- Reproducible builds: Same source → same binary
- Code signing for releases (future requirement)

### Security Practices

- No secrets in code or configuration files
- Database credentials from environment variables
- JWT tokens for API authentication
- HTTPS for API server (TLS required in production)
- Input validation on all user-provided data
- SQL injection prevention via parameterized queries
- Path traversal prevention in file operations
- Regular security audits of dependencies (`go list -m all | nancy sleuth`)
- CVE monitoring for Go runtime and dependencies
- Principle of least privilege - run as non-root user

### Backward Compatibility

- Public APIs (CLI commands, flags, configuration, database schema) MUST remain backward compatible within major version
- Deprecation policy:
  1. Mark feature as deprecated in release N with warning message
  2. Provide alternative/migration path in documentation
  3. Remove deprecated feature in release N+2 (two major versions later)
- Configuration files: New fields added, old fields never removed (deprecated fields ignored with warning)
- Database migrations: Always forward-compatible (new version reads old schema, migrates automatically)
- Breaking changes require major version bump and explicit migration guide
- Test backward compatibility: Run new binary against old database/config in CI

## Governance

This constitution is the supreme authority for all development decisions. It supersedes individual preferences and ad-hoc practices.

### Amendment Process

1. Propose amendment via GitHub issue with justification
2. Discuss with maintainers and community
3. Document amendment impact (affected code, templates, processes)
4. Update constitution version:
   - MAJOR bump: Principle removal or fundamental redefinition
   - MINOR bump: New principle added or substantial expansion
   - PATCH bump: Clarifications, wording improvements
5. Update last amended date to amendment approval date
6. Propagate changes to all templates and guidance documents
7. Commit with message: `docs: amend constitution to vX.Y.Z (summary)`

### Compliance

- Pull request reviews MUST verify constitution compliance
- Complexity or principle violations require explicit justification in PR
- Justifications documented in code comments or design docs
- Technical debt tracked: Violations logged as issues for remediation
- Quarterly constitution review: Ensure principles remain relevant and achievable

### Configuration Management

- Constitution stored at `.specify/memory/constitution.md`
- Templates reference constitution principles (cross-check on changes)
- Runtime guidance for developers in README.md and `docs/`
- Principle violations tracked as technical debt items

**Version**: 1.0.0 | **Ratified**: 2025-12-23 | **Last Amended**: 2025-12-23
