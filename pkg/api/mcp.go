package api

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

const mcpIntrospectionQuery = `query LinctlMCPIntrospection {
  __schema {
    queryType { name }
    mutationType { name }
    types {
      kind
      name
      description
      fields(includeDeprecated: true) {
        name
        description
        args {
          name
          description
          type { ...TypeRef }
          defaultValue
        }
        type { ...TypeRef }
      }
      inputFields {
        name
        description
        type { ...TypeRef }
        defaultValue
      }
    }
  }
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
            }
          }
        }
      }
    }
  }
}`

type MCPTool struct {
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	SourceField      string   `json:"sourceField"`
	Description      string   `json:"description,omitempty"`
	Args             []MCPArg `json:"args,omitempty"`
	ReturnType       string   `json:"returnType"`
	ReturnNamedType  string   `json:"returnNamedType"`
	ReturnNamedKind  string   `json:"returnNamedKind"`
	DefaultSelection string   `json:"defaultSelection,omitempty"`
}

type MCPArg struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
}

type introspectionResponse struct {
	Schema struct {
		QueryType struct {
			Name string `json:"name"`
		} `json:"queryType"`
		MutationType *struct {
			Name string `json:"name"`
		} `json:"mutationType"`
		Types []introspectionType `json:"types"`
	} `json:"__schema"`
}

