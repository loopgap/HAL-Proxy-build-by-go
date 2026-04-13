# Contributing to HAL-Proxy

Thank you for your interest in contributing to HAL-Proxy! This document provides guidelines and instructions for contributing.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)

## 📜 Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## 🚀 Getting Started

### Prerequisites

- Go 1.26 or later
- Node.js 20 or later
- Git
- Docker (optional)

### Fork and Clone

```bash
# Fork the repository on GitHub

# Clone your fork
git clone https://github.com/YOUR_USERNAME/hal-proxy.git
cd hal-proxy

# Add upstream remote
git remote add upstream https://github.com/your-org/hal-proxy.git
```

### Install Dependencies

```bash
# Go dependencies
go mod download

# Frontend dependencies
cd ui && npm install && cd ..
```

## 🔄 Development Workflow

### 1. Create a Branch

```bash
# Update your local main branch
git checkout main
git pull upstream main

# Create a new feature branch
git checkout -b feature/your-feature-name
```

### 2. Make Changes

Make your changes following our coding standards.

### 3. Test Your Changes

```bash
# Run all tests
make test

# Run Go tests
go test ./...

# Run frontend tests
cd ui && npm test && cd ..
```

### 4. Commit Your Changes

```bash
# Stage changes
git add .

# Commit with a clear message
git commit -m "Add feature: your feature description"

# Follow conventional commit format
# feat: Add new feature
# fix: Fix a bug
# docs: Documentation changes
# style: Code style changes
# refactor: Code refactoring
# test: Test changes
# chore: Maintenance tasks
```

### 5. Push and Create PR

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create Pull Request on GitHub
```

## 📏 Coding Standards

### Go

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Run `go fmt` before committing
- Run `golangci-lint run` to check for issues
- Add comments for exported functions and types
- Keep functions small and focused

### React/TypeScript

- Follow ESLint and Prettier configurations
- Use functional components with hooks
- Prefer TypeScript types over PropTypes
- Write self-documenting code with clear naming

### Git

- Write clear, concise commit messages
- Keep commits focused and atomic
- Reference issues in commit messages (e.g., "Fix #123")

## 🧪 Testing

### Go Testing

```bash
# Run all tests with race detection
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestFunctionName ./...
```

### Frontend Testing

```bash
# Run all tests
cd ui && npm test

# Run tests in watch mode
cd ui && npm test -- --watch

# Run with coverage
cd ui && npm run test:coverage
```

### E2E Testing

```bash
# Run E2E tests
npm run test:e2e
```

## 📤 Submitting Changes

### Pull Request Checklist

- [ ] Code follows our coding standards
- [ ] Tests pass locally
- [ ] Documentation updated (if needed)
- [ ] Commits are atomic and well-described
- [ ] No merge conflicts with main branch

### After Submitting

- Respond to review feedback promptly
- Keep PR focused and avoid scope creep
- Update PR if issues are found

## 📝 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://reactjs.org/docs/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)

## ❓ Questions?

Feel free to:

- Open an issue for bugs or feature requests
- Join our community discussions
- Contact the maintainers

Thank you for contributing to HAL-Proxy! 🎉
