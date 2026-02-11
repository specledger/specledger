---
description: Authenticate with Specledger using the CLI.
---

## Execution

When user calls `/specledger.login`, run the CLI command:

```bash
sl auth login
```

The CLI will automatically:
1. Open browser to the login page
2. Start callback server to receive token
3. Save credentials to `~/.specledger/credentials.json`

**If `sl` is not installed**, run:
```bash
go run ./cmd/sl auth login
```

### Verify success

After CLI completes, verify credentials:
```bash
sl auth status
```

Or:
```bash
cat ~/.specledger/credentials.json
```
