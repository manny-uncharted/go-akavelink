# Contributing Guide

Thanks for helping improve `go-akavelink`! Here's how to get started:

---

## 🛠 Local Dev Setup

```bash
git clone https://github.com/akave-ai/go-akavelink
cd go-akavelink
go run ./cmd/server
```

Visit `http://localhost:8080/health` to test the server.

---

## 📋 Branch Naming Convention

Use descriptive branch prefixes:

- `feat/` for new features
- `fix/` for bug fixes
- `chore/` for cleanup, tooling, or non-feature code
- `docs/` for documentation changes

---

## ✅ Pull Request Checklist

- [ ] PR title is clear and follows prefix convention
- [ ] Code compiles and passes basic tests
- [ ] Linter and formatter have been run (use `go fmt`, `golangci-lint`)
- [ ] Any new functionality is covered by tests (if applicable)
- [ ] Relevant issue is linked or referenced

---

## 💡 Tips

- Be kind in reviews and helpful in issues.
- If something is unclear in the SDK or architecture, open a `question` issue — chances are others are wondering the same.
- Contributions of all kinds are welcome — code, docs, tests, discussions.

---

Feel free to pick a [`good first issue`](https://github.com/akave-ai/go-akavelink/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)!
