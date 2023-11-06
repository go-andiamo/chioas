package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExamples_WriteYaml(t *testing.T) {
	testCases := []struct {
		examples  Examples
		expect    string
		expectErr bool
	}{
		{},
		{
			examples: Examples{
				{},
			},
			expectErr: true,
		},
		{
			examples: Examples{
				{
					Name:        "foo",
					Description: "foo desc",
					Summary:     "foo summary",
				},
			},
			expect: `examples:
  foo:
    summary: "foo summary"
    description: "foo desc"
    value: null
`,
		},
		{
			examples: Examples{
				{
					Name:        "foo",
					Description: "foo desc",
					Summary:     "foo summary",
					Value:       []string{"foo", "bar"},
				},
			},
			expect: `examples:
  foo:
    summary: "foo summary"
    description: "foo desc"
    value:
      - foo
      - bar
`,
		},
		{
			examples: Examples{
				{
					Name:        "foo",
					Description: "foo desc",
					Summary:     "foo summary",
					Value: struct {
						Foo string
						Bar int
					}{
						Foo: "foo val",
						Bar: 1,
					},
				},
			},
			expect: `examples:
  foo:
    summary: "foo summary"
    description: "foo desc"
    value:
      foo: foo val
      bar: 1
`,
		},
		{
			examples: Examples{
				{
					ExampleRef:  "foo",
					Name:        "test",
					Description: "won't see this",
				},
			},
			expect: `examples:
  test:
    $ref: "#/components/examples/foo"
`,
		},
		{
			examples: Examples{
				{
					Name: "foo",
				},
			},
			expect: `examples:
  foo:
    value: null
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.examples.writeYaml(w)
			data, err := w.Bytes()
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expect, string(data))
			}
		})
	}
}
