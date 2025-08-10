package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jfeliu007/goplantuml/pkg/config"
	"github.com/spf13/afero"
)

// TestHelper provides utilities for testing GoPlantUML
type TestHelper struct {
	FileSystem  afero.Fs
	TempDir     string
	BaseConfig  *config.BaseConfig
	YamlConfig  *config.Config
	TestDataDir string
	DataDir     string // Path to etc/test/data
}

// NewTestHelper creates a new test helper with memory filesystem
func NewTestHelper() *TestHelper {
	fs := afero.NewMemMapFs()

	baseConfig := &config.BaseConfig{}

	return &TestHelper{
		FileSystem:  fs,
		BaseConfig:  baseConfig,
		TestDataDir: "etc/test/data",
		DataDir:     "etc/test/data",
	}
}

// LoadTestDataFile reads content from etc/test/data
func (th *TestHelper) LoadTestDataFile(relativePath string) (string, error) {
	// Get the working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Navigate to project root by looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(workingDir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(workingDir)
		if parent == workingDir {
			return "", fmt.Errorf("could not find project root (go.mod not found)")
		}
		workingDir = parent
	}

	fullPath := filepath.Join(workingDir, th.DataDir, relativePath)
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read test data file %s: %w", fullPath, err)
	}

	return string(content), nil
}

// NewTestHelperWithRealFS creates a test helper with real filesystem for integration tests
func NewTestHelperWithRealFS(t *testing.T) *TestHelper {
	tempDir, err := os.MkdirTemp("", "goplantuml_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	fs := afero.NewOsFs()
	baseConfig := &config.BaseConfig{}

	return &TestHelper{
		FileSystem:  fs,
		TempDir:     tempDir,
		BaseConfig:  baseConfig,
		TestDataDir: "etc/test/data",
	}
}

// CreateTestGoFile creates a Go source file for testing
func (th *TestHelper) CreateTestGoFile(path, content string) error {
	return afero.WriteFile(th.FileSystem, path, []byte(content), 0644)
}

// CreateTestStructFile creates a test Go struct file
func (th *TestHelper) CreateTestStructFile(path, packageName, structName string, fields []string) error {
	content := "package " + packageName + "\n\n"
	content += "type " + structName + " struct {\n"
	for _, field := range fields {
		content += "\t" + field + "\n"
	}
	content += "}\n"

	return th.CreateTestGoFile(path, content)
}

// CreateTestInterfaceFile creates a test Go interface file
func (th *TestHelper) CreateTestInterfaceFile(path, packageName, interfaceName string, methods []string) error {
	content := "package " + packageName + "\n\n"
	content += "type " + interfaceName + " interface {\n"
	for _, method := range methods {
		content += "\t" + method + "\n"
	}
	content += "}\n"

	return th.CreateTestGoFile(path, content)
}

// CreateTestDirectory creates a directory for testing
func (th *TestHelper) CreateTestDirectory(path string) error {
	return th.FileSystem.MkdirAll(path, 0755)
}

// LoadYamlConfig loads a YAML configuration for testing
func (th *TestHelper) LoadYamlConfig(configPath string) error {
	yamlConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}
	th.YamlConfig = yamlConfig
	return nil
}

// CreateTestYamlConfig creates a test YAML configuration
func (th *TestHelper) CreateTestYamlConfig(path string, version string, directories []string) error {
	content := "version: " + version + "\n"
	content += "directories:\n"
	for _, dir := range directories {
		content += "  - " + dir + "\n"
	}
	content += "recursive: true\n"
	content += "rendering_options:\n"
	content += "  show_aggregations: true\n"
	content += "  show_compositions: true\n"
	content += "  show_implementations: true\n"

	return afero.WriteFile(th.FileSystem, path, []byte(content), 0644)
}

// Cleanup cleans up test resources
func (th *TestHelper) Cleanup() {
	if th.TempDir != "" {
		os.RemoveAll(th.TempDir)
	}
}

// AssertFileExists checks if a file exists
func (th *TestHelper) AssertFileExists(t *testing.T, path string) {
	exists, err := afero.Exists(th.FileSystem, path)
	if err != nil {
		t.Fatalf("Error checking file existence: %v", err)
	}
	if !exists {
		t.Fatalf("Expected file %s to exist", path)
	}
}

// AssertFileContent checks if file content matches expected
func (th *TestHelper) AssertFileContent(t *testing.T, path, expected string) {
	content, err := afero.ReadFile(th.FileSystem, path)
	if err != nil {
		t.Fatalf("Error reading file %s: %v", path, err)
	}
	if string(content) != expected {
		t.Fatalf("File content mismatch.\nExpected:\n%s\nActual:\n%s", expected, string(content))
	}
}

// GetTestDataPath returns the path to test data files
func (th *TestHelper) GetTestDataPath(filename string) string {
	return filepath.Join(th.TestDataDir, filename)
}

