package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bishnuag/graphql-inspector/pkg/core"
	"github.com/bishnuag/graphql-inspector/pkg/loader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// coverageCmd represents the coverage command
var coverageCmd = &cobra.Command{
	Use:   "coverage <documents> <schema>",
	Short: "Analyze GraphQL schema coverage based on documents",
	Long: `Analyze GraphQL schema coverage based on GraphQL documents.
	
The coverage command analyzes how much of your GraphQL schema is actually used
by your GraphQL documents (queries, mutations, subscriptions). It provides
detailed information about type and field coverage.

Examples:
  # Analyze coverage
  graphql-inspector coverage "queries/*.graphql" schema.graphql
  
  # Set coverage threshold
  graphql-inspector coverage queries/ schema.graphql --threshold 0.8
  
  # Find unused types and fields
  graphql-inspector coverage queries/ schema.graphql --show-unused`,
	Args: cobra.ExactArgs(2),
	RunE: runCoverage,
}

func init() {
	rootCmd.AddCommand(coverageCmd)
	
	// Coverage-specific flags
	coverageCmd.Flags().Float64("threshold", 0.8, "minimum coverage threshold")
	coverageCmd.Flags().Bool("show-unused", false, "show unused types and fields")
	coverageCmd.Flags().Bool("show-details", false, "show detailed coverage information")
	coverageCmd.Flags().Bool("fail-on-threshold", false, "exit with non-zero code if coverage is below threshold")
	
	// Bind flags to viper
	viper.BindPFlag("coverage.threshold", coverageCmd.Flags().Lookup("threshold"))
	viper.BindPFlag("coverage.show-unused", coverageCmd.Flags().Lookup("show-unused"))
	viper.BindPFlag("coverage.show-details", coverageCmd.Flags().Lookup("show-details"))
	viper.BindPFlag("coverage.fail-on-threshold", coverageCmd.Flags().Lookup("fail-on-threshold"))
}

func runCoverage(cmd *cobra.Command, args []string) error {
	documentsPattern := args[0]
	schemaPath := args[1]
	
	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Analyzing coverage for documents: %s against schema: %s\n", documentsPattern, schemaPath)
	}
	
	// Load schema
	schema, err := loader.LoadSchema(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}
	
	// Load documents
	documents, err := loader.LoadDocuments(documentsPattern)
	if err != nil {
		return fmt.Errorf("failed to load documents: %w", err)
	}
	
	if len(documents) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: No documents found matching pattern: %s\n", documentsPattern)
		return nil
	}
	
	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Found %d documents to analyze\n", len(documents))
	}
	
	// Configure coverage options
	options := &core.CoverageOptions{
		Schema:    schema,
		Documents: documents,
		Threshold: viper.GetFloat64("coverage.threshold"),
	}
	
	// Analyze coverage
	result, err := core.AnalyzeCoverage(schema, documents, options)
	if err != nil {
		return fmt.Errorf("coverage analysis failed: %w", err)
	}
	
	// Get additional information if requested
	var unusedTypes []string
	var unusedFields map[string][]string
	
	if viper.GetBool("coverage.show-unused") {
		unusedTypes, err = core.FindUnusedTypes(schema, documents)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to find unused types: %v\n", err)
		}
		
		unusedFields, err = core.FindUnusedFields(schema, documents)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to find unused fields: %v\n", err)
		}
	}
	
	// Output results
	if viper.GetBool("json") {
		return outputCoverageJSON(result, unusedTypes, unusedFields)
	} else {
		return outputCoverageText(result, unusedTypes, unusedFields)
	}
}

