package repository

import (
	"strings"
	"testing"

	repo "github.com/jfeliu007/goplantuml/pkg/client/repository"
	"github.com/jfeliu007/goplantuml/test"
)

func TestPlantUMLRepository_ConnectionLabels(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Load test data from file
	testContent, err := testHelper.LoadTestDataFile("go/connectionlabels/connectionlabels.go")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Create test directory and file
	err = testHelper.CreateTestDirectory("connectionlabels")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err = testHelper.CreateTestGoFile("connectionlabels/connectionlabels.go", testContent)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	plantUMLRepo := repo.NewPlantUMLRepositoryWithFS(testHelper.FileSystem)

	// Test with connection labels enabled
	options := repo.DiagramOptions{
		ShowConnectionLabels: true,
		ShowImplementations:  true,
		ShowCompositions:     true,
	}

	result, err := plantUMLRepo.GenerateDiagram([]string{"connectionlabels"}, []string{}, false, options)
	if err != nil {
		t.Fatalf("Failed to generate diagram: %v", err)
	}

	// Verify the diagram was generated
	if result == "" {
		t.Error("Generated diagram is empty")
	}

	// Verify it contains expected elements
	expectedElements := []string{
		"AbstractInterface",
		"ImplementsAbstractInterface",
		"AliasOfInt",
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Expected element '%s' not found in diagram", element)
		}
	}

	t.Logf("Successfully generated connection labels diagram")
}

func TestPlantUMLRepository_NamedImports(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Load test data from file
	testContent, err := testHelper.LoadTestDataFile("go/namedimports/namedimports.go")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Create test directory and file
	err = testHelper.CreateTestDirectory("namedimports")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err = testHelper.CreateTestGoFile("namedimports/namedimports.go", testContent)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repository
	plantUMLRepo := repo.NewPlantUMLRepositoryWithFS(testHelper.FileSystem)

	options := repo.DiagramOptions{
		ShowAliases: true,
	}

	result, err := plantUMLRepo.GenerateDiagram([]string{"namedimports"}, []string{}, false, options)
	if err != nil {
		t.Fatalf("Failed to generate diagram: %v", err)
	}

	// Verify the diagram was generated
	if result == "" {
		t.Error("Generated diagram is empty")
	}

	// Verify it contains expected elements
	expectedElements := []string{
		"MyTypeWithNamedImport",
		"ProcessTimeWithNamedImport",
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Expected element '%s' not found in diagram", element)
		}
	}

	t.Logf("Successfully generated named imports diagram")
}
