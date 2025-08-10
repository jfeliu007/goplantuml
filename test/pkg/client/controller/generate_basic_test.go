package controller

import (
	"strings"
	"testing"

	"github.com/jfeliu007/goplantuml/pkg/client/usecase"
	"github.com/jfeliu007/goplantuml/pkg/config"
	"github.com/jfeliu007/goplantuml/pkg/entity/request"
	"github.com/jfeliu007/goplantuml/test"
)

func TestGenerateController_Generate_Success(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Create test Go files
	err := testHelper.CreateTestDirectory("testpkg")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err = testHelper.CreateTestStructFile("testpkg/user.go", "testpkg", "User", []string{
		"ID   int    `json:\"id\"`",
		"Name string `json:\"name\"`",
	})
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create configuration
	baseConfig := &config.BaseConfig{}

	// Create usecase and controller (note: this would need proper DI in real implementation)
	uc := usecase.NewPlantUMLUsecase(baseConfig)

	// Test data
	req := request.GenerateRequest{
		Directories:         []string{"testpkg"},
		IgnoredDirectories:  []string{},
		Recursive:           false,
		OutputPath:          "",
		ShowAggregations:    true,
		HideFields:          false,
		HideMethods:         false,
		ShowCompositions:    true,
		ShowImplementations: true,
		Title:               "Test Controller Diagram",
	}

	// Execute
	result, err := uc.Generate(req)

	// Assert
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == "" {
		t.Fatal("Expected non-empty result")
	}

	// Check that result contains expected elements
	expectedElements := []string{
		"@startuml",
		"User",
		"@enduml",
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Expected result to contain '%s'\nActual result:\n%s", element, result)
		}
	}
}

func TestGenerateController_Validate_InvalidDirectory(t *testing.T) {
	// Setup
	baseConfig := &config.BaseConfig{}
	uc := usecase.NewPlantUMLUsecase(baseConfig)

	// Test data with invalid directory
	req := request.ValidateRequest{
		Directories: []string{"nonexistent_directory"},
	}

	// Execute
	result := uc.ValidateDirectories(req)

	// Assert
	if result["valid"].(bool) {
		t.Error("Expected validation to fail for nonexistent directory")
	}

	errors := result["errors"].([]string)
	if len(errors) == 0 {
		t.Error("Expected validation errors for nonexistent directory")
	}

	// Check error message contains expected text
	errorFound := false
	for _, err := range errors {
		if strings.Contains(err, "nonexistent_directory") {
			errorFound = true
			break
		}
	}
	if !errorFound {
		t.Errorf("Expected error message to mention directory name. Errors: %v", errors)
	}
}

func TestGenerateController_Generate_WithOptions(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Create test structure with interface and implementation
	err := testHelper.CreateTestDirectory("optionspkg")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err = testHelper.CreateTestInterfaceFile("optionspkg/service.go", "optionspkg", "UserService", []string{
		"GetUser(id int) (*User, error)",
		"CreateUser(user *User) error",
	})
	if err != nil {
		t.Fatalf("Failed to create interface file: %v", err)
	}

	err = testHelper.CreateTestGoFile("optionspkg/impl.go", `package optionspkg

type User struct {
	ID   int    
	Name string 
}

type UserServiceImpl struct {
	users map[int]*User
}

func (s *UserServiceImpl) GetUser(id int) (*User, error) {
	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserServiceImpl) CreateUser(user *User) error {
	s.users[user.ID] = user
	return nil
}
`)
	if err != nil {
		t.Fatalf("Failed to create implementation file: %v", err)
	}

	baseConfig := &config.BaseConfig{}
	uc := usecase.NewPlantUMLUsecase(baseConfig)

	// Test with different rendering options
	testCases := []struct {
		name               string
		hideFields         bool
		hideMethods        bool
		showImplements     bool
		expectedElements   []string
		unexpectedElements []string
	}{
		{
			name:               "Show All",
			hideFields:         false,
			hideMethods:        false,
			showImplements:     true,
			expectedElements:   []string{"UserService", "UserServiceImpl", "GetUser", "CreateUser", "ID int", "Name string"},
			unexpectedElements: []string{},
		},
		{
			name:               "Hide Fields",
			hideFields:         true,
			hideMethods:        false,
			showImplements:     true,
			expectedElements:   []string{"UserService", "UserServiceImpl", "GetUser", "CreateUser"},
			unexpectedElements: []string{"ID int", "Name string"},
		},
		{
			name:               "Hide Methods",
			hideFields:         false,
			hideMethods:        true,
			showImplements:     true,
			expectedElements:   []string{"UserService", "UserServiceImpl", "ID int", "Name string"},
			unexpectedElements: []string{"GetUser", "CreateUser"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := request.GenerateRequest{
				Directories:         []string{"optionspkg"},
				Recursive:           false,
				HideFields:          tc.hideFields,
				HideMethods:         tc.hideMethods,
				ShowImplementations: tc.showImplements,
				Title:               "Options Test: " + tc.name,
			}

			result, err := uc.Generate(req)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			// Check expected elements
			for _, expected := range tc.expectedElements {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s' in test '%s'", expected, tc.name)
				}
			}

			// Check unexpected elements (should not be present)
			for _, unexpected := range tc.unexpectedElements {
				if strings.Contains(result, unexpected) {
					t.Errorf("Expected result NOT to contain '%s' in test '%s'", unexpected, tc.name)
				}
			}
		})
	}
}
