package core

import (
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/visitor"
)

// AnalyzeCoverage analyzes schema coverage based on GraphQL documents
func AnalyzeCoverage(schema *Schema, documents []Document, options *CoverageOptions) (*CoverageResult, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is required")
	}

	if options == nil {
		options = &CoverageOptions{
			Schema:    schema,
			Documents: documents,
			Threshold: 0.8,
		}
	}

	// Initialize coverage tracking
	coverage := initializeCoverage(schema.Schema)

	// Analyze each document
	for _, doc := range documents {
		if err := analyzeDocument(schema, doc, coverage); err != nil {
			continue // Skip invalid documents
		}
	}

	// Calculate coverage statistics
	result := calculateCoverageStats(coverage)
	return result, nil
}

// initializeCoverage initializes the coverage tracking structure
func initializeCoverage(schema *graphql.Schema) map[string]*TypeCoverage {
	coverage := make(map[string]*TypeCoverage)

	// Initialize coverage for all types
	for typeName, graphqlType := range schema.TypeMap() {
		// Skip built-in types
		if strings.HasPrefix(typeName, "__") {
			continue
		}

		typeCoverage := &TypeCoverage{
			Type:       typeName,
			Covered:    false,
			Fields:     make(map[string]bool),
			UsageCount: 0,
		}

		// Initialize fields for object and interface types
		switch t := graphqlType.(type) {
		case *graphql.Object:
			for fieldName := range t.Fields() {
				typeCoverage.Fields[fieldName] = false
			}
		case *graphql.Interface:
			for fieldName := range t.Fields() {
				typeCoverage.Fields[fieldName] = false
			}
		}

		coverage[typeName] = typeCoverage
	}

	return coverage
}

// analyzeDocument analyzes a single document for coverage
func analyzeDocument(schema *Schema, doc Document, coverage map[string]*TypeCoverage) error {
	// Parse the document if AST is not provided
	var docAST *ast.Document
	if doc.AST != nil {
		docAST = doc.AST
	} else {
		parsed, err := parser.Parse(parser.ParseParams{
			Source: doc.Content,
		})
		if err != nil {
			return err
		}
		docAST = parsed
	}

	// Visit the document and track field usage
	visitor.Visit(docAST, &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			if field, ok := p.Node.(*ast.Field); ok {
				// Track field usage
				if err := trackFieldUsage(schema, field, coverage, nil); err != nil {
					// Continue on error
				}
			}
			return visitor.ActionNoChange, nil
		},
	}, nil)

	return nil
}

// trackFieldUsage tracks the usage of a field in the coverage analysis
func trackFieldUsage(schema *Schema, field *ast.Field, coverage map[string]*TypeCoverage, parentType *graphql.Object) error {
	// This is a simplified implementation
	// In a real implementation, you would need to maintain context about the current type
	// and traverse the schema to find the correct field
	
	fieldName := field.Name.Value
	
	// For now, we'll mark any type that has this field name as used
	for _, typeCoverage := range coverage {
		if _, exists := typeCoverage.Fields[fieldName]; exists {
			typeCoverage.Covered = true
			typeCoverage.Fields[fieldName] = true
			typeCoverage.UsageCount++
		}
	}
	
	return nil
}

// calculateCoverageStats calculates the final coverage statistics
func calculateCoverageStats(coverage map[string]*TypeCoverage) *CoverageResult {
	totalTypes := len(coverage)
	typesCovered := 0
	totalFields := 0
	fieldsCovered := 0

	for _, typeCoverage := range coverage {
		if typeCoverage.Covered {
			typesCovered++
		}

		for _, covered := range typeCoverage.Fields {
			totalFields++
			if covered {
				fieldsCovered++
			}
		}
	}

	var overallCoverage float64
	if totalFields > 0 {
		overallCoverage = float64(fieldsCovered) / float64(totalFields)
	}

	return &CoverageResult{
		Coverage:      overallCoverage,
		TypesCovered:  typesCovered,
		TotalTypes:    totalTypes,
		FieldsCovered: fieldsCovered,
		TotalFields:   totalFields,
		Details:       convertCoverageMap(coverage),
	}
}

// convertCoverageMap converts the internal coverage map to the public format
func convertCoverageMap(coverage map[string]*TypeCoverage) map[string]TypeCoverage {
	result := make(map[string]TypeCoverage)
	for typeName, typeCoverage := range coverage {
		result[typeName] = *typeCoverage
	}
	return result
}

