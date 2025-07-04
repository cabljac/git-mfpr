# Contributing to git-mfpr

First off, thank you for considering contributing to git-mfpr! It's people like you that make git-mfpr such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible. Fill out the required template, the information it asks for helps us resolve issues faster.

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

- A clear and descriptive title
- A detailed description of the proposed enhancement
- Explain why this enhancement would be useful
- List any alternative solutions or features you've considered

### Pull Requests

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Development Process

1. **Setup Development Environment**
   ```bash
   git clone https://github.com/yourusername/git-mfpr.git
   cd git-mfpr
   go mod download
   ```

2. **Make Your Changes**
   - Write clean, readable Go code
   - Follow Go best practices and idioms
   - Add tests for new functionality
   - Update documentation as needed

3. **Run Tests**
   ```bash
   make test
   ```

4. **Run Linters**
   ```bash
   make lint
   ```

5. **Build and Test Locally**
   ```bash
   make build
   ./bin/git-mfpr --help
   ```

## Coding Standards

- Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Write meaningful commit messages following [Conventional Commits](https://www.conventionalcommits.org/)
- Keep functions small and focused
- Write descriptive variable and function names
- Add comments for complex logic
- Ensure all exported functions have documentation

## Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat:` - A new feature
- `fix:` - A bug fix
- `docs:` - Documentation only changes
- `style:` - Changes that don't affect the code meaning
- `refactor:` - Code change that neither fixes a bug nor adds a feature
- `perf:` - Code change that improves performance
- `test:` - Adding missing tests or correcting existing tests
- `build:` - Changes that affect the build system or external dependencies
- `ci:` - Changes to CI configuration files and scripts
- `chore:` - Other changes that don't modify src or test files

Example:
```
feat: add support for custom branch patterns

- Allow users to specify custom branch naming patterns
- Add validation for branch names
- Update documentation with examples
```

## Testing

- Write unit tests for all new functionality
- Aim for at least 80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Test both success and failure cases

Example test structure:
```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Documentation

- Update the README.md if you change functionality
- Add inline documentation for exported functions
- Include examples in documentation where helpful
- Keep documentation concise but comprehensive

## Questions?

Feel free to open an issue with your question or reach out to the maintainers directly.

Thank you for contributing! 🎉