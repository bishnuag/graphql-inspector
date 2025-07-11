package core

import (
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// ChangeType represents the type of change detected
type ChangeType string

const (
	// Breaking changes
	ChangeTypeBreaking ChangeType = "BREAKING"
	// Dangerous changes
	ChangeTypeDangerous ChangeType = "DANGEROUS"
	// Non-breaking changes
	ChangeTypeNonBreaking ChangeType = "NON_BREAKING"
)

// Change represents a detected change between two schemas
type Change struct {
	Type        ChangeType `json:"type"`
	Message     string     `json:"message"`
	Path        string     `json:"path,omitempty"`
	Criticality string     `json:"criticality"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// Schema represents a GraphQL schema with additional metadata
type Schema struct {
	Schema    *graphql.Schema `json:"-"`
	SDL       string          `json:"sdl"`
	Hash      string          `json:"hash"`
	Source    string          `json:"source"`
	Timestamp time.Time       `json:"timestamp"`
}

// Document represents a GraphQL document/operation
type Document struct {
	Source    string             `json:"source"`
	Content   string             `json:"content"`
	AST       *ast.Document      `json:"-"`
	Hash      string             `json:"hash"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// ValidationResult represents the result of document validation
type ValidationResult struct {
	IsValid bool     `json:"isValid"`
	Errors  []string `json:"errors,omitempty"`
}

// CoverageResult represents schema coverage analysis
type CoverageResult struct {
	Coverage    float64                    `json:"coverage"`
	TypesCovered int                      `json:"typesCovered"`
	TotalTypes   int                      `json:"totalTypes"`
	FieldsCovered int                     `json:"fieldsCovered"`
	TotalFields   int                     `json:"totalFields"`
	Details      map[string]TypeCoverage  `json:"details"`
}

// TypeCoverage represents coverage for a specific type
type TypeCoverage struct {
	Type         string            `json:"type"`
	Covered      bool              `json:"covered"`
	Fields       map[string]bool   `json:"fields"`
	UsageCount   int               `json:"usageCount"`
}

// SimilarType represents a similar type found in the schema
type SimilarType struct {
	Type       string  `json:"type"`
	Similarity float64 `json:"similarity"`
	Reason     string  `json:"reason"`
}

// DiffOptions represents options for schema comparison
type DiffOptions struct {
	IgnoreDescriptions bool     `json:"ignoreDescriptions"`
	IgnoreDirectives   bool     `json:"ignoreDirectives"`
	CustomRules        []string `json:"customRules,omitempty"`
}

// ValidateOptions represents options for document validation
type ValidateOptions struct {
	Schema       *Schema   `json:"-"`
	MaxDepth     int       `json:"maxDepth"`
	MaxTokens    int       `json:"maxTokens"`
	MaxAliases   int       `json:"maxAliases"`
	CustomRules  []string  `json:"customRules,omitempty"`
}

// CoverageOptions represents options for coverage analysis
type CoverageOptions struct {
	Schema     *Schema     `json:"-"`
	Documents  []Document  `json:"documents"`
	Threshold  float64     `json:"threshold"`
}

// InspectorConfig represents the configuration for GraphQL Inspector
type InspectorConfig struct {
	SchemaPath     string   `yaml:"schemaPath"`
	DocumentsPaths []string `yaml:"documentsPaths"`
	Rules          []string `yaml:"rules"`
	Thresholds     struct {
		Coverage float64 `yaml:"coverage"`
		MaxDepth int     `yaml:"maxDepth"`
	} `yaml:"thresholds"`
} 