// GenerateCoverageReport generates a detailed coverage report
func GenerateCoverageReport(result *CoverageResult) string {
	var report strings.Builder

	report.WriteString("GraphQL Schema Coverage Report\n")
	report.WriteString("==============================\n\n")

	report.WriteString(fmt.Sprintf("Overall Coverage: %.2f%%\n", result.Coverage*100))
	report.WriteString(fmt.Sprintf("Types Covered: %d/%d (%.2f%%)\n", 
		result.TypesCovered, result.TotalTypes, 
		float64(result.TypesCovered)/float64(result.TotalTypes)*100))
	report.WriteString(fmt.Sprintf("Fields Covered: %d/%d (%.2f%%)\n\n", 
		result.FieldsCovered, result.TotalFields, 
		float64(result.FieldsCovered)/float64(result.TotalFields)*100))

	report.WriteString("Type Coverage Details:\n")
	report.WriteString("=====================\n\n")

	for typeName, typeCoverage := range result.Details {
		status := "❌ NOT COVERED"
		if typeCoverage.Covered {
			status = "✅ COVERED"
		}

		report.WriteString(fmt.Sprintf("%s %s (used %d times)\n", status, typeName, typeCoverage.UsageCount))

		if len(typeCoverage.Fields) > 0 {
			report.WriteString("  Fields:\n")
			for fieldName, covered := range typeCoverage.Fields {
				fieldStatus := "❌"
				if covered {
					fieldStatus = "✅"
				}
				report.WriteString(fmt.Sprintf("    %s %s\n", fieldStatus, fieldName))
			}
		}
		report.WriteString("\n")
	}

	return report.String()
}

// FindUnusedTypes finds types that are not used in any documents
func FindUnusedTypes(schema *Schema, documents []Document) ([]string, error) {
	result, err := AnalyzeCoverage(schema, documents, nil)
	if err != nil {
		return nil, err
	}

	var unusedTypes []string
	for typeName, typeCoverage := range result.Details {
		if !typeCoverage.Covered {
			unusedTypes = append(unusedTypes, typeName)
		}
	}

	return unusedTypes, nil
}

// FindUnusedFields finds fields that are not used in any documents
func FindUnusedFields(schema *Schema, documents []Document) (map[string][]string, error) {
	result, err := AnalyzeCoverage(schema, documents, nil)
	if err != nil {
		return nil, err
	}

	unusedFields := make(map[string][]string)
	for typeName, typeCoverage := range result.Details {
		var unused []string
		for fieldName, covered := range typeCoverage.Fields {
			if !covered {
				unused = append(unused, fieldName)
			}
		}
		if len(unused) > 0 {
			unusedFields[typeName] = unused
		}
	}

	return unusedFields, nil
}

// AnalyzeFieldUsage analyzes how frequently fields are used
func AnalyzeFieldUsage(schema *Schema, documents []Document) (map[string]FieldUsage, error) {
	fieldUsage := make(map[string]FieldUsage)

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

		// Visit the document and count field usage
		visitor.Visit(docAST, &visitor.VisitorOptions{
			Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
				if field, ok := p.Node.(*ast.Field); ok {
					fieldName := field.Name.Value
					if usage, exists := fieldUsage[fieldName]; exists {
						usage.Count++
						fieldUsage[fieldName] = usage
					} else {
						fieldUsage[fieldName] = FieldUsage{
							Field: fieldName,
							Count: 1,
						}
					}
				}
				return visitor.ActionNoChange, nil
			},
		}, nil)
	}

	return fieldUsage, nil
}

// FieldUsage represents field usage statistics
type FieldUsage struct {
	Field string `json:"field"`
	Count int    `json:"count"`
}

// GetCoverageSummary returns a summary of coverage statistics
func GetCoverageSummary(result *CoverageResult) CoverageSummary {
	return CoverageSummary{
		OverallCoverage: result.Coverage,
		TypeCoverage:    float64(result.TypesCovered) / float64(result.TotalTypes),
		FieldCoverage:   float64(result.FieldsCovered) / float64(result.TotalFields),
		TotalTypes:      result.TotalTypes,
		TotalFields:     result.TotalFields,
		CoveredTypes:    result.TypesCovered,
		CoveredFields:   result.FieldsCovered,
	}
}

// CoverageSummary represents a summary of coverage statistics
type CoverageSummary struct {
	OverallCoverage float64 `json:"overallCoverage"`
	TypeCoverage    float64 `json:"typeCoverage"`
	FieldCoverage   float64 `json:"fieldCoverage"`
	TotalTypes      int     `json:"totalTypes"`
	TotalFields     int     `json:"totalFields"`
	CoveredTypes    int     `json:"coveredTypes"`
	CoveredFields   int     `json:"coveredFields"`
} 