# Contributing Guide

## Git Workflow (Simplified GitFlow)

```
master (stable)
    │
    ├── feature/*   (new functionality)
    ├── bugfix/*    (fixes)
    ├── release/*   (version prep)
    └── hotfix/*    (urgent fixes)
```

## Branch Naming

| Type | Pattern | Example |
|------|---------|---------|
| Feature | `feature/<name>` | `feature/cli-framework` |
| Bugfix | `bugfix/<name>` | `bugfix/executor-timeout` |
| Release | `release/v<ver>` | `release/v0.1.0` |
| Hotfix | `hotfix/<name>` | `hotfix/critical-fix` |

## Basic Workflow

### Start New Work

```bash
git checkout master
git checkout -b feature/my-feature
```

### Commit Changes

```bash
git add .
git commit -m "feat(scope): description"
```

### Merge to Master

```bash
git checkout master
git merge --no-ff feature/my-feature
git branch -d feature/my-feature
```

## Commit Message Format

```
<type>(<scope>): <description>
```

**Types:** `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

**Scopes:** `cli`, `executor`, `config`, `output`, `bank`, `keys`, `gov`, `staking`, etc.

**Examples:**
```
feat(cli): implement command parser
fix(executor): handle timeout error
docs(readme): update usage examples
test(bank): add send command tests
```

## Development Phases

```bash
# Phase 1: Foundation
feature/cli-framework
feature/executor
feature/config
feature/output
feature/types

# Phase 2: Core Modules
feature/status-module
feature/keys-module
feature/bank-module

# Phase 3: Governance
feature/gov-module

# Phase 4: Staking
feature/staking-module
feature/multistaking-module

# Phase 5+: Additional Modules
feature/tokens-module
feature/spending-module
...
```

## Before Merging

```bash
make test      # All tests pass
make fmt       # Code formatted
make lint      # No lint errors
```
