package core

import (
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/visitor"
)

// ValidateDocuments validates GraphQL documents against a schema
func ValidateDocuments(schema *Schema, documents []Document, options *ValidateOptions) ([]ValidationResult, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is required")
	}

	if options == nil {
		options = &ValidateOptions{
			Schema:     schema,
			MaxDepth:   15,
			MaxTokens:  1000,
			MaxAliases: 15,
		}
	}

	results := make([]ValidationResult, 0, len(documents))

	for _, doc := range documents {
		result := validateDocument(schema, doc, options)
		results = append(results, result)
	}

	return results, nil
}

// validateDocument validates a single GraphQL document
func validateDocument(schema *Schema, doc Document, options *ValidateOptions) ValidationResult {
	var errors []string

	// Parse the document if AST is not provided
	var docAST *ast.Document
	if doc.AST != nil {
		docAST = doc.AST
	} else {
		parsed, err := parser.Parse(parser.ParseParams{
			Source: doc.Content,
		})
		if err != nil {
			return ValidationResult{
				IsValid: false,
				Errors:  []string{fmt.Sprintf("Parse error: %v", err)},
			}
		}
		docAST = parsed
	}

	// Validate document against schema
	validationResult := graphql.ValidateDocument(schema.Schema, docAST, nil)
	if validationResult.IsValid == false {
		for _, err := range validationResult.Errors {
			errors = append(errors, fmt.Sprintf("Validation error: %v", err))
		}
	}

	// Custom validation rules
	customErrors := applyCustomValidationRules(docAST, options)
	errors = append(errors, customErrors...)

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// applyCustomValidationRules applies custom validation rules to the document
func applyCustomValidationRules(docAST *ast.Document, options *ValidateOptions) []string {
	var errors []string

	// Validate query depth
	if options.MaxDepth > 0 {
		if depthErrors := validateQueryDepth(docAST, options.MaxDepth); len(depthErrors) > 0 {
			errors = append(errors, depthErrors...)
		}
	}

	// Validate token count
	if options.MaxTokens > 0 {
		if tokenErrors := validateTokenCount(docAST, options.MaxTokens); len(tokenErrors) > 0 {
			errors = append(errors, tokenErrors...)
		}
	}

	// Validate alias count
	if options.MaxAliases > 0 {
		if aliasErrors := validateAliasCount(docAST, options.MaxAliases); len(aliasErrors) > 0 {
			errors = append(errors, aliasErrors...)
		}
	}

	return errors
}

// validateQueryDepth validates the depth of GraphQL queries
func validateQueryDepth(docAST *ast.Document, maxDepth int) []string {
	var errors []string
	
	visitor.Visit(docAST, &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			if field, ok := p.Node.(*ast.Field); ok {
				depth := calculateFieldDepth(field, 0)
				if depth > maxDepth {
					errors = append(errors, fmt.Sprintf("Query depth %d exceeds maximum allowed depth of %d", depth, maxDepth))
				}
			}
			return visitor.ActionNoChange, nil
		},
	}, nil)

	return errors
}

// calculateFieldDepth calculates the depth of a field
func calculateFieldDepth(field *ast.Field, currentDepth int) int {
	if field.SelectionSet == nil {
		return currentDepth + 1
	}

	maxDepth := currentDepth + 1
	for _, selection := range field.SelectionSet.Selections {
		switch sel := selection.(type) {
		case *ast.Field:
			depth := calculateFieldDepth(sel, currentDepth+1)
			if depth > maxDepth {
				maxDepth = depth
			}
		case *ast.InlineFragment:
			for _, fragSelection := range sel.SelectionSet.Selections {
				if fragField, ok := fragSelection.(*ast.Field); ok {
					depth := calculateFieldDepth(fragField, currentDepth+1)
					if depth > maxDepth {
						maxDepth = depth
					}
				}
			}
		}
	}

	return maxDepth
}

// validateTokenCount validates the number of tokens in a GraphQL query
func validateTokenCount(docAST *ast.Document, maxTokens int) []string {
	var errors []string
	tokenCount := 0

	visitor.Visit(docAST, &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			tokenCount++
			return visitor.ActionNoChange, nil
		},
	}, nil)

	if tokenCount > maxTokens {
		errors = append(errors, fmt.Sprintf("Query has %d tokens, exceeding maximum of %d", tokenCount, maxTokens))
	}

	return errors
}

