package repository

import (
	"strings"
	"testing"

	repo "github.com/jfeliu007/goplantuml/pkg/client/repository"
	"github.com/jfeliu007/goplantuml/test"
)

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestPlantUMLRepository_GenerateDiagram_BasicStruct(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Create test Go files
	err := testHelper.CreateTestDirectory("testpkg")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	structContent := `package testpkg

type Person struct {
	Name string
	Age  int
}

func (p *Person) GetName() string {
	return p.Name
}

func (p *Person) SetAge(age int) {
	p.Age = age
}
`
	err = testHelper.CreateTestGoFile("testpkg/person.go", structContent)
	if err != nil {
		t.Fatalf("Failed to create test struct file: %v", err)
	}

	// Create repository with memory filesystem
	plantUMLRepo := repo.NewPlantUMLRepositoryWithFS(testHelper.FileSystem)

	// Test data
	directories := []string{"testpkg"}
	ignoredDirs := []string{}
	recursive := false
	options := repo.DiagramOptions{
		ShowAggregations:    false,
		HideFields:          false,
		HideMethods:         false,
		ShowCompositions:    true,
		ShowImplementations: true,
		Title:               "Test Diagram",
	}

	// Execute
	result, err := plantUMLRepo.GenerateDiagram(directories, ignoredDirs, recursive, options)

	// Assert
	if err != nil {
		t.Fatalf("GenerateDiagram failed: %v", err)
	}

	if result == "" {
		t.Fatal("Expected non-empty diagram result")
	}

	// Check for PlantUML format
	if !contains(result, "@startuml") {
		t.Error("Expected PlantUML to start with @startuml")
	}

	if !contains(result, "@enduml") {
		t.Error("Expected PlantUML to end with @enduml")
	}

	// Check for struct content
	if !contains(result, "Person") {
		t.Error("Expected Person struct in diagram")
	}

	// Check for title
	if !contains(result, "Test Diagram") {
		t.Error("Expected title in diagram")
	}
}

func TestPlantUMLRepository_GenerateDiagram_WithInterface(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Create test directory
	err := testHelper.CreateTestDirectory("testpkg")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create interface file
	interfaceContent := `package testpkg

type Animal interface {
	Speak() string
	Move() bool
}
`
	err = testHelper.CreateTestGoFile("testpkg/animal.go", interfaceContent)
	if err != nil {
		t.Fatalf("Failed to create test interface file: %v", err)
	}

	// Create implementation file
	dogContent := `package testpkg

type Dog struct {
	Name string
	Breed string
}

func (d *Dog) Speak() string {
	return "Woof!"
}

func (d *Dog) Move() bool {
	return true
}
`
	err = testHelper.CreateTestGoFile("testpkg/dog.go", dogContent)
	if err != nil {
		t.Fatalf("Failed to create test implementation file: %v", err)
	}

	// Create repository
	plantUMLRepo := repo.NewPlantUMLRepositoryWithFS(testHelper.FileSystem)

	// Test data
	directories := []string{"testpkg"}
	options := repo.DiagramOptions{
		ShowImplementations: true,
		ShowCompositions:    true,
		Title:               "Interface Test Diagram",
	}

	// Execute
	result, err := plantUMLRepo.GenerateDiagram(directories, []string{}, false, options)

	// Assert
	if err != nil {
		t.Fatalf("GenerateDiagram failed: %v", err)
	}

	if result == "" {
		t.Fatal("Expected non-empty diagram result")
	}

	// Check for interface and implementation
	if !contains(result, "Animal") {
		t.Error("Expected Animal interface in diagram")
	}

	if !contains(result, "Dog") {
		t.Error("Expected Dog struct in diagram")
	}

	// Check for methods
	if !contains(result, "Speak") {
		t.Error("Expected Speak method in diagram")
	}

	if !contains(result, "Move") {
		t.Error("Expected Move method in diagram")
	}
}