func outputCoverageJSON(result *core.CoverageResult, unusedTypes []string, unusedFields map[string][]string) error {
	output := map[string]interface{}{
		"coverage":     result,
		"summary":      core.GetCoverageSummary(result),
		"unusedTypes":  unusedTypes,
		"unusedFields": unusedFields,
	}
	
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputCoverageText(result *core.CoverageResult, unusedTypes []string, unusedFields map[string][]string) error {
	summary := core.GetCoverageSummary(result)
	
	// Print coverage summary
	fmt.Printf("GraphQL Schema Coverage Analysis\n")
	fmt.Printf("===============================\n\n")
	
	fmt.Printf("📊 Coverage Summary:\n")
	fmt.Printf("  Overall Coverage: %.2f%%\n", summary.OverallCoverage*100)
	fmt.Printf("  Type Coverage:    %.2f%% (%d/%d)\n", summary.TypeCoverage*100, summary.CoveredTypes, summary.TotalTypes)
	fmt.Printf("  Field Coverage:   %.2f%% (%d/%d)\n", summary.FieldCoverage*100, summary.CoveredFields, summary.TotalFields)
	fmt.Println()
	
	// Check threshold
	threshold := viper.GetFloat64("coverage.threshold")
	if summary.OverallCoverage < threshold {
		fmt.Printf("⚠️  Coverage %.2f%% is below threshold %.2f%%\n", summary.OverallCoverage*100, threshold*100)
		fmt.Println()
	} else {
		fmt.Printf("✅ Coverage %.2f%% meets threshold %.2f%%\n", summary.OverallCoverage*100, threshold*100)
		fmt.Println()
	}
	
	// Show detailed coverage if requested
	if viper.GetBool("coverage.show-details") {
		fmt.Printf("📋 Detailed Coverage:\n")
		fmt.Printf("====================\n")
		
		for typeName, typeCoverage := range result.Details {
			status := "❌"
			if typeCoverage.Covered {
				status = "✅"
			}
			
			fmt.Printf("%s %s", status, typeName)
			if typeCoverage.UsageCount > 0 {
				fmt.Printf(" (used %d times)", typeCoverage.UsageCount)
			}
			fmt.Println()
			
			if len(typeCoverage.Fields) > 0 {
				for fieldName, covered := range typeCoverage.Fields {
					fieldStatus := "❌"
					if covered {
						fieldStatus = "✅"
					}
					fmt.Printf("  %s %s\n", fieldStatus, fieldName)
				}
			}
		}
		fmt.Println()
	}
	
	// Show unused types and fields if requested
	if viper.GetBool("coverage.show-unused") {
		if len(unusedTypes) > 0 {
			fmt.Printf("🗑️  Unused Types (%d):\n", len(unusedTypes))
			fmt.Printf("====================\n")
			for _, typeName := range unusedTypes {
				fmt.Printf("  • %s\n", typeName)
			}
			fmt.Println()
		}
		
		if len(unusedFields) > 0 {
			fmt.Printf("🗑️  Unused Fields:\n")
			fmt.Printf("==================\n")
			for typeName, fields := range unusedFields {
				fmt.Printf("  %s:\n", typeName)
				for _, fieldName := range fields {
					fmt.Printf("    • %s\n", fieldName)
				}
			}
			fmt.Println()
		}
	}
	
	// Generate coverage report
	if !viper.GetBool("json") && !viper.GetBool("coverage.show-details") {
		fmt.Println("💡 Use --show-details to see detailed coverage information")
		fmt.Println("💡 Use --show-unused to see unused types and fields")
	}
	
	// Check failure condition
	if viper.GetBool("coverage.fail-on-threshold") && summary.OverallCoverage < threshold {
		return fmt.Errorf("coverage %.2f%% is below threshold %.2f%%", summary.OverallCoverage*100, threshold*100)
	}
	
	return nil
}

// Additional helper functions for coverage analysis

func printCoverageBar(coverage float64) string {
	const barWidth = 20
	filled := int(coverage * barWidth)
	bar := "["
	
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	
	bar += "]"
	return bar
}

func getCoverageColor(coverage float64) string {
	if coverage >= 0.8 {
		return "🟢" // Green
	} else if coverage >= 0.6 {
		return "🟡" // Yellow
	} else {
		return "🔴" // Red
	}
} 