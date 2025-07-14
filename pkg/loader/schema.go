package loader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bishnuag/graphql-inspector/pkg/core"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

// LoadSchemaFromContent loads a GraphQL schema from GQL content
func LoadSchemaFromContent(content string) (*core.Schema, error) {
	var err error

	// Parse and build the schema
	schema, err := buildSchemaFromSDL(content)
	if err != nil {
		return nil, fmt.Errorf("failed to build schema: %w", err)
	}

	// Create hash for the schema
	hash := createHash(content)

	return &core.Schema{
		Schema:    schema,
		SDL:       content,
		Hash:      hash,
		Timestamp: time.Now(),
	}, nil
}

// LoadSchema loads a GraphQL schema from various sources
func LoadSchema(source string) (*core.Schema, error) {
	var content string
	var err error

	// Determine the source type and load accordingly
	if isURL(source) {
		content, err = loadFromURL(source)
	} else if isFile(source) {
		content, err = loadFromFile(source)
	} else {
		// Assume it's SDL content
		content = source
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load schema from %s: %w", source, err)
	}

	// Parse and build the schema
	schema, err := buildSchemaFromSDL(content)
	if err != nil {
		return nil, fmt.Errorf("failed to build schema: %w", err)
	}

	// Create hash for the schema
	hash := createHash(content)

	return &core.Schema{
		Schema:    schema,
		SDL:       content,
		Hash:      hash,
		Source:    source,
		Timestamp: time.Now(),
	}, nil
}

// LoadDocument loads a GraphQL document from various sources
func LoadDocument(source string) (*core.Document, error) {
	var content string
	var err error

	if isURL(source) {
		content, err = loadFromURL(source)
	} else if isFile(source) {
		content, err = loadFromFile(source)
	} else {
		// Assume it's GraphQL content
		content = source
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load document from %s: %w", source, err)
	}

	// Parse the document
	docAST, err := parser.Parse(parser.ParseParams{
		Source: content,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	// Create hash for the document
	hash := createHash(content)

	return &core.Document{
		Source:  source,
		Content: content,
		AST:     docAST,
		Hash:    hash,
	}, nil
}

// LoadDocuments loads multiple GraphQL documents from a glob pattern
func LoadDocuments(pattern string) ([]core.Document, error) {
	var documents []core.Document

	// Handle glob patterns
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to expand glob pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			if isGraphQLFile(match) {
				doc, err := LoadDocument(match)
				if err != nil {
					// Log error but continue with other files
					fmt.Fprintf(os.Stderr, "Warning: failed to load document %s: %v\n", match, err)
					continue
				}
				documents = append(documents, *doc)
			}
		}
	} else {
		// Single file or directory
		if isDirectory(pattern) {
			// Load all GraphQL files in directory
			err := filepath.Walk(pattern, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() && isGraphQLFile(path) {
					doc, err := LoadDocument(path)
					if err != nil {
						// Log error but continue with other files
						fmt.Fprintf(os.Stderr, "Warning: failed to load document %s: %v\n", path, err)
						return nil
					}
					documents = append(documents, *doc)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to walk directory %s: %w", pattern, err)
			}
		} else {
			// Single file
			doc, err := LoadDocument(pattern)
			if err != nil {
				return nil, err
			}
			documents = append(documents, *doc)
		}
	}

	return documents, nil
}

// buildSchemaFromSDL builds a GraphQL schema from SDL
func buildSchemaFromSDL(sdl string) (*graphql.Schema, error) {
	// Parse the SDL
	doc, err := parser.Parse(parser.ParseParams{
		Source: sdl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse SDL: %w", err)
	}

	// TODO: Implement proper SDL to schema conversion
	// The graphql-go library doesn't have a direct schema_from_ast utility
	// A proper implementation would:
	// 1. Walk the AST and extract type definitions
	// 2. Build GraphQL types from the definitions
	// 3. Create the schema with proper types and resolvers
	//
	// For now, we'll create a basic schema to demonstrate the structure
	// In a production implementation, you would parse the AST and build the schema

	// Extract basic information from the parsed SDL (simplified)
	hasQuery := false
	hasMutation := false
	hasSubscription := false

	for _, def := range doc.Definitions {
		switch def := def.(type) {
		case *ast.SchemaDefinition:
			for _, opType := range def.OperationTypes {
				switch opType.Operation {
				case "query":
					hasQuery = true
				case "mutation":
					hasMutation = true
				case "subscription":
					hasSubscription = true
				}
			}
		}
	}

	// Create a basic schema config
	schemaConfig := graphql.SchemaConfig{}

	// Create query type (required)
	if hasQuery || len(doc.Definitions) > 0 {
		schemaConfig.Query = graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"hello": &graphql.Field{
					Type: graphql.String,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return "Hello from GraphQL Inspector!", nil
					},
				},
			},
		})
	}

	// Create mutation type if detected
	if hasMutation {
		schemaConfig.Mutation = graphql.NewObject(graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				"noop": &graphql.Field{
					Type: graphql.String,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return "noop", nil
					},
				},
			},
		})
	}

	// Create subscription type if detected
	if hasSubscription {
		schemaConfig.Subscription = graphql.NewObject(graphql.ObjectConfig{
			Name: "Subscription",
			Fields: graphql.Fields{
				"noop": &graphql.Field{
					Type: graphql.String,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return "noop", nil
					},
				},
			},
		})
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build schema: %w", err)
	}

	return &schema, nil
}

