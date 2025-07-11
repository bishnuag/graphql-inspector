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

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff <old-schema> <new-schema>",
	Short: "Compare two GraphQL schemas and detect changes",
	Long: `Compare two GraphQL schemas and detect breaking, dangerous, and non-breaking changes.
	
The diff command analyzes the differences between two GraphQL schemas and provides
detailed information about each change, including its type and potential impact.

Examples:
  # Compare two schema files
  graphql-inspector diff old-schema.graphql new-schema.graphql
  
  # Compare with options
  graphql-inspector diff old-schema.graphql new-schema.graphql --ignore-descriptions
  
  # Output in JSON format
  graphql-inspector diff old-schema.graphql new-schema.graphql --json`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)
	
	// Diff-specific flags
	diffCmd.Flags().Bool("ignore-descriptions", false, "ignore description changes")
	diffCmd.Flags().Bool("ignore-directives", false, "ignore directive changes")
	diffCmd.Flags().StringSlice("rules", []string{}, "custom rules to apply")
	diffCmd.Flags().Bool("fail-on-breaking", false, "exit with non-zero code if breaking changes are found")
	diffCmd.Flags().Bool("fail-on-dangerous", false, "exit with non-zero code if dangerous changes are found")
	
	// Bind flags to viper
	viper.BindPFlag("diff.ignore-descriptions", diffCmd.Flags().Lookup("ignore-descriptions"))
	viper.BindPFlag("diff.ignore-directives", diffCmd.Flags().Lookup("ignore-directives"))
	viper.BindPFlag("diff.rules", diffCmd.Flags().Lookup("rules"))
	viper.BindPFlag("diff.fail-on-breaking", diffCmd.Flags().Lookup("fail-on-breaking"))
	viper.BindPFlag("diff.fail-on-dangerous", diffCmd.Flags().Lookup("fail-on-dangerous"))
}

func runDiff(cmd *cobra.Command, args []string) error {
	oldSchemaPath := args[0]
	newSchemaPath := args[1]
	
	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Comparing schemas: %s -> %s\n", oldSchemaPath, newSchemaPath)
	}
	
	// Load schemas
	oldSchema, err := loader.LoadSchema(oldSchemaPath)
	if err != nil {
		return fmt.Errorf("failed to load old schema: %w", err)
	}
	
	newSchema, err := loader.LoadSchema(newSchemaPath)
	if err != nil {
		return fmt.Errorf("failed to load new schema: %w", err)
	}
	
	// Configure diff options
	options := &core.DiffOptions{
		IgnoreDescriptions: viper.GetBool("diff.ignore-descriptions"),
		IgnoreDirectives:   viper.GetBool("diff.ignore-directives"),
		CustomRules:        viper.GetStringSlice("diff.rules"),
	}
	
	// Compare schemas
	changes, err := core.DiffSchemas(oldSchema, newSchema, options)
	if err != nil {
		return fmt.Errorf("failed to compare schemas: %w", err)
	}
	
	// Output results
	if viper.GetBool("json") {
		return outputDiffJSON(changes)
	} else {
		return outputDiffText(changes)
	}
}

func outputDiffJSON(changes []core.Change) error {
	output := map[string]interface{}{
		"changes": changes,
		"summary": calculateDiffSummary(changes),
	}
	
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputDiffText(changes []core.Change) error {
	if len(changes) == 0 {
		fmt.Println("✅ No changes detected")
		return nil
	}
	
	summary := calculateDiffSummary(changes)
	
	// Print summary
	fmt.Printf("Found %d changes:\n", len(changes))
	fmt.Printf("  - %d breaking\n", summary.Breaking)
	fmt.Printf("  - %d dangerous\n", summary.Dangerous)
	fmt.Printf("  - %d non-breaking\n", summary.NonBreaking)
	fmt.Println()
	
	// Group changes by type
	breakingChanges := filterChangesByType(changes, core.ChangeTypeBreaking)
	dangerousChanges := filterChangesByType(changes, core.ChangeTypeDangerous)
	nonBreakingChanges := filterChangesByType(changes, core.ChangeTypeNonBreaking)
	
	// Print breaking changes
	if len(breakingChanges) > 0 {
		fmt.Printf("🔴 Breaking Changes (%d):\n", len(breakingChanges))
		fmt.Println("========================")
		for _, change := range breakingChanges {
			printChange(change)
		}
		fmt.Println()
	}
	
	// Print dangerous changes
	if len(dangerousChanges) > 0 {
		fmt.Printf("🟡 Dangerous Changes (%d):\n", len(dangerousChanges))
		fmt.Println("=========================")
		for _, change := range dangerousChanges {
			printChange(change)
		}
		fmt.Println()
	}
	
	// Print non-breaking changes
	if len(nonBreakingChanges) > 0 {
		fmt.Printf("🟢 Non-Breaking Changes (%d):\n", len(nonBreakingChanges))
		fmt.Println("=============================")
		for _, change := range nonBreakingChanges {
			printChange(change)
		}
		fmt.Println()
	}
	
	// Check for failure conditions
	if viper.GetBool("diff.fail-on-breaking") && summary.Breaking > 0 {
		return fmt.Errorf("breaking changes detected")
	}
	
	if viper.GetBool("diff.fail-on-dangerous") && summary.Dangerous > 0 {
		return fmt.Errorf("dangerous changes detected")
	}
	
	return nil
}

func printChange(change core.Change) {
	icon := getChangeIcon(change.Type)
	fmt.Printf("  %s %s", icon, change.Message)
	if change.Path != "" {
		fmt.Printf(" (at %s)", change.Path)
	}
	fmt.Println()
}

func getChangeIcon(changeType core.ChangeType) string {
	switch changeType {
	case core.ChangeTypeBreaking:
		return "💥"
	case core.ChangeTypeDangerous:
		return "⚠️"
	case core.ChangeTypeNonBreaking:
		return "✨"
	default:
		return "•"
	}
}

func filterChangesByType(changes []core.Change, changeType core.ChangeType) []core.Change {
	var filtered []core.Change
	for _, change := range changes {
		if change.Type == changeType {
			filtered = append(filtered, change)
		}
	}
	return filtered
}

func calculateDiffSummary(changes []core.Change) DiffSummary {
	summary := DiffSummary{}
	
	for _, change := range changes {
		switch change.Type {
		case core.ChangeTypeBreaking:
			summary.Breaking++
		case core.ChangeTypeDangerous:
			summary.Dangerous++
		case core.ChangeTypeNonBreaking:
			summary.NonBreaking++
		}
	}
	
	return summary
}

type DiffSummary struct {
	Breaking    int `json:"breaking"`
	Dangerous   int `json:"dangerous"`
	NonBreaking int `json:"nonBreaking"`
} 