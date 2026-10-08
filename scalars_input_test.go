package graphql_test

import (
	"testing"

	"github.com/graphql-go/graphql"
)

func TestScalarVariables_RejectMismatchingTypes(t *testing.T) {
	schema := scalarArgsSchema(t)
	tests := []struct {
		name    string
		varType string
		value   interface{}
		valid   bool
	}{
		{"string accepts string", "String", "a", true},
		{"string rejects number", "String", 1.0, false},
		{"string rejects bool", "String", true, false},
		{"string rejects object", "String", map[string]interface{}{"a": "b"}, false},
		{"string rejects list of non-strings", "String", []interface{}{1.0}, false},
		{"id accepts string", "ID", "a", true},
		{"id accepts integral number", "ID", 4.0, true},
		{"id rejects fractional number", "ID", 4.5, false},
		{"id rejects object", "ID", map[string]interface{}{}, false},
		{"boolean accepts bool", "Boolean", false, true},
		{"boolean rejects string", "Boolean", "false", false},
		{"boolean rejects number", "Boolean", 0.0, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := graphql.Do(graphql.Params{
				Schema:         schema,
				RequestString:  `query($v: ` + test.varType + `) { ` + test.varType + `(arg: $v) }`,
				VariableValues: map[string]interface{}{"v": test.value},
			})
			if valid := len(result.Errors) == 0; valid != test.valid {
				t.Errorf("expected valid=%v, got errors: %v", test.valid, result.Errors)
			}
		})
	}
}

func TestScalarVariables_DeeplyNestedValueIsRejected(t *testing.T) {
	var value interface{} = "leaf"
	for i := 0; i < 5_000_000; i++ {
		value = []interface{}{value}
	}
	result := graphql.Do(graphql.Params{
		Schema:         scalarArgsSchema(t),
		RequestString:  `query($v: String) { String(arg: $v) }`,
		VariableValues: map[string]interface{}{"v": value},
	})
	if len(result.Errors) == 0 {
		t.Fatal("expected an error")
	}
}

func scalarArgsSchema(t *testing.T) graphql.Schema {
	t.Helper()
	fields := graphql.Fields{}
	for _, scalar := range []*graphql.Scalar{graphql.String, graphql.ID, graphql.Boolean} {
		fields[scalar.Name()] = &graphql.Field{
			Type:    graphql.String,
			Args:    graphql.FieldConfigArgument{"arg": &graphql.ArgumentConfig{Type: scalar}},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return "ok", nil },
		}
	}
	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: fields})})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}
