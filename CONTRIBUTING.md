# Contributing to Sublime Admin

Thank you for your interest in contributing to Sublime Admin! This document provides guidelines for contributors.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git

### Setting Up

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/sublime-admin.git
   cd sublime-admin
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run tests:
   ```bash
   go test ./...
   ```

## Development Workflow

### Making Changes

1. Create a new branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following Go conventions

3. Run tests:
   ```bash
   go test ./... -count=1
   ```

4. Commit with a clear message:
   ```bash
   git commit -m "feat: add new feature"
   ```

### Commit Message Format

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `refactor:` - Code refactoring
- `test:` - Tests
- `chore:` - Maintenance

### Pull Request Process

1. Push to your fork
2. Open a PR against `main`
3. Wait for review

## Code Style

- Follow `gofmt` formatting
- Use meaningful names
- Add comments for exported items
- Handle errors explicitly

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
