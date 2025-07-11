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

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate <documents> <schema>",
	Short: "Validate GraphQL documents against a schema",
	Long: `Validate GraphQL documents against a schema and check for errors.
	
The validate command checks GraphQL documents for syntax errors, schema compliance,
and custom validation rules like query depth, complexity, and token limits.

Examples:
  # Validate documents against a schema
  graphql-inspector validate "queries/*.graphql" schema.graphql
  
  # Validate with custom limits
  graphql-inspector validate queries/ schema.graphql --max-depth 10 --max-tokens 500
  
  # Find deprecated field usage
  graphql-inspector validate queries/ schema.graphql --check-deprecated`,
	Args: cobra.ExactArgs(2),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	
	// Validation-specific flags
	validateCmd.Flags().Int("max-depth", 15, "maximum query depth allowed")
	validateCmd.Flags().Int("max-tokens", 1000, "maximum tokens allowed in a query")
	validateCmd.Flags().Int("max-aliases", 15, "maximum aliases allowed in a query")
	validateCmd.Flags().Int("max-complexity", 1000, "maximum query complexity allowed")
	validateCmd.Flags().Bool("check-deprecated", false, "check for deprecated field usage")
	validateCmd.Flags().StringSlice("rules", []string{}, "custom validation rules")
	validateCmd.Flags().Bool("fail-on-error", true, "exit with non-zero code if validation errors are found")
	
	// Bind flags to viper
	viper.BindPFlag("validate.max-depth", validateCmd.Flags().Lookup("max-depth"))
	viper.BindPFlag("validate.max-tokens", validateCmd.Flags().Lookup("max-tokens"))
	viper.BindPFlag("validate.max-aliases", validateCmd.Flags().Lookup("max-aliases"))
	viper.BindPFlag("validate.max-complexity", validateCmd.Flags().Lookup("max-complexity"))
	viper.BindPFlag("validate.check-deprecated", validateCmd.Flags().Lookup("check-deprecated"))
	viper.BindPFlag("validate.rules", validateCmd.Flags().Lookup("rules"))
	viper.BindPFlag("validate.fail-on-error", validateCmd.Flags().Lookup("fail-on-error"))
}

func runValidate(cmd *cobra.Command, args []string) error {
	documentsPattern := args[0]
	schemaPath := args[1]
	
	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Validating documents: %s against schema: %s\n", documentsPattern, schemaPath)
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
		fmt.Fprintf(os.Stderr, "Found %d documents to validate\n", len(documents))
	}
	
	// Configure validation options
	options := &core.ValidateOptions{
		Schema:      schema,
		MaxDepth:    viper.GetInt("validate.max-depth"),
		MaxTokens:   viper.GetInt("validate.max-tokens"),
		MaxAliases:  viper.GetInt("validate.max-aliases"),
		CustomRules: viper.GetStringSlice("validate.rules"),
	}
	
	// Validate documents
	results, err := core.ValidateDocuments(schema, documents, options)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Check for deprecated usage if requested
	var deprecatedUsage []core.DeprecatedUsage
	if viper.GetBool("validate.check-deprecated") {
		deprecatedUsage, err = core.FindDeprecatedUsage(schema, documents)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to check deprecated usage: %v\n", err)
		}
	}
	
	// Check complexity if requested
	var complexityResults []core.ComplexityResult
	maxComplexity := viper.GetInt("validate.max-complexity")
	if maxComplexity > 0 {
		complexityResults, err = core.ValidateOperationComplexity(schema, documents, maxComplexity)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to check complexity: %v\n", err)
		}
	}
	
	// Output results
	if viper.GetBool("json") {
		return outputValidationJSON(results, deprecatedUsage, complexityResults)
	} else {
		return outputValidationText(results, deprecatedUsage, complexityResults)
	}
}

func outputValidationJSON(results []core.ValidationResult, deprecated []core.DeprecatedUsage, complexity []core.ComplexityResult) error {
	output := map[string]interface{}{
		"results":    results,
		"summary":    calculateValidationSummary(results),
		"deprecated": deprecated,
		"complexity": complexity,
	}
	
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputValidationText(results []core.ValidationResult, deprecated []core.DeprecatedUsage, complexity []core.ComplexityResult) error {
	summary := calculateValidationSummary(results)
	
	// Print summary
	fmt.Printf("Validation Results:\n")
	fmt.Printf("==================\n")
	fmt.Printf("Total documents: %d\n", summary.Total)
	fmt.Printf("Valid documents: %d\n", summary.Valid)
	fmt.Printf("Invalid documents: %d\n", summary.Invalid)
	fmt.Printf("Total errors: %d\n", summary.TotalErrors)
	fmt.Println()
	
	// Print validation errors
	if summary.Invalid > 0 {
		fmt.Printf("❌ Validation Errors:\n")
		fmt.Println("====================")
		
		for i, result := range results {
			if !result.IsValid {
				fmt.Printf("Document %d:\n", i+1)
				for _, error := range result.Errors {
					fmt.Printf("  • %s\n", error)
				}
				fmt.Println()
			}
		}
	}
	
	// Print deprecated usage
	if len(deprecated) > 0 {
		fmt.Printf("⚠️  Deprecated Usage (%d):\n", len(deprecated))
		fmt.Println("========================")
		for _, usage := range deprecated {
			fmt.Printf("  • %s in %s (%s)\n", usage.Field, usage.Source, usage.Reason)
		}
		fmt.Println()
	}
	
	// Print complexity results
	if len(complexity) > 0 {
		fmt.Printf("🔍 Complexity Analysis:\n")
		fmt.Println("=====================")
		for _, result := range complexity {
			status := "✅"
			if !result.IsValid {
				status = "❌"
			}
			fmt.Printf("  %s %s: %d (in %s)\n", status, result.Operation, result.Complexity, result.Source)
		}
		fmt.Println()
	}
	
	// Print success message or failure
	if summary.Invalid == 0 {
		fmt.Println("✅ All documents are valid!")
	} else {
		fmt.Printf("❌ %d documents have validation errors\n", summary.Invalid)
		
		if viper.GetBool("validate.fail-on-error") {
			return fmt.Errorf("validation failed")
		}
	}
	
	return nil
}

func calculateValidationSummary(results []core.ValidationResult) ValidationSummary {
	summary := ValidationSummary{
		Total: len(results),
	}
	
	for _, result := range results {
		if result.IsValid {
			summary.Valid++
		} else {
			summary.Invalid++
			summary.TotalErrors += len(result.Errors)
		}
	}
	
	return summary
}

type ValidationSummary struct {
	Total       int `json:"total"`
	Valid       int `json:"valid"`
	Invalid     int `json:"invalid"`
	TotalErrors int `json:"totalErrors"`
} 