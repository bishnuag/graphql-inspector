# GraphQL Inspector Architecture

This document describes the architecture of the GraphQL Inspector project, rewritten in Go.

## Project Structure

```
graphql-inspector/
├── main.go                 # Entry point for the CLI
├── go.mod                  # Go module definition
├── go.sum                  # Go module dependencies
├── README.md               # Project documentation
├── LICENSE                 # MIT license
├── ARCHITECTURE.md         # This file
├── examples/               # Example schemas and queries
│   ├── old-schema.graphql
│   ├── new-schema.graphql
│   └── query.graphql
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and configuration
│   ├── diff.go            # Schema comparison command
│   ├── validate.go        # Document validation command
│   └── coverage.go        # Coverage analysis command
└── pkg/                   # Core packages
    ├── core/              # Core functionality
    │   ├── types.go       # Common types and interfaces
    │   ├── diff.go        # Schema comparison logic
    │   ├── validate.go    # Document validation logic
    │   └── coverage.go    # Coverage analysis logic
    └── loader/            # Schema and document loading
        └── schema.go      # Schema loading utilities
```

## Core Components

### 1. CLI Layer (`cmd/`)

The CLI layer provides the user interface using the Cobra framework:

- **root.go**: Main command configuration, global flags, and config file handling
- **diff.go**: Schema comparison command implementation
- **validate.go**: Document validation command implementation
- **coverage.go**: Coverage analysis command implementation

### 2. Core Library (`pkg/core/`)

The core library contains the main business logic:

- **types.go**: Common data structures and interfaces
- **diff.go**: Schema comparison algorithms
- **validate.go**: Document validation and analysis
- **coverage.go**: Schema coverage analysis

### 3. Loader (`pkg/loader/`)

The loader package handles loading schemas and documents from various sources:

- **schema.go**: Schema loading from files, URLs, and strings

## Key Features

### Schema Comparison (`diff`)

- Compares two GraphQL schemas
- Detects breaking, dangerous, and non-breaking changes
- Provides detailed change information with paths and metadata
- Supports various output formats (text, JSON)

### Document Validation (`validate`)

- Validates GraphQL documents against schemas
- Checks for syntax errors and schema compliance
- Analyzes query depth, complexity, and token counts
- Detects deprecated field usage

### Coverage Analysis (`coverage`)

- Analyzes schema coverage based on documents
- Identifies unused types and fields
- Provides detailed coverage statistics
- Supports coverage thresholds

## Technical Decisions

### GraphQL Library Choice

The project uses `github.com/graphql-go/graphql` as the core GraphQL library. This library provides:

- GraphQL schema parsing and validation
- AST (Abstract Syntax Tree) manipulation
- Document parsing and validation

### Limitations and Future Improvements

1. **Schema Loading**: The current implementation uses a simplified schema loader that creates basic schemas. A full implementation would need to:
   - Parse SDL (Schema Definition Language) completely
   - Build proper GraphQL types from AST
   - Handle complex type relationships

2. **Type System**: The current diff algorithm is simplified. A complete implementation would need to:
   - Handle all GraphQL type kinds (scalars, objects, interfaces, unions, enums, input types)
   - Implement proper type compatibility checking
   - Support directive comparison

3. **Coverage Analysis**: The current coverage tracking is basic. Improvements could include:
   - Better type context tracking
   - Field-level usage statistics
   - Integration with schema introspection

### Dependencies

- **github.com/spf13/cobra**: CLI framework
- **github.com/spf13/viper**: Configuration management
- **github.com/graphql-go/graphql**: GraphQL library
- **gopkg.in/yaml.v3**: YAML parsing

## Extension Points

The architecture is designed to be extensible:

1. **Custom Validation Rules**: The validation system can be extended with custom rules
2. **Additional Loaders**: New schema/document loaders can be added
3. **Output Formats**: New output formats can be easily added
4. **Diff Rules**: Custom diff rules can be implemented

## Testing Strategy

The project structure supports comprehensive testing:

- Unit tests for core functionality
- Integration tests for CLI commands
- Example schemas and queries for manual testing

## Performance Considerations

- Schema loading is optimized for single-use scenarios
- AST parsing is done once per schema/document
- Memory usage is optimized for large schemas
- Concurrent processing can be added for multiple documents

## Configuration

The tool supports configuration through:

- Command-line flags
- Environment variables
- Configuration files (YAML format)
- Default values

## Error Handling

The project uses Go's standard error handling patterns:

- Errors are returned from functions
- CLI commands handle errors gracefully
- Detailed error messages are provided
- Exit codes indicate success/failure

## Future Enhancements

1. **Complete SDL Parser**: Implement full Schema Definition Language parsing
2. **GraphQL Server Integration**: Add support for introspection from live servers
3. **Advanced Diff Rules**: Implement more sophisticated change detection
4. **Performance Optimization**: Add caching and concurrent processing
5. **Plugin System**: Allow for custom validation rules and output formats
6. **Web UI**: Add a web interface for visualization
7. **CI/CD Integration**: Add GitHub Actions and other CI/CD integrations 