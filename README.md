# GraphQL Inspector

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/bishnuag/graphql-inspector)

A comprehensive GraphQL schema management and evolution tool written in Go. Compare schemas, validate documents, find breaking changes, analyze coverage, and more.

## 🚀 Features

- **Schema Comparison**: Compare two GraphQL schemas and detect breaking, dangerous, and non-breaking changes
- **Document Validation**: Validate GraphQL documents against schemas with custom rules
- **Coverage Analysis**: Analyze how much of your schema is used by your documents
- **Deprecated Usage Detection**: Find usage of deprecated fields and types
- **Query Complexity Analysis**: Analyze and limit query complexity
- **Flexible Input**: Support for files, URLs, and direct schema/document strings
- **Multiple Output Formats**: Human-readable text and JSON output
- **Configurable Rules**: Custom validation rules and thresholds

## 📦 Installation

### Install from source

```bash
go install github.com/bishnuag/graphql-inspector@latest
```

### Build from source

```bash
git clone https://github.com/bishnuag/graphql-inspector.git
cd graphql-inspector
go build -o graphql-inspector .
```

## 🔧 Usage

### Schema Comparison

Compare two GraphQL schemas to detect changes:

```bash
# Compare two schema files
graphql-inspector diff old-schema.graphql new-schema.graphql

# Compare with options
graphql-inspector diff old-schema.graphql new-schema.graphql --ignore-descriptions

# JSON output
graphql-inspector diff old-schema.graphql new-schema.graphql --json

# Fail on breaking changes
graphql-inspector diff old-schema.graphql new-schema.graphql --fail-on-breaking
```

### Document Validation

Validate GraphQL documents against a schema:

```bash
# Validate documents
graphql-inspector validate "queries/*.graphql" schema.graphql

# With custom limits
graphql-inspector validate queries/ schema.graphql --max-depth 10 --max-tokens 500

# Check for deprecated usage
graphql-inspector validate queries/ schema.graphql --check-deprecated
```

### Coverage Analysis

Analyze schema coverage based on your documents:

```bash
# Basic coverage analysis
graphql-inspector coverage "queries/*.graphql" schema.graphql

# With threshold
graphql-inspector coverage queries/ schema.graphql --threshold 0.8

# Show unused types and fields
graphql-inspector coverage queries/ schema.graphql --show-unused --show-details
```

### Global Options

```bash
# Verbose output
graphql-inspector --verbose <command>

# JSON output
graphql-inspector --json <command>

# Configuration file
graphql-inspector --config ~/.graphql-inspector.yaml <command>
```

## 📝 Configuration

Create a `.graphql-inspector.yaml` configuration file:

```yaml
# Schema and documents paths
schemaPath: "schema.graphql"
documentsPaths:
  - "queries/**/*.graphql"
  - "mutations/**/*.graphql"

# Validation rules
rules:
  - "no-unused-types"
  - "no-deprecated-fields"

# Thresholds
thresholds:
  coverage: 0.8
  maxDepth: 15
```

## 🛠️ API Usage

Use GraphQL Inspector as a Go library:

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/bishnuag/graphql-inspector/pkg/core"
    "github.com/bishnuag/graphql-inspector/pkg/loader"
)

func main() {
    // Load schemas
    oldSchema, err := loader.LoadSchema("old-schema.graphql")
    if err != nil {
        log.Fatal(err)
    }
    
    newSchema, err := loader.LoadSchema("new-schema.graphql")
    if err != nil {
        log.Fatal(err)
    }
    
    // Compare schemas
    changes, err := core.DiffSchemas(oldSchema, newSchema, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Print changes
    for _, change := range changes {
        fmt.Printf("%s: %s\n", change.Type, change.Message)
    }
}
```

## 🧪 Examples

### Example Schema Comparison

```bash
$ graphql-inspector diff old-schema.graphql new-schema.graphql

Found 3 changes:
  - 1 breaking
  - 1 dangerous
  - 1 non-breaking

🔴 Breaking Changes (1):
========================
  💥 Field 'User.email' was removed (at User.email)

🟡 Dangerous Changes (1):
=========================
  ⚠️  Field 'User.name' changed type from String! to String (at User.name)

🟢 Non-Breaking Changes (1):
=============================
  ✨ Field 'User.avatar' was added (at User.avatar)
```

### Example Coverage Analysis

```bash
$ graphql-inspector coverage "queries/*.graphql" schema.graphql --show-unused

GraphQL Schema Coverage Analysis
===============================

📊 Coverage Summary:
  Overall Coverage: 75.50%
  Type Coverage:    80.00% (8/10)
  Field Coverage:   72.50% (29/40)

✅ Coverage 75.50% meets threshold 80.00%

🗑️  Unused Types (2):
====================
  • InternalUser
  • DebugInfo

🗑️  Unused Fields:
==================
  User:
    • internalId
    • debugInfo
  Post:
    • internalNotes
```

## 🔧 Development

### Prerequisites

- Go 1.21 or higher
- Make (optional)

### Setup

```bash
# Clone the repository
git clone https://github.com/bishnuag/graphql-inspector.git
cd graphql-inspector

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build
go build -o graphql-inspector .
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/core/...
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgements

- Inspired by the original [GraphQL Inspector](https://github.com/graphql-hive/graphql-inspector) project
- Built with [graphql-go](https://github.com/graphql-go/graphql)
- CLI built with [Cobra](https://github.com/spf13/cobra)

## 📚 Related Projects

- [GraphQL Inspector (Original)](https://github.com/graphql-hive/graphql-inspector) - The original TypeScript/JavaScript version
- [GraphQL Hive](https://graphql-hive.com) - GraphQL schema registry and monitoring
- [GraphQL Tools](https://github.com/ardatan/graphql-tools) - GraphQL schema manipulation tools

## 🔗 Links

- [Documentation](https://github.com/bishnuag/graphql-inspector/docs)
- [Issues](https://github.com/bishnuag/graphql-inspector/issues)
- [Releases](https://github.com/bishnuag/graphql-inspector/releases) 