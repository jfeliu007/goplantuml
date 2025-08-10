package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rodaine/table"
)

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	failStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("51")).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)
)

type TestResult struct {
	Type     string
	Message  string
	Package  string
	Test     string
	Duration string
}

type TestStats struct {
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
}

func createBanner() string {
	banner := `
╔═══════════════════════════════════════════════╗
║            GoPlantUML Test Runner             ║
╚═══════════════════════════════════════════════╝`
	return bannerStyle.Render(banner)
}

func createResultsBanner() string {
	banner := `
╔═══════════════════════════════════════════════╗
║                 Test Results                  ║
╚═══════════════════════════════════════════════╝`
	return headerStyle.Render(banner)
}

func createDetailsBanner() string {
	banner := `
╔═══════════════════════════════════════════════╗
║                Test Details                   ║
╚═══════════════════════════════════════════════╝`
	return headerStyle.Render(banner)
}

func createCoverageBanner() string {
	banner := `
╔═══════════════════════════════════════════════╗
║               Coverage Report                 ║
╚═══════════════════════════════════════════════╝`
	return headerStyle.Render(banner)
}

func main() {
	var (
		testName = flag.String("test", "", "特定のテスト関数を実行 (例: TestPlantUMLRepository_GenerateDiagram_BasicStruct)")
		testFile = flag.String("file", "", "特定のテストファイルを実行 (例: plantuml_test.go)")
		pkg      = flag.String("pkg", "", "特定のパッケージのテストを実行 (例: repository, controller, usecase)")
		verbose  = flag.Bool("v", false, "詳細な出力を表示")
		coverage = flag.Bool("cover", false, "カバレッジレポートを生成")
		help     = flag.Bool("help", false, "ヘルプを表示")
		pretty   = flag.Bool("pretty", true, "美しい出力を使用")
	)
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// プロジェクトルートディレクトリを取得
	projectRoot, err := getProjectRoot()
	if err != nil {
		fmt.Printf("エラー: プロジェクトルートが見つかりません: %v\n", err)
		os.Exit(1)
	}

	// テストディレクトリに移動
	testDir := filepath.Join(projectRoot, "test")
	if err := os.Chdir(testDir); err != nil {
		fmt.Printf("エラー: テストディレクトリに移動できません: %v\n", err)
		os.Exit(1)
	}

	// テストコマンドを構築
	args := []string{"test", "-mod=mod"}

	// 詳細出力
	if *verbose {
		args = append(args, "-v")
	}

	// カバレッジ
	if *coverage {
		args = append(args, "-cover")
		args = append(args, "-coverprofile=results/coverage.out")
	}

	// 特定のテスト関数を実行
	if *testName != "" {
		args = append(args, "-run", *testName)
		if *pkg != "" {
			args = append(args, fmt.Sprintf("./pkg/client/%s", *pkg))
		} else {
			args = append(args, "./...")
		}
	} else if *testFile != "" {
		// 特定のテストファイルを実行
		if *pkg != "" {
			testPath := fmt.Sprintf("./pkg/client/%s", *pkg)
			args = append(args, testPath)
		} else {
			// ファイル名からパッケージを推測
			pkgName := inferPackageFromFile(*testFile)
			if pkgName != "" {
				testPath := fmt.Sprintf("./pkg/client/%s", pkgName)
				args = append(args, testPath)
			} else {
				args = append(args, "./...")
			}
		}
		if *testName == "" && strings.HasSuffix(*testFile, "_test.go") {
			// ファイル名に基づいてテスト関数を制限
			prefix := strings.TrimSuffix(filepath.Base(*testFile), "_test.go")
			args = append(args, "-run", fmt.Sprintf(".*%s.*", strings.Title(prefix)))
		}
	} else if *pkg != "" {
		// 特定のパッケージのテストを実行
		args = append(args, fmt.Sprintf("./pkg/client/%s", *pkg))
	} else {
		// すべてのテストを実行
		args = append(args, "./...")
	}

	if *pretty {
		runTestsWithPrettyOutput(args, *coverage)
	} else {
		runTestsPlain(args)
	}
}

