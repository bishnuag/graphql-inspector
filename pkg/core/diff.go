package core

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/graphql-go/graphql"
)

// DiffSchemas compares two GraphQL schemas and returns the differences
func DiffSchemas(oldSchema, newSchema *Schema, options *DiffOptions) ([]Change, error) {
	if oldSchema == nil || newSchema == nil {
		return nil, fmt.Errorf("both schemas must be provided")
	}

	if options == nil {
		options = &DiffOptions{}
	}

	var changes []Change

	// Compare types
	typeChanges := compareTypes(oldSchema.Schema, newSchema.Schema, options)
	changes = append(changes, typeChanges...)

	// Compare directives
	if !options.IgnoreDirectives {
		directiveChanges := compareDirectives(oldSchema.Schema, newSchema.Schema, options)
		changes = append(changes, directiveChanges...)
	}

	// Compare schema definition (query, mutation, subscription)
	schemaChanges := compareSchemaDefinition(oldSchema.Schema, newSchema.Schema, options)
	changes = append(changes, schemaChanges...)

	// Sort changes by criticality and path
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Type != changes[j].Type {
			return changes[i].Type == ChangeTypeBreaking
		}
		return changes[i].Path < changes[j].Path
	})

	return changes, nil
}

// compareTypes compares types between two schemas
func compareTypes(oldSchema, newSchema *graphql.Schema, options *DiffOptions) []Change {
	var changes []Change

	oldTypes := oldSchema.TypeMap()
	newTypes := newSchema.TypeMap()

	// Find removed types
	for name, oldType := range oldTypes {
		if _, exists := newTypes[name]; !exists {
			changes = append(changes, Change{
				Type:        ChangeTypeBreaking,
				Message:     fmt.Sprintf("Type '%s' was removed", name),
				Path:        name,
				Criticality: "HIGH",
				Meta: map[string]interface{}{
					"typeName": name,
					"typeKind": getTypeKind(oldType),
				},
			})
		}
	}

	// Find added types
	for name, newType := range newTypes {
		if _, exists := oldTypes[name]; !exists {
			changes = append(changes, Change{
				Type:        ChangeTypeNonBreaking,
				Message:     fmt.Sprintf("Type '%s' was added", name),
				Path:        name,
				Criticality: "LOW",
				Meta: map[string]interface{}{
					"typeName": name,
					"typeKind": getTypeKind(newType),
				},
			})
		}
	}

	// Find modified types
	for name, oldType := range oldTypes {
		if newType, exists := newTypes[name]; exists {
			typeChanges := compareType(name, oldType, newType, options)
			changes = append(changes, typeChanges...)
		}
	}

	return changes
}

// compareType compares a specific type between schemas
func compareType(typeName string, oldType, newType graphql.Type, options *DiffOptions) []Change {
	var changes []Change

	// Check if type kind changed
	if getTypeKind(oldType) != getTypeKind(newType) {
		changes = append(changes, Change{
			Type:        ChangeTypeBreaking,
			Message:     fmt.Sprintf("Type '%s' changed from %s to %s", typeName, getTypeKind(oldType), getTypeKind(newType)),
			Path:        typeName,
			Criticality: "HIGH",
			Meta: map[string]interface{}{
				"typeName": typeName,
				"oldKind":  getTypeKind(oldType),
				"newKind":  getTypeKind(newType),
			},
		})
		return changes
	}

	// Handle different type kinds
	switch oldType := oldType.(type) {
	case *graphql.Object:
		if newType, ok := newType.(*graphql.Object); ok {
			changes = append(changes, compareObjectType(typeName, oldType, newType, options)...)
		}
	case *graphql.Interface:
		if newType, ok := newType.(*graphql.Interface); ok {
			changes = append(changes, compareInterfaceType(typeName, oldType, newType, options)...)
		}
	case *graphql.Union:
		if newType, ok := newType.(*graphql.Union); ok {
			changes = append(changes, compareUnionType(typeName, oldType, newType, options)...)
		}
	case *graphql.Enum:
		if newType, ok := newType.(*graphql.Enum); ok {
			changes = append(changes, compareEnumType(typeName, oldType, newType, options)...)
		}
	case *graphql.InputObject:
		if newType, ok := newType.(*graphql.InputObject); ok {
			changes = append(changes, compareInputObjectType(typeName, oldType, newType, options)...)
		}
	case *graphql.Scalar:
		if newType, ok := newType.(*graphql.Scalar); ok {
			changes = append(changes, compareScalarType(typeName, oldType, newType, options)...)
		}
	}

	return changes
}

