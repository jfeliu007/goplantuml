package controller

import (
	"strings"
	"testing"

	"github.com/jfeliu007/goplantuml/pkg/client/repository"
	"github.com/jfeliu007/goplantuml/test"
)

func TestController_DataDrivenTests(t *testing.T) {
	testCases := []struct {
		name              string
		dataFile          string
		directory         string
		options           repository.DiagramOptions
		expectedInDiagram []string
	}{
		{
			name:      "ConnectionLabels",
			dataFile:  "go/connectionlabels/connectionlabels.go",
			directory: "connectionlabels",
			options: repository.DiagramOptions{
				ShowConnectionLabels: true,
				ShowImplementations:  true,
				ShowCompositions:     true,
			},
			expectedInDiagram: []string{
				"AbstractInterface",
				"ImplementsAbstractInterface",
				"AliasOfInt",
			},
		},
		{
			name:      "NamedImports",
			dataFile:  "go/namedimports/namedimports.go",
			directory: "namedimports",
			options: repository.DiagramOptions{
				ShowAliases: true,
			},
			expectedInDiagram: []string{
				"MyTypeWithNamedImport",
				"ProcessTimeWithNamedImport",
			},
		},
		{
			name:      "RenderingOptions",
			dataFile:  "go/renderingoptions/teststruct.go",
			directory: "renderingoptions",
			options: repository.DiagramOptions{
				HidePrivateMembers:      false,
				AggregatePrivateMembers: true,
			},
			expectedInDiagram: []string{
				"TestStruct",
				"PublicField",
				"PublicMethod",
			},
		},
		{
			name:      "ParenthesizedTypes",
			dataFile:  "go/parenthesizedtypes/types.go",
			directory: "parenthesizedtypes",
			options:   repository.DiagramOptions{},
			expectedInDiagram: []string{
				"interface Foo",
				"interface Bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			helper := test.NewTestHelper()
			defer helper.Cleanup()

			// Load test data from file
			testContent, err := helper.LoadTestDataFile(tc.dataFile)
			if err != nil {
				t.Fatalf("Failed to load test data for %s: %v", tc.name, err)
			}

			// Create test directory and file
			err = helper.CreateTestDirectory(tc.directory)
			if err != nil {
				t.Fatalf("Failed to create test directory for %s: %v", tc.name, err)
			}

			filename := tc.directory + "/" + tc.directory + ".go"
			err = helper.CreateTestGoFile(filename, testContent)
			if err != nil {
				t.Fatalf("Failed to create test file for %s: %v", tc.name, err)
			}

			// Create repository with test filesystem
			repo := repository.NewPlantUMLRepositoryWithFS(helper.FileSystem)

			// Generate diagram
			result, err := repo.GenerateDiagram([]string{tc.directory}, []string{}, false, tc.options)
			if err != nil {
				t.Errorf("Failed to generate diagram for %s: %v", tc.name, err)
				return
			}

			if result == "" {
				t.Errorf("Generated PlantUML is empty for %s", tc.name)
				return
			}

			// Validate that expected elements are in the diagram
			for _, expected := range tc.expectedInDiagram {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected element '%s' not found in diagram for %s", expected, tc.name)
				}
			}

			t.Logf("Successfully generated diagram for %s", tc.name)
		})
	}
}