func runTestsPlain(args []string) {
	fmt.Printf("実行中: go %s\n", strings.Join(args, " "))
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("テスト実行エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("テスト実行完了!")
}

func runTestsWithPrettyOutput(args []string, withCoverage bool) {
	// ヘッダーを表示
	fmt.Println(createBanner())
	fmt.Println()

	// 実行コマンドを表示
	commandInfo := boxStyle.Render(fmt.Sprintf("Running: go %s", strings.Join(args, " ")))
	fmt.Println(infoStyle.Render(commandInfo))
	fmt.Println()

	startTime := time.Now()

	// テストを実行
	cmd := exec.Command("go", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("エラー: パイプの作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("エラー: パイプの作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("エラー: コマンドの開始に失敗しました: %v\n", err)
		os.Exit(1)
	}

	var results []TestResult
	var stats TestStats

	// 標準出力を読み取り
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		result := parseTestLine(line)
		if result.Type != "" {
			results = append(results, result)
			updateStats(&stats, result)
		}
	}

	// 標準エラー出力を読み取り
	errScanner := bufio.NewScanner(stderr)
	for errScanner.Scan() {
		line := errScanner.Text()
		fmt.Println(failStyle.Render("ERROR: ") + line)
	}

	if err := cmd.Wait(); err != nil {
		stats.Duration = time.Since(startTime)
		displayResults(results, stats, false)
		fmt.Printf(failStyle.Render("テスト実行エラー: %v\n"), err)
		os.Exit(1)
	}

	stats.Duration = time.Since(startTime)
	displayResults(results, stats, true)

	// カバレッジレポートを表示
	if withCoverage {
		fmt.Println()
		displayCoverageReport()
	}
}

func parseTestLine(line string) TestResult {
	line = strings.TrimSpace(line)

	// Regular expressions for different test output patterns
	runPattern := regexp.MustCompile(`^=== RUN\s+(.+)`)
	passPattern := regexp.MustCompile(`^--- PASS:\s+(.+)\s+\((.+)\)`)
	failPattern := regexp.MustCompile(`^--- FAIL:\s+(.+)\s+\((.+)\)`)
	skipPattern := regexp.MustCompile(`^--- SKIP:\s+(.+)\s+\((.+)\)`)
	pkgPassPattern := regexp.MustCompile(`^ok\s+(.+)\s+(.+)`)
	pkgFailPattern := regexp.MustCompile(`^FAIL\s+(.+)\s+(.+)`)
	coveragePattern := regexp.MustCompile(`coverage:\s+(.+)`)

	switch {
	case runPattern.MatchString(line):
		matches := runPattern.FindStringSubmatch(line)
		return TestResult{Type: "RUN", Test: matches[1]}
	case passPattern.MatchString(line):
		matches := passPattern.FindStringSubmatch(line)
		return TestResult{Type: "PASS", Test: matches[1], Duration: matches[2]}
	case failPattern.MatchString(line):
		matches := failPattern.FindStringSubmatch(line)
		return TestResult{Type: "FAIL", Test: matches[1], Duration: matches[2]}
	case skipPattern.MatchString(line):
		matches := skipPattern.FindStringSubmatch(line)
		return TestResult{Type: "SKIP", Test: matches[1], Duration: matches[2]}
	case pkgPassPattern.MatchString(line):
		matches := pkgPassPattern.FindStringSubmatch(line)
		return TestResult{Type: "PKG_PASS", Package: matches[1], Duration: matches[2]}
	case pkgFailPattern.MatchString(line):
		matches := pkgFailPattern.FindStringSubmatch(line)
		return TestResult{Type: "PKG_FAIL", Package: matches[1], Duration: matches[2]}
	case coveragePattern.MatchString(line):
		matches := coveragePattern.FindStringSubmatch(line)
		return TestResult{Type: "COVERAGE", Message: matches[1]}
	default:
		if strings.Contains(line, "panic:") {
			return TestResult{Type: "ERROR", Message: line}
		}
		if line != "" && !strings.HasPrefix(line, "?") {
			return TestResult{Type: "INFO", Message: line}
		}
	}

	return TestResult{}
}

func updateStats(stats *TestStats, result TestResult) {
	switch result.Type {
	case "PASS":
		stats.Total++
		stats.Passed++
	case "FAIL":
		stats.Total++
		stats.Failed++
	case "SKIP":
		stats.Total++
		stats.Skipped++
	}
}

func displayResults(results []TestResult, stats TestStats, success bool) {
	fmt.Println()
	fmt.Println(createResultsBanner())

	// 統計情報をテーブルで表示
	tbl := table.New("Metric", "Value")
	tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
		return headerStyle.Render(fmt.Sprintf(format, vals...))
	})

	// 成功率を計算
	successRate := float64(0)
	if stats.Total > 0 {
		successRate = float64(stats.Passed) / float64(stats.Total) * 100
	}

	tbl.AddRow("Total Tests", fmt.Sprintf("%d", stats.Total))
	tbl.AddRow("Passed", successStyle.Render(fmt.Sprintf("%d", stats.Passed)))
	if stats.Failed > 0 {
		tbl.AddRow("Failed", failStyle.Render(fmt.Sprintf("%d", stats.Failed)))
	} else {
		tbl.AddRow("Failed", fmt.Sprintf("%d", stats.Failed))
	}
	if stats.Skipped > 0 {
		tbl.AddRow("Skipped", warnStyle.Render(fmt.Sprintf("%d", stats.Skipped)))
	} else {
		tbl.AddRow("Skipped", fmt.Sprintf("%d", stats.Skipped))
	}
	tbl.AddRow("Success Rate", fmt.Sprintf("%.1f%%", successRate))
	tbl.AddRow("Duration", stats.Duration.String())

	tbl.Print()

	// 詳細なテスト結果
	if len(results) > 0 {
		fmt.Println()
		fmt.Println(createDetailsBanner())

		detailTable := table.New("Test", "Status", "Duration")
		detailTable.WithHeaderFormatter(func(format string, vals ...interface{}) string {
			return headerStyle.Render(fmt.Sprintf(format, vals...))
		})

		for _, result := range results {
			if result.Type == "PASS" || result.Type == "FAIL" || result.Type == "SKIP" {
				status := ""
				switch result.Type {
				case "PASS":
					status = successStyle.Render("[PASS]")
				case "FAIL":
					status = failStyle.Render("[FAIL]")
				case "SKIP":
					status = warnStyle.Render("[SKIP]")
				}
				detailTable.AddRow(result.Test, status, result.Duration)
			}
		}

		detailTable.Print()
	}

	// 最終結果
	fmt.Println()
	if success && stats.Failed == 0 {
		summary := boxStyle.Render(successStyle.Render("[ SUCCESS ] All tests passed successfully!"))
		fmt.Println(summary)
	} else {
		summary := boxStyle.Render(failStyle.Render("[ FAILED ] Some tests failed"))
		fmt.Println(summary)
	}
}