type introspectionType struct {
	Kind        string               `json:"kind"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Fields      []introspectionField `json:"fields"`
	InputFields []introspectionInput `json:"inputFields"`
}

type introspectionField struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Args        []introspectionArg `json:"args"`
	Type        *introspectionRef  `json:"type"`
}

type introspectionInput struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Type         *introspectionRef `json:"type"`
	DefaultValue *string           `json:"defaultValue"`
}

type introspectionArg struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Type         *introspectionRef `json:"type"`
	DefaultValue *string           `json:"defaultValue"`
}

type introspectionRef struct {
	Kind   string            `json:"kind"`
	Name   string            `json:"name"`
	OfType *introspectionRef `json:"ofType"`
}

func (c *Client) DiscoverMCPTools(ctx context.Context) ([]MCPTool, error) {
	var payload introspectionResponse
	if err := c.Execute(ctx, mcpIntrospectionQuery, nil, &payload); err != nil {
		return nil, err
	}

	typesByName := make(map[string]introspectionType, len(payload.Schema.Types))
	for _, t := range payload.Schema.Types {
		if strings.TrimSpace(t.Name) == "" {
			continue
		}
		typesByName[t.Name] = t
	}

	var tools []MCPTool
	if qt := strings.TrimSpace(payload.Schema.QueryType.Name); qt != "" {
		tools = append(tools, buildToolsForRootType("query", qt, typesByName)...)
	}
	if payload.Schema.MutationType != nil {
		if mt := strings.TrimSpace(payload.Schema.MutationType.Name); mt != "" {
			tools = append(tools, buildToolsForRootType("mutation", mt, typesByName)...)
		}
	}

	sort.Slice(tools, func(i, j int) bool {
		if tools[i].Kind != tools[j].Kind {
			return tools[i].Kind < tools[j].Kind
		}
		return tools[i].Name < tools[j].Name
	})
	return tools, nil
}

func BuildMCPCall(tool MCPTool, arguments map[string]interface{}, customSelection string) (string, map[string]interface{}, error) {
	if arguments == nil {
		arguments = map[string]interface{}{}
	}

	allowed := map[string]MCPArg{}
	for _, arg := range tool.Args {
		allowed[arg.Name] = arg
		if arg.Required {
			if _, ok := arguments[arg.Name]; !ok {
				return "", nil, fmt.Errorf("missing required argument %q", arg.Name)
			}
		}
	}
	for key := range arguments {
		if _, ok := allowed[key]; !ok {
			return "", nil, fmt.Errorf("unknown argument %q", key)
		}
	}

	varDefs := make([]string, 0, len(tool.Args))
	callArgs := make([]string, 0, len(tool.Args))
	vars := make(map[string]interface{}, len(arguments))
	for _, arg := range tool.Args {
		if value, ok := arguments[arg.Name]; ok {
			varDefs = append(varDefs, fmt.Sprintf("$%s: %s", arg.Name, arg.Type))
			callArgs = append(callArgs, fmt.Sprintf("%s: $%s", arg.Name, arg.Name))
			vars[arg.Name] = value
		}
	}

	selection := strings.TrimSpace(customSelection)
	if selection == "" {
		selection = strings.TrimSpace(tool.DefaultSelection)
	}
	needsSelection := tool.ReturnNamedKind == "OBJECT" || tool.ReturnNamedKind == "INTERFACE" || tool.ReturnNamedKind == "UNION"
	if needsSelection {
		if selection == "" {
			selection = "{ __typename }"
		} else {
			selection = normalizeSelection(selection)
		}
	} else if selection != "" {
		return "", nil, fmt.Errorf("tool %q returns %s and does not accept a selection set", tool.Name, tool.ReturnNamedKind)
	}

	opKeyword := strings.TrimSpace(tool.Kind)
	if opKeyword == "" {
		opKeyword = "query"
	}

	rootField := strings.TrimSpace(tool.SourceField)
	if rootField == "" {
		rootField = strings.TrimPrefix(tool.Name, opKeyword+".")
	}

	var sb strings.Builder
	sb.WriteString(opKeyword)
	sb.WriteString(" LinctlMCPCall")
	if len(varDefs) > 0 {
		sb.WriteString("(")
		sb.WriteString(strings.Join(varDefs, ", "))
		sb.WriteString(")")
	}
	sb.WriteString(" { ")
	sb.WriteString(rootField)
	if len(callArgs) > 0 {
		sb.WriteString("(")
		sb.WriteString(strings.Join(callArgs, ", "))
		sb.WriteString(")")
	}
	if selection != "" {
		sb.WriteString(" ")
		sb.WriteString(selection)
	}
	sb.WriteString(" }")

	return sb.String(), vars, nil
}

func buildToolsForRootType(kind, rootTypeName string, typesByName map[string]introspectionType) []MCPTool {
	rootType, ok := typesByName[rootTypeName]
	if !ok {
		return nil
	}

	tools := make([]MCPTool, 0, len(rootType.Fields))
	for _, field := range rootType.Fields {
		fieldName := strings.TrimSpace(field.Name)
		if fieldName == "" || strings.HasPrefix(fieldName, "__") {
			continue
		}

		args := make([]MCPArg, 0, len(field.Args))
		for _, arg := range field.Args {
			typeName := formatTypeRef(arg.Type)
			if typeName == "" {
				continue
			}
			args = append(args, MCPArg{
				Name:        arg.Name,
				Description: strings.TrimSpace(arg.Description),
				Type:        typeName,
				Required:    isNonNull(arg.Type),
			})
		}

		namedKind, namedType := unwrapNamedType(field.Type)
		tools = append(tools, MCPTool{
			Name:             kind + "." + fieldName,
			Kind:             kind,
			SourceField:      fieldName,
			Description:      strings.TrimSpace(field.Description),
			Args:             args,
			ReturnType:       formatTypeRef(field.Type),
			ReturnNamedType:  namedType,
			ReturnNamedKind:  namedKind,
			DefaultSelection: buildDefaultSelection(field.Type, typesByName, 1),
		})
	}

	return tools
}

func normalizeSelection(selection string) string {
	trimmed := strings.TrimSpace(selection)
	if strings.HasPrefix(trimmed, "{") {
		return trimmed
	}
	return "{ " + trimmed + " }"
}

func buildDefaultSelection(ref *introspectionRef, typesByName map[string]introspectionType, depth int) string {
	kind, name := unwrapNamedType(ref)
	switch kind {
	case "SCALAR", "ENUM":
		return ""
	case "UNION":
		return "{ __typename }"
	case "OBJECT", "INTERFACE":
		if strings.TrimSpace(name) == "" {
			return "{ __typename }"
		}
		selection := buildSelectionForNamedType(name, typesByName, depth, map[string]bool{})
		if selection == "" {
			return "{ __typename }"
		}
		return selection
	default:
		return ""
	}
}

func buildSelectionForNamedType(typeName string, typesByName map[string]introspectionType, depth int, stack map[string]bool) string {
	if stack[typeName] {
		return "{ __typename }"
	}
	t, ok := typesByName[typeName]
	if !ok {
		return "{ __typename }"
	}
	if t.Kind == "UNION" {
		return "{ __typename }"
	}
	if t.Kind != "OBJECT" && t.Kind != "INTERFACE" {
		return ""
	}

	stack[typeName] = true
	defer delete(stack, typeName)

	parts := []string{"__typename"}
	added := 0
	for _, field := range t.Fields {
		if strings.HasPrefix(field.Name, "__") || strings.TrimSpace(field.Name) == "" {
			continue
		}
		if hasRequiredArguments(field.Args) {
			continue
		}
		fieldKind, _ := unwrapNamedType(field.Type)
		switch fieldKind {
		case "SCALAR", "ENUM":
			parts = append(parts, field.Name)
			added++
		case "OBJECT", "INTERFACE", "UNION":
			if depth > 0 {
				parts = append(parts, field.Name+" { __typename }")
				added++
			}
		}
		if added >= 24 {
			break
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return "{ " + strings.Join(parts, " ") + " }"
}

func hasRequiredArguments(args []introspectionArg) bool {
	for _, arg := range args {
		if isNonNull(arg.Type) && arg.DefaultValue == nil {
			return true
		}
	}
	return false
}

func isNonNull(ref *introspectionRef) bool {
	return ref != nil && ref.Kind == "NON_NULL"
}

func unwrapNamedType(ref *introspectionRef) (string, string) {
	for ref != nil {
		if strings.TrimSpace(ref.Name) != "" {
			return ref.Kind, ref.Name
		}
		ref = ref.OfType
	}
	return "", ""
}

func formatTypeRef(ref *introspectionRef) string {
	if ref == nil {
		return ""
	}
	switch ref.Kind {
	case "NON_NULL":
		inner := formatTypeRef(ref.OfType)
		if inner == "" {
			return ""
		}
		return inner + "!"
	case "LIST":
		inner := formatTypeRef(ref.OfType)
		if inner == "" {
			return ""
		}
		return "[" + inner + "]"
	default:
		return strings.TrimSpace(ref.Name)
	}
}
