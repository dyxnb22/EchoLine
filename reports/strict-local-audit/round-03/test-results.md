# Round 03 Test Results

| Command | Result |
|---------|--------|
| `make test` | PASS |
| `go test ./...` | PASS |
| `go vet ./...` | PASS |
| `npm run lint` / `npm run build` | PASS |
| `make smoke-full` | PASS (16/16) |
| `RUN_INTEGRATION=1 go test -run Integration ./tests/...` | PASS |
| `npx playwright test` | PASS (4/4) |
