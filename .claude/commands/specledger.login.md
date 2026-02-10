---
description: Authenticate with Specledger using the CLI.
---

## Execution

Khi user gọi `/specledger.login`, chạy lệnh CLI:

```bash
sl auth login
```

CLI sẽ tự động:
1. Mở browser đến trang đăng nhập
2. Khởi động callback server để nhận token
3. Lưu credentials vào `~/.specledger/credentials.json`

**Nếu `sl` chưa được cài đặt**, chạy:
```bash
go run ./cmd/sl auth login
```

### Xác nhận thành công

Sau khi CLI hoàn tất, verify credentials:
```bash
sl auth status
```

Hoặc:
```bash
cat ~/.specledger/credentials.json
```
