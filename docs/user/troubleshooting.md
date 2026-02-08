# Troubleshooting

Common issues and solutions when using SpecLedger.

## "Not a SpecLedger project" Error

**Cause**: Running a spec-specific command outside a project directory.

**Solutions**:
```bash
# Navigate to a project directory
cd ~/demos/myproject

# Or create a new project first
sl new --ci --project-name myproject --short-code mp
```

## Permission Denied Errors

**Cause**: Insufficient permissions for installation directory.

**Solutions**:
```bash
# Run with sudo
sudo sl new --ci --project-name myproject --short-code mp

# Or install to user directory
mkdir -p ~/.local/bin
curl -fsSL https://raw.githubusercontent.com/specledger/specledger/main/scripts/install.sh | bash
export PATH=$PATH:~/.local/bin
```

## TUI Not Working in CI/CD

**Cause**: TUI requires interactive terminal.

**Solution**: Always use `--ci` flag in non-interactive environments:
```bash
sl new --ci --project-name myproject --short-code mp
```

## Framework Not Found

**Cause**: Framework wasn't installed during project creation.

**Solutions**:
```bash
# Check tool status
sl doctor

# Install manually via mise
mise install pipx:git+https://github.com/github/spec-kit.git

# Initialize manually
specify init --here --ai claude --force --script sh --no-git
```

## Dependency Resolution Fails

**Cause**: Network issue or invalid repository URL.

**Solutions**:
```bash
# Check network connectivity
git ls-remote git@github.com:org/repo.git HEAD

# Verify repository exists and is accessible
# Try resolving again
sl deps resolve
```

## File: `specledger.yaml` Not Found

**Cause**: Not in a SpecLedger project directory.

**Solution**:
```bash
# Initialize current directory
sl init

# Or navigate to project root
cd ~/demos/myproject
```

## Tool Not Installed Errors

**Cause**: Required tool (mise, bd, perles) not installed.

**Solution**:
```bash
# Install mise first
curl https://mise.run | sh

# Then use mise to install other tools
mise install bd
mise install perles
```

## Getting More Help

- **Documentation**: [https://specledger.io/docs](https://specledger.io/docs)
- **GitHub Issues**: [https://github.com/specledger/specledger/issues](https://github.com/specledger/specledger/issues)
- **GitHub Discussions**: [https://github.com/specledger/specledger/discussions](https://github.com/specledger/specledger/discussions)
