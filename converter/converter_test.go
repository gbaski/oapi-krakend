package converter

import (
	_ "embed"
	"fmt"
	"testing"
)

func TestConverter_Convert(t *testing.T) {

	type fields struct {
		Spec      string
		Includes  string
		Excludes  string
		OutputDir string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Test Petstore",
			fields: fields{
				Spec:      "https://petstore.swagger.io/v2/swagger.json",
				Includes:  "",
				Excludes:  "",
				OutputDir: "/tmp/oapi_codegen_test",
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			c := &Converter{
				Spec:      tt.fields.Spec,
				Includes:  tt.fields.Includes,
				Excludes:  tt.fields.Excludes,
				OutputDir: tt.fields.OutputDir,
			}

			err := c.Convert()

			if err != nil {
				fmt.Println("Updated Unit Tests")
			}

		})
	}
}