func displayCoverageReport() {
	fmt.Println(createCoverageBanner())

	// カバレッジファイルが存在するかチェック
	if _, err := os.Stat("results/coverage.out"); os.IsNotExist(err) {
		fmt.Println(warnStyle.Render("カバレッジファイルが見つかりません"))
		return
	}

	// go tool coverを使用してカバレッジ情報を取得
	cmd := exec.Command("go", "tool", "cover", "-func=results/coverage.out")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(failStyle.Render(fmt.Sprintf("カバレッジレポートの生成に失敗しました: %v", err)))
		return
	}

	lines := strings.Split(string(output), "\n")

	coverageTable := table.New("Function", "Coverage")
	coverageTable.WithHeaderFormatter(func(format string, vals ...interface{}) string {
		return headerStyle.Render(fmt.Sprintf(format, vals...))
	})

	var totalCoverage string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "total:") {
			totalCoverage = strings.TrimPrefix(line, "total:")
			totalCoverage = strings.TrimSpace(totalCoverage)
			continue
		}

		// パッケージ内の関数のカバレッジ情報を解析
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			funcName := parts[0]
			coverage := parts[len(parts)-1]

			// カバレッジに応じて色分け
			if strings.Contains(coverage, "100.0%") {
				coverage = successStyle.Render(coverage)
			} else if strings.Contains(coverage, "0.0%") {
				coverage = failStyle.Render(coverage)
			} else {
				coverage = warnStyle.Render(coverage)
			}

			coverageTable.AddRow(funcName, coverage)
		}
	}

	coverageTable.Print()

	if totalCoverage != "" {
		fmt.Println()
		fmt.Printf("Total Coverage: %s\n", successStyle.Render(totalCoverage))
	}
}

func printHelp() {
	fmt.Println(createBanner())
	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Println("  go run test/cmd/main.go [オプション]")
	fmt.Println()
	fmt.Println("オプション:")
	fmt.Println("  -test string    特定のテスト関数を実行")
	fmt.Println("                  例: -test TestPlantUMLRepository_GenerateDiagram_BasicStruct")
	fmt.Println("  -file string    特定のテストファイルを実行")
	fmt.Println("                  例: -file plantuml_test.go")
	fmt.Println("  -pkg string     特定のパッケージのテストを実行")
	fmt.Println("                  例: -pkg repository")
	fmt.Println("                  選択肢: repository, controller, usecase, config")
	fmt.Println("  -v              詳細な出力を表示")
	fmt.Println("  -cover          カバレッジレポートを生成")
	fmt.Println("  -pretty         美しい出力を使用 (デフォルト: true)")
	fmt.Println("  -help           このヘルプを表示")
	fmt.Println()
	fmt.Println("例:")
	fmt.Println("  # すべてのテストを実行")
	fmt.Println("  go run test/cmd/main.go")
	fmt.Println()
	fmt.Println("  # repositoryパッケージのテストを実行")
	fmt.Println("  go run test/cmd/main.go -pkg repository")
	fmt.Println()
	fmt.Println("  # 特定のテスト関数を実行")
	fmt.Println("  go run test/cmd/main.go -test TestPlantUMLRepository_GenerateDiagram_BasicStruct")
	fmt.Println()
	fmt.Println("  # 特定のテストファイルを実行")
	fmt.Println("  go run test/cmd/main.go -file plantuml_test.go -pkg repository")
	fmt.Println()
	fmt.Println("  # カバレッジ付きで詳細出力")
	fmt.Println("  go run test/cmd/main.go -v -cover")
	fmt.Println()
	fmt.Println("  # シンプルな出力")
	fmt.Println("  go run test/cmd/main.go -pretty=false")
}

func getProjectRoot() (string, error) {
	// 現在のディレクトリから上位にgo.modファイルを探す
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// ルートディレクトリに到達
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.modファイルが見つかりません")
}

func inferPackageFromFile(filename string) string {
	// ファイル名からパッケージを推測
	if strings.Contains(filename, "controller") {
		return "controller"
	}
	if strings.Contains(filename, "repository") {
		return "repository"
	}
	if strings.Contains(filename, "usecase") {
		return "usecase"
	}
	if strings.Contains(filename, "config") {
		return "config"
	}
	return ""
}
