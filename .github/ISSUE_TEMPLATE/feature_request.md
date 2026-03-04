---
name: Feature request
about: Suggest an idea for this project
title: '[FEATURE] '
labels: enhancement
assignees: ''

---

**Is your feature request related to a problem? Please describe.**
A clear and concise description of what the problem is. Ex. I'm always frustrated when [...]

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions or features you've considered.

**Proposed Command Interface**
If this feature involves new CLI commands or flags, please describe the proposed interface:

```bash
# Example command
usm <new-command> --flag value
```

**API Requirements**
If this feature requires new API endpoints, please describe:
- Required API method(s)
- Expected request/response format
- Authentication requirements

**Controller Compatibility**
Please indicate which controller types this feature should support:
- [ ] UniFi Cloud Controller
- [ ] UniFi Local/On-Premises Controller
- [ ] Both

**Additional context**
Add any other context or screenshots about the feature request here.

**Implementation Notes**
If you have ideas about how to implement this feature, please share them here. Remember that new features must:
1. Be added to the `SiteManager` interface
2. Be implemented in `CloudClient`
3. Be stubbed in `LocalClient` (if not supported locally)
4. Include router tests in `internal/pkg/cli/router_test.go`