// CreateTestDataStructure creates comprehensive test data structure using external files
func (th *TestHelper) CreateTestDataStructure() error {
	// Create connection labels test data from file
	connectionLabelsContent, err := th.LoadTestDataFile("go/connectionlabels/connectionlabels.go")
	if err != nil {
		return fmt.Errorf("failed to load connectionlabels test data: %w", err)
	}

	if err := th.CreateTestDirectory("connectionlabels"); err != nil {
		return err
	}
	if err := th.CreateTestGoFile("connectionlabels/connectionlabels.go", connectionLabelsContent); err != nil {
		return err
	}

	// Create named imports test data from file
	namedImportsContent, err := th.LoadTestDataFile("go/namedimports/namedimports.go")
	if err != nil {
		return fmt.Errorf("failed to load namedimports test data: %w", err)
	}

	if err := th.CreateTestDirectory("namedimports"); err != nil {
		return err
	}
	if err := th.CreateTestGoFile("namedimports/namedimports.go", namedImportsContent); err != nil {
		return err
	}

	// Create rendering options test data from file
	renderingOptionsContent, err := th.LoadTestDataFile("go/renderingoptions/teststruct.go")
	if err != nil {
		return fmt.Errorf("failed to load renderingoptions test data: %w", err)
	}

	if err := th.CreateTestDirectory("renderingoptions"); err != nil {
		return err
	}
	if err := th.CreateTestGoFile("renderingoptions/teststruct.go", renderingOptionsContent); err != nil {
		return err
	}

	return nil
}

// CreateComplexTestScenario creates a complex test scenario with multiple packages
func (th *TestHelper) CreateComplexTestScenario() error {
	// Create main package with interface
	mainContent := `package main

type ProcessorInterface interface {
	Process(input string) (string, error)
	Validate(data map[string]interface{}) bool
}

type MainStruct struct {
	Name string
	ID   int
}`
	if err := th.CreateTestGoFile("main.go", mainContent); err != nil {
		return err
	}

	// Create implementation package
	if err := th.CreateTestDirectory("impl"); err != nil {
		return err
	}
	implContent := `package impl

import "main"

type ProcessorImpl struct {
	Config map[string]string
}

func (p *ProcessorImpl) Process(input string) (string, error) {
	return input, nil
}

func (p *ProcessorImpl) Validate(data map[string]interface{}) bool {
	return true
}`
	if err := th.CreateTestGoFile("impl/processor.go", implContent); err != nil {
		return err
	}

	return nil
}

// CreateParenthesizedTypesTestData creates test data for parenthesized type declarations from file
func (th *TestHelper) CreateParenthesizedTypesTestData() error {
	content, err := th.LoadTestDataFile("go/parenthesizedtypes/types.go")
	if err != nil {
		return fmt.Errorf("failed to load parenthesizedtypes test data: %w", err)
	}

	if err := th.CreateTestDirectory("parenthesizedtypes"); err != nil {
		return err
	}

	return th.CreateTestGoFile("parenthesizedtypes/types.go", content)
}

// CreateSubfolderTestData creates test data with subfolders
func (th *TestHelper) CreateSubfolderTestData() error {
	// Create subfolder with interfaces
	if err := th.CreateTestDirectory("subfolder"); err != nil {
		return err
	}

	subfolderContent := `package subfolder

type test2 interface {
	TestInterfaceAsField
	test()
}

// TestInterfaceAsField testing interface
type TestInterfaceAsField interface {
}`

	if err := th.CreateTestGoFile("subfolder/subfolder.go", subfolderContent); err != nil {
		return err
	}

	// Create subfolder2 with struct and methods
	if err := th.CreateTestDirectory("subfolder2"); err != nil {
		return err
	}

	subfolder2Content := `package subfolder2

// Subfolder2 structure for testing purpose only
type Subfolder2 struct {
}

// SubfolderFunction is for testing purposes
func (s *Subfolder2) SubfolderFunction(b bool, i int) bool {
	return true
}

func (s *Subfolder2) SubfolderFunctionWithReturnListParametrized() (a, b, c []byte, err error) {
	return
}`

	if err := th.CreateTestGoFile("subfolder2/subfolder2.go", subfolder2Content); err != nil {
		return err
	}

	// Create subfolder3 with interface
	if err := th.CreateTestDirectory("subfolder3"); err != nil {
		return err
	}

	subfolder3Content := `package subfolder3

// SubfolderInterface for testing purposes
type SubfolderInterface interface {
	SubfolderFunction(bool, int) bool
}`

	return th.CreateTestGoFile("subfolder3/subfolder3.go", subfolder3Content)
}

// CreateRecursiveTestStructure creates a comprehensive recursive test structure
func (th *TestHelper) CreateRecursiveTestStructure() error {
	// Create all test data structures
	if err := th.CreateTestDataStructure(); err != nil {
		return err
	}

	if err := th.CreateComplexTestScenario(); err != nil {
		return err
	}

	if err := th.CreateParenthesizedTypesTestData(); err != nil {
		return err
	}

	if err := th.CreateSubfolderTestData(); err != nil {
		return err
	}

	return nil
}

// ValidateTestStructure validates that all test files were created correctly
func (th *TestHelper) ValidateTestStructure(t *testing.T) {
	testDirs := []string{
		"connectionlabels",
		"namedimports",
		"renderingoptions",
		"parenthesizedtypes",
		"subfolder",
		"subfolder2",
		"subfolder3",
	}

	for _, dir := range testDirs {
		exists, err := afero.DirExists(th.FileSystem, dir)
		if err != nil {
			t.Errorf("Error checking directory %s: %v", dir, err)
		}
		if !exists {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}
}