// compareObjectType compares object types
func compareObjectType(typeName string, oldType, newType *graphql.Object, options *DiffOptions) []Change {
	var changes []Change

	// Compare description
	if !options.IgnoreDescriptions && oldType.Description() != newType.Description() {
		changes = append(changes, Change{
			Type:        ChangeTypeNonBreaking,
			Message:     fmt.Sprintf("Description for type '%s' changed", typeName),
			Path:        typeName,
			Criticality: "LOW",
		})
	}

	// Compare fields
	fieldChanges := compareFields(typeName, oldType.Fields(), newType.Fields(), options)
	changes = append(changes, fieldChanges...)

	// Compare interfaces
	interfaceChanges := compareImplementedInterfaces(typeName, oldType.Interfaces(), newType.Interfaces(), options)
	changes = append(changes, interfaceChanges...)

	return changes
}

// compareFields compares fields between types
func compareFields(typeName string, oldFields, newFields graphql.FieldDefinitionMap, options *DiffOptions) []Change {
	var changes []Change

	// Find removed fields
	for fieldName := range oldFields {
		if _, exists := newFields[fieldName]; !exists {
			changes = append(changes, Change{
				Type:        ChangeTypeBreaking,
				Message:     fmt.Sprintf("Field '%s.%s' was removed", typeName, fieldName),
				Path:        fmt.Sprintf("%s.%s", typeName, fieldName),
				Criticality: "HIGH",
				Meta: map[string]interface{}{
					"typeName":  typeName,
					"fieldName": fieldName,
				},
			})
		}
	}

	// Find added fields
	for fieldName := range newFields {
		if _, exists := oldFields[fieldName]; !exists {
			changes = append(changes, Change{
				Type:        ChangeTypeNonBreaking,
				Message:     fmt.Sprintf("Field '%s.%s' was added", typeName, fieldName),
				Path:        fmt.Sprintf("%s.%s", typeName, fieldName),
				Criticality: "LOW",
				Meta: map[string]interface{}{
					"typeName":  typeName,
					"fieldName": fieldName,
				},
			})
		}
	}

	// Find modified fields
	for fieldName, oldField := range oldFields {
		if newField, exists := newFields[fieldName]; exists {
			fieldChanges := compareField(typeName, fieldName, oldField, newField, options)
			changes = append(changes, fieldChanges...)
		}
	}

	return changes
}

// compareField compares a specific field
func compareField(typeName, fieldName string, oldField, newField *graphql.FieldDefinition, options *DiffOptions) []Change {
	var changes []Change

	// Compare field type
	if !areTypesEqual(oldField.Type, newField.Type) {
		criticality := "HIGH"
		changeType := ChangeTypeBreaking

		// Check if change is safe (widening)
		if isTypeWidening(oldField.Type, newField.Type) {
			criticality = "MEDIUM"
			changeType = ChangeTypeDangerous
		}

		changes = append(changes, Change{
			Type:        changeType,
			Message:     fmt.Sprintf("Field '%s.%s' changed type from %s to %s", typeName, fieldName, getTypeString(oldField.Type), getTypeString(newField.Type)),
			Path:        fmt.Sprintf("%s.%s", typeName, fieldName),
			Criticality: criticality,
			Meta: map[string]interface{}{
				"typeName":  typeName,
				"fieldName": fieldName,
				"oldType":   getTypeString(oldField.Type),
				"newType":   getTypeString(newField.Type),
			},
		})
	}

	// Compare field description
	if !options.IgnoreDescriptions && oldField.Description != newField.Description {
		changes = append(changes, Change{
			Type:        ChangeTypeNonBreaking,
			Message:     fmt.Sprintf("Field '%s.%s' description changed", typeName, fieldName),
			Path:        fmt.Sprintf("%s.%s", typeName, fieldName),
			Criticality: "LOW",
		})
	}

	// Compare field arguments
	argChanges := compareFieldArguments(typeName, fieldName, oldField.Args, newField.Args, options)
	changes = append(changes, argChanges...)

	return changes
}

