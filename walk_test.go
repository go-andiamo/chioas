package chioas

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestDefinition_WalkPaths(t *testing.T) {
	d := Definition{
		Paths: Paths{
			"/test": {
				Paths: Paths{
					"/sub": {},
				},
			},
		},
	}
	count := 0
	err := d.WalkPaths(func(path string, pathDef *Path) (bool, error) {
		pathDef.HideDocs = true
		count++
		return true, nil
	})
	require.NoError(t, err)
	require.Equal(t, 2, count)
	require.True(t, d.Paths["/test"].HideDocs)
	require.True(t, d.Paths["/test"].Paths["/sub"].HideDocs)
}

func TestDefinition_WalkMethods(t *testing.T) {
	d := Definition{
		Methods: Methods{
			http.MethodGet: {},
		},
		Paths: Paths{
			"/test": {
				Methods: Methods{
					http.MethodGet: {},
				},
				Paths: Paths{
					"/sub": {
						Methods: Methods{
							http.MethodGet: {},
						},
					},
				},
			},
		},
	}
	count := 0
	err := d.WalkMethods(func(path string, method string, methodDef *Method) (bool, error) {
		methodDef.HideDocs = true
		count++
		return true, nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, count)
	require.True(t, d.Methods["GET"].HideDocs)
	require.True(t, d.Paths["/test"].Methods["GET"].HideDocs)
	require.True(t, d.Paths["/test"].Paths["/sub"].Methods["GET"].HideDocs)
}