// validateAliasCount validates the number of aliases in a GraphQL query
func validateAliasCount(docAST *ast.Document, maxAliases int) []string {
	var errors []string
	aliasCount := 0

	visitor.Visit(docAST, &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			if field, ok := p.Node.(*ast.Field); ok {
				if field.Alias != nil {
					aliasCount++
				}
			}
			return visitor.ActionNoChange, nil
		},
	}, nil)

	if aliasCount > maxAliases {
		errors = append(errors, fmt.Sprintf("Query has %d aliases, exceeding maximum of %d", aliasCount, maxAliases))
	}

	return errors
}

// FindDeprecatedUsage finds deprecated field usage in documents
func FindDeprecatedUsage(schema *Schema, documents []Document) ([]DeprecatedUsage, error) {
	var deprecated []DeprecatedUsage

	for _, doc := range documents {
		// Parse the document if AST is not provided
		var docAST *ast.Document
		if doc.AST != nil {
			docAST = doc.AST
		} else {
			parsed, err := parser.Parse(parser.ParseParams{
				Source: doc.Content,
			})
			if err != nil {
				continue // Skip invalid documents
			}
			docAST = parsed
		}

		// Find deprecated usage
		visitor.Visit(docAST, &visitor.VisitorOptions{
			Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
				if field, ok := p.Node.(*ast.Field); ok {
					if usage := checkDeprecatedField(schema, field); usage != nil {
						usage.Source = doc.Source
						deprecated = append(deprecated, *usage)
					}
				}
				return visitor.ActionNoChange, nil
			},
		}, nil)
	}

	return deprecated, nil
}

// DeprecatedUsage represents usage of a deprecated field
type DeprecatedUsage struct {
	Source     string `json:"source"`
	Field      string `json:"field"`
	Type       string `json:"type"`
	Reason     string `json:"reason"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
}

// checkDeprecatedField checks if a field is deprecated
func checkDeprecatedField(schema *Schema, field *ast.Field) *DeprecatedUsage {
	// This is a simplified implementation
	// In a real implementation, you would need to traverse the schema
	// and check for deprecated fields based on the field path
	
	// For now, we'll just check if the field name contains "deprecated"
	if strings.Contains(strings.ToLower(field.Name.Value), "deprecated") {
		return &DeprecatedUsage{
			Field:  field.Name.Value,
			Type:   "FIELD",
			Reason: "Field is deprecated",
			Line:   field.Loc.Start,
			Column: field.Loc.End,
		}
	}

	return nil
}

// ValidateOperationComplexity validates the complexity of GraphQL operations
func ValidateOperationComplexity(schema *Schema, documents []Document, maxComplexity int) ([]ComplexityResult, error) {
	var results []ComplexityResult

	for _, doc := range documents {
		// Parse the document if AST is not provided
		var docAST *ast.Document
		if doc.AST != nil {
			docAST = doc.AST
		} else {
			parsed, err := parser.Parse(parser.ParseParams{
				Source: doc.Content,
			})
			if err != nil {
				continue // Skip invalid documents
			}
			docAST = parsed
		}

		// Calculate complexity for each operation
		for _, def := range docAST.Definitions {
			if opDef, ok := def.(*ast.OperationDefinition); ok {
				complexity := calculateOperationComplexity(opDef)
				results = append(results, ComplexityResult{
					Source:     doc.Source,
					Operation:  getOperationName(opDef),
					Complexity: complexity,
					IsValid:    complexity <= maxComplexity,
				})
			}
		}
	}

	return results, nil
}

// ComplexityResult represents the complexity analysis result
type ComplexityResult struct {
	Source     string `json:"source"`
	Operation  string `json:"operation"`
	Complexity int    `json:"complexity"`
	IsValid    bool   `json:"isValid"`
}

// calculateOperationComplexity calculates the complexity of an operation
func calculateOperationComplexity(opDef *ast.OperationDefinition) int {
	// Simple complexity calculation - count the number of fields
	// In a real implementation, this would be more sophisticated
	complexity := 0
	
	if opDef.SelectionSet != nil {
		complexity = countSelections(opDef.SelectionSet)
	}
	
	return complexity
}

// countSelections counts the number of selections in a selection set
func countSelections(selectionSet *ast.SelectionSet) int {
	count := 0
	
	for _, selection := range selectionSet.Selections {
		switch sel := selection.(type) {
		case *ast.Field:
			count++
			if sel.SelectionSet != nil {
				count += countSelections(sel.SelectionSet)
			}
		case *ast.InlineFragment:
			if sel.SelectionSet != nil {
				count += countSelections(sel.SelectionSet)
			}
		case *ast.FragmentSpread:
			count++
		}
	}
	
	return count
}

// getOperationName gets the name of an operation
func getOperationName(opDef *ast.OperationDefinition) string {
	if opDef.Name != nil {
		return opDef.Name.Value
	}
	return fmt.Sprintf("Anonymous%s", strings.Title(opDef.Operation))
} 