// compareFieldArguments compares field arguments
func compareFieldArguments(typeName, fieldName string, oldArgs, newArgs []*graphql.Argument, options *DiffOptions) []Change {
	var changes []Change

	// Create maps for easier comparison
	oldArgMap := make(map[string]*graphql.Argument)
	newArgMap := make(map[string]*graphql.Argument)
	
	for _, arg := range oldArgs {
		oldArgMap[arg.Name()] = arg
	}
	
	for _, arg := range newArgs {
		newArgMap[arg.Name()] = arg
	}

	// Find removed arguments
	for argName := range oldArgMap {
		if _, exists := newArgMap[argName]; !exists {
			changes = append(changes, Change{
				Type:        ChangeTypeBreaking,
				Message:     fmt.Sprintf("Argument '%s' was removed from field '%s.%s'", argName, typeName, fieldName),
				Path:        fmt.Sprintf("%s.%s(%s:)", typeName, fieldName, argName),
				Criticality: "HIGH",
				Meta: map[string]interface{}{
					"typeName":  typeName,
					"fieldName": fieldName,
					"argName":   argName,
				},
			})
		}
	}

	// Find added arguments
	for argName, newArg := range newArgMap {
		if _, exists := oldArgMap[argName]; !exists {
			changeType := ChangeTypeNonBreaking
			criticality := "LOW"

			// Check if the new argument is required
			if isRequiredType(newArg.Type) {
				changeType = ChangeTypeBreaking
				criticality = "HIGH"
			}

			changes = append(changes, Change{
				Type:        changeType,
				Message:     fmt.Sprintf("Argument '%s' was added to field '%s.%s'", argName, typeName, fieldName),
				Path:        fmt.Sprintf("%s.%s(%s:)", typeName, fieldName, argName),
				Criticality: criticality,
				Meta: map[string]interface{}{
					"typeName":  typeName,
					"fieldName": fieldName,
					"argName":   argName,
					"argType":   getTypeString(newArg.Type),
				},
			})
		}
	}

	return changes
}

// Helper functions

func getTypeKind(t graphql.Type) string {
	switch t.(type) {
	case *graphql.Object:
		return "OBJECT"
	case *graphql.Interface:
		return "INTERFACE"
	case *graphql.Union:
		return "UNION"
	case *graphql.Enum:
		return "ENUM"
	case *graphql.InputObject:
		return "INPUT_OBJECT"
	case *graphql.Scalar:
		return "SCALAR"
	case *graphql.List:
		return "LIST"
	case *graphql.NonNull:
		return "NON_NULL"
	default:
		return "UNKNOWN"
	}
}

func getTypeString(t graphql.Type) string {
	switch t := t.(type) {
	case *graphql.NonNull:
		return getTypeString(t.OfType) + "!"
	case *graphql.List:
		return "[" + getTypeString(t.OfType) + "]"
	case *graphql.Object:
		return t.Name()
	case *graphql.Scalar:
		return t.Name()
	case *graphql.Enum:
		return t.Name()
	case *graphql.Interface:
		return t.Name()
	case *graphql.Union:
		return t.Name()
	case *graphql.InputObject:
		return t.Name()
	default:
		return "Unknown"
	}
}

func areTypesEqual(oldType, newType graphql.Type) bool {
	return reflect.DeepEqual(oldType, newType)
}

func isTypeWidening(oldType, newType graphql.Type) bool {
	// Check if changing from non-null to nullable (widening)
	if oldNonNull, ok := oldType.(*graphql.NonNull); ok {
		if newNonNull, ok := newType.(*graphql.NonNull); !ok {
			return areTypesEqual(oldNonNull.OfType, newType)
		} else {
			return isTypeWidening(oldNonNull.OfType, newNonNull.OfType)
		}
	}

	// Check if changing from specific type to union containing that type
	// This is a simplified check - in practice, this would be more complex
	return false
}

func isRequiredType(t graphql.Type) bool {
	_, ok := t.(*graphql.NonNull)
	return ok
}

// Placeholder functions for other type comparisons
func compareInterfaceType(typeName string, oldType, newType *graphql.Interface, options *DiffOptions) []Change {
	// TODO: Implement interface comparison
	return []Change{}
}

func compareUnionType(typeName string, oldType, newType *graphql.Union, options *DiffOptions) []Change {
	// TODO: Implement union comparison
	return []Change{}
}

func compareEnumType(typeName string, oldType, newType *graphql.Enum, options *DiffOptions) []Change {
	// TODO: Implement enum comparison
	return []Change{}
}

func compareInputObjectType(typeName string, oldType, newType *graphql.InputObject, options *DiffOptions) []Change {
	// TODO: Implement input object comparison
	return []Change{}
}

func compareScalarType(typeName string, oldType, newType *graphql.Scalar, options *DiffOptions) []Change {
	// TODO: Implement scalar comparison
	return []Change{}
}

func compareDirectives(oldSchema, newSchema *graphql.Schema, options *DiffOptions) []Change {
	// TODO: Implement directive comparison
	return []Change{}
}

func compareSchemaDefinition(oldSchema, newSchema *graphql.Schema, options *DiffOptions) []Change {
	// TODO: Implement schema definition comparison
	return []Change{}
}

func compareImplementedInterfaces(typeName string, oldInterfaces, newInterfaces []*graphql.Interface, options *DiffOptions) []Change {
	// TODO: Implement interface implementation comparison
	return []Change{}
} 