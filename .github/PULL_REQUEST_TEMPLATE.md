## Summary
<!-- Brief description of the changes in this PR -->

## Type of Change
<!-- Mark the relevant option with an "x" -->
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Refactoring (no functional changes)
- [ ] Documentation update
- [ ] CI/CD changes
- [ ] Dependencies update

## Related Issue
<!-- Link to related issues: Fixes #123, Closes #456 -->

## Testing
<!-- Describe the tests you ran and how to reproduce -->
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Code Review Checklist
<!-- These items are from CLAUDE.md coding standards -->
- [ ] All errors are handled properly
- [ ] Context is passed to all I/O operations
- [ ] No goroutine leaks (all goroutines can terminate)
- [ ] Race conditions checked (`go test -race`)
- [ ] Sensitive data is not logged
- [ ] Input validation is present
- [ ] Tests are included for new functionality
- [ ] No hardcoded credentials or secrets
- [ ] Database queries use parameterized inputs
- [ ] Resources are properly closed (defer)

## Breaking Changes
<!-- If this is a breaking change, describe the impact and migration path -->
N/A

## Screenshots (if applicable)
<!-- Add screenshots for UI changes -->

## Additional Notes
<!-- Any additional information reviewers should know -->
