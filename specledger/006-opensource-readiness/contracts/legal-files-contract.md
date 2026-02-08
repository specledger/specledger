# Legal Files Contract

**Feature**: 006-opensource-readiness | **Version**: 1.0

## Overview

This contract defines the required legal files for open source compliance and their expected content.

## Required Files

### LICENSE (MUST EXIST)

**Location**: Repository root (`/LICENSE`)
**Format**: Plain text
**Required Content**:
- License identifier (MIT)
- Copyright notice
- Permission grant
- Conditions and limitations

**Validation**:
```bash
# File must exist
test -f /LICENSE

# Must contain "MIT License"
grep -q "MIT License" /LICENSE

# Must contain copyright notice
grep -q "Copyright" /LICENSE
```

### NOTICE (MUST EXIST)

**Location**: Repository root (`/NOTICE`)
**Format**: Plain text
**Required Content**:
- Project name and copyright
- Third-party attributions
- License references

**Validation**:
```bash
# File must exist
test -f /NOTICE

# Must mention third-party software
grep -q "third-party" /NOTICE
```

### CODE_OF_CONDUCT.md (MUST EXIST)

**Location**: Repository root
**Format**: Markdown
**Required Content**:
- Pledge for inclusive community
- Standards for behavior
- Reporting instructions

### SECURITY.md (MUST EXIST)

**Location**: Repository root
**Format**: Markdown
**Required Content**:
- Supported versions policy
- Reporting vulnerability process
- Security update announcements

### CONTRIBUTING.md (MUST EXIST)

**Location**: Repository root
**Format**: Markdown
**Required Content**:
- Pull request process
- Code standards
- Contributor license agreement

### GOVERNANCE.md (MUST EXIST)

**Location**: Repository root
**Format**: Markdown
**Required Content**:
- Maintainer list
- Decision-making process
- Release process
- Security policy reference

## Contract Compliance

### Pre-merge Validation

Before any merge to main branch:

```yaml
check_legal_files:
  - exists: LICENSE
  - exists: NOTICE
  - exists: CODE_OF_CONDUCT.md
  - exists: SECURITY.md
  - exists: CONTRIBUTING.md
  - exists: GOVERNANCE.md
  - contains:
      file: LICENSE
      text: "MIT License"
  - contains:
      file: NOTICE
      text: "SpecLedger"
```

### Automated Testing

```bash
#!/bin/bash
# legal-files-check.sh

ERRORS=0

check_file() {
  if [ ! -f "$1" ]; then
    echo "ERROR: Missing $1"
    ERRORS=$((ERRORS + 1))
  fi
}

check_file "LICENSE"
check_file "NOTICE"
check_file "CODE_OF_CONDUCT.md"
check_file "SECURITY.md"
check_file "CONTRIBUTING.md"
check_file "GOVERNANCE.md"

if [ $ERRORS -gt 0 ]; then
  echo "Found $ERRORS missing legal files"
  exit 1
fi

echo "All legal files present"
exit 0
```

## Source File Headers

All Go source files MUST include the following header:

```go
// Copyright (c) 2025 SpecLedger Contributors
// SPDX-License-Identifier: MIT
//
// See LICENSE file in the project root for full license information.
```

### Validation Script

```bash
#!/bin/bash
# check-headers.sh

for file in $(find . -name "*.go" ! -path "./vendor/*"); do
  if ! head -n 5 "$file" | grep -q "SPDX-License-Identifier: MIT"; then
    echo "Missing license header: $file"
    exit 1
  fi
done

echo "All source files have license headers"
exit 0
```

## Acceptance Criteria

1. All 6 legal files exist in repository root
2. LICENSE file contains MIT license text
3. NOTICE file lists third-party dependencies
4. All Go source files have license headers
5. Legal files pass CI validation
6. Legal files are referenced in README