// isURL checks if a string is a URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// isFile checks if a string is a file path
func isFile(s string) bool {
	_, err := os.Stat(s)
	return err == nil
}

// isDirectory checks if a string is a directory path
func isDirectory(s string) bool {
	info, err := os.Stat(s)
	return err == nil && info.IsDir()
}

// isGraphQLFile checks if a file is a GraphQL file based on extension
func isGraphQLFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".graphql" || ext == ".gql" || ext == ".graphqls"
}

// loadFromURL loads content from a URL
func loadFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(content), nil
}

// loadFromFile loads content from a file
func loadFromFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// createHash creates a SHA256 hash of the content
func createHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// LoadSchemaFromIntrospection loads a schema from introspection result
func LoadSchemaFromIntrospection(introspectionResult map[string]interface{}) (*core.Schema, error) {
	// This is a simplified implementation
	// In a real implementation, you would convert the introspection result to a schema

	// For now, we'll return an error indicating this is not implemented
	return nil, fmt.Errorf("loading schema from introspection is not yet implemented")
}

// LoadSchemaFromEndpoint loads a schema from a GraphQL endpoint via introspection
func LoadSchemaFromEndpoint(endpoint string, headers map[string]string) (*core.Schema, error) {
	// Construct introspection query
	introspectionQuery := `
		query IntrospectionQuery {
			__schema {
				queryType { name }
				mutationType { name }
				subscriptionType { name }
				types {
					...FullType
				}
				directives {
					name
					description
					locations
					args {
						...InputValue
					}
				}
			}
		}

		fragment FullType on __Type {
			kind
			name
			description
			fields(includeDeprecated: true) {
				name
				description
				args {
					...InputValue
				}
				type {
					...TypeRef
				}
				isDeprecated
				deprecationReason
			}
			inputFields {
				...InputValue
			}
			interfaces {
				...TypeRef
			}
			enumValues(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
			}
			possibleTypes {
				...TypeRef
			}
		}

		fragment InputValue on __InputValue {
			name
			description
			type { ...TypeRef }
			defaultValue
		}

		fragment TypeRef on __Type {
			kind
			name
			ofType {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
									}
								}
							}
						}
					}
				}
			}
		}
	`

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(fmt.Sprintf(`{"query": %q}`, introspectionQuery)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// For now, we'll return an error indicating this is not fully implemented
	return nil, fmt.Errorf("loading schema from endpoint is not yet fully implemented")
}

// ValidateSchema validates a GraphQL schema
func ValidateSchema(schema *core.Schema) []error {
	if schema == nil || schema.Schema == nil {
		return []error{fmt.Errorf("schema is nil")}
	}

	// The graphql-go library doesn't have a ValidateSchema function
	// In a real implementation, you would validate the schema here
	// For now, we'll just return no errors
	return []error{}
}