func TestPlantUMLRepository_GenerateDiagram_WithRenderingOptions(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Create test directory and files
	err := testHelper.CreateTestDirectory("testpkg")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	structContent := `package testpkg

type TestStruct struct {
	PublicField  string
	privateField int
}

func (ts *TestStruct) PublicMethod() string {
	return ts.PublicField
}

func (ts *TestStruct) privateMethod() int {
	return ts.privateField
}
`
	err = testHelper.CreateTestGoFile("testpkg/teststruct.go", structContent)
	if err != nil {
		t.Fatalf("Failed to create test struct file: %v", err)
	}

	// Create repository
	plantUMLRepo := repo.NewPlantUMLRepositoryWithFS(testHelper.FileSystem)

	// Test with different rendering options
	testCases := []struct {
		name     string
		options  repo.DiagramOptions
		expected []string
	}{
		{
			name: "Hide fields",
			options: repo.DiagramOptions{
				HideFields: true,
				Title:      "Hide Fields Test",
			},
			expected: []string{"@startuml", "@enduml", "TestStruct", "hide fields"},
		},
		{
			name: "Hide methods",
			options: repo.DiagramOptions{
				HideMethods: true,
				Title:       "Hide Methods Test",
			},
			expected: []string{"@startuml", "@enduml", "TestStruct", "hide methods"},
		},
		{
			name: "With connection labels",
			options: repo.DiagramOptions{
				ShowConnectionLabels: true,
				Title:                "Connection Labels Test",
			},
			expected: []string{"@startuml", "@enduml", "TestStruct"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := plantUMLRepo.GenerateDiagram([]string{"testpkg"}, []string{}, false, tc.options)
			if err != nil {
				t.Fatalf("GenerateDiagram failed: %v", err)
			}

			// Check expected content
			for _, expected := range tc.expected {
				if !contains(result, expected) {
					t.Errorf("Expected '%s' in result, but not found", expected)
				}
			}
		})
	}
}

func TestPlantUMLRepository_ValidateDirectories(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Create repository
	plantUMLRepo := repo.NewPlantUMLRepositoryWithFS(testHelper.FileSystem)

	// Test cases
	testCases := []struct {
		name        string
		directories []string
		expectError bool
		setupFunc   func() error
	}{
		{
			name:        "Valid existing directory",
			directories: []string{"validdir"},
			expectError: false,
			setupFunc: func() error {
				return testHelper.CreateTestDirectory("validdir")
			},
		},
		{
			name:        "Non-existent directory",
			directories: []string{"nonexistent"},
			expectError: true,
			setupFunc:   func() error { return nil },
		},
		{
			name:        "Multiple directories - all valid",
			directories: []string{"dir1", "dir2"},
			expectError: false,
			setupFunc: func() error {
				if err := testHelper.CreateTestDirectory("dir1"); err != nil {
					return err
				}
				return testHelper.CreateTestDirectory("dir2")
			},
		},
		{
			name:        "Multiple directories - one invalid",
			directories: []string{"validdir2", "invaliddir"},
			expectError: true,
			setupFunc: func() error {
				return testHelper.CreateTestDirectory("validdir2")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			if err := tc.setupFunc(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Execute
			err := plantUMLRepo.ValidateDirectories(tc.directories)

			// Assert
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestPlantUMLRepository_AnalyzeCodeStructure(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Create test directory and files
	err := testHelper.CreateTestDirectory("analysis")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	structContent := `package analysis

type AnalysisStruct struct {
	Field1 string
	Field2 int
}

func (as *AnalysisStruct) Method1() string {
	return as.Field1
}

func (as *AnalysisStruct) Method2() int {
	return as.Field2
}
`
	err = testHelper.CreateTestGoFile("analysis/struct.go", structContent)
	if err != nil {
		t.Fatalf("Failed to create test struct file: %v", err)
	}

	interfaceContent := `package analysis

type AnalysisInterface interface {
	InterfaceMethod() bool
}
`
	err = testHelper.CreateTestGoFile("analysis/interface.go", interfaceContent)
	if err != nil {
		t.Fatalf("Failed to create test interface file: %v", err)
	}

	// Create repository
	plantUMLRepo := repo.NewPlantUMLRepositoryWithFS(testHelper.FileSystem)

	// Execute
	result, err := plantUMLRepo.AnalyzeCodeStructure([]string{"analysis"}, false)

	// Assert
	if err != nil {
		t.Fatalf("AnalyzeCodeStructure failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil analysis result")
	}

	if result.Summary == "" {
		t.Error("Expected non-empty summary")
	}

	// Check that analysis contains some meaningful information
	if !contains(result.Summary, "generated") && !contains(result.Summary, "characters") {
		t.Error("Expected summary to contain generation information")
	}
}
