# Security Policy

## Supported Versions

Currently, only the latest version of SpecLedger receives security updates.

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly.

### How to Report

**Do NOT** open a public issue.

Instead, send an email to: security@specledger.io

Include:

* **Description**: A clear description of the vulnerability
* **Steps to Reproduce**: How to trigger the vulnerability
* **Impact**: Potential impact of the vulnerability
* **Proof of Concept**: If applicable, a proof of concept or exploit code

### What Happens Next

1. We will acknowledge receipt of your report within 48 hours
2. We will investigate the vulnerability
3. We will work on a fix and coordinate disclosure with you
4. Once fixed, we will release a security update
5. We will publicly disclose the vulnerability after the fix is released

### Security Best Practices

When using SpecLedger:

* **Keep updated**: Always use the latest version
* **Verify downloads**: Only download from official sources (GitHub Releases)
* **Check signatures**: Verify binary signatures when available
* **Review dependencies**: Regularly update dependencies with `sl deps update`
* **Secure cache**: Ensure `~/.specledger/cache` has appropriate permissions

### Dependency Vulnerabilities

SpecLedger relies on external Git repositories for specification dependencies.
To stay secure:

* Only add dependencies from trusted sources
* Pin to specific versions when possible
* Review dependency changes before resolving
* Keep your dependencies updated

### Private Data

SpecLedger does not collect or transmit any personal data. All operations are
performed locally on your machine.

The CLI caches specification files locally at `~/.specledger/cache` for
offline use. Ensure this directory has appropriate permissions if you're working
with sensitive specifications.
