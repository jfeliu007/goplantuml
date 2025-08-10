package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var (
		testName = flag.String("test", "", "特定のテスト関数を実行 (例: TestPlantUMLRepository_GenerateDiagram_BasicStruct)")
		testFile = flag.String("file", "", "特定のテストファイルを実行 (例: plantuml_test.go)")
		pkg      = flag.String("pkg", "", "特定のパッケージのテストを実行 (例: repository, controller, usecase)")
		verbose  = flag.Bool("v", false, "詳細な出力を表示")
		coverage = flag.Bool("cover", false, "カバレッジレポートを生成")
		help     = flag.Bool("help", false, "ヘルプを表示")
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

	// テストを実行
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

func printHelp() {
	fmt.Println("GoPlantUML テスト実行ツール")
	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Println("  go run cmd/test/main.go [オプション]")
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
	fmt.Println("  -help           このヘルプを表示")
	fmt.Println()
	fmt.Println("例:")
	fmt.Println("  # すべてのテストを実行")
	fmt.Println("  go run cmd/test/main.go")
	fmt.Println()
	fmt.Println("  # repositoryパッケージのテストを実行")
	fmt.Println("  go run cmd/test/main.go -pkg repository")
	fmt.Println()
	fmt.Println("  # 特定のテスト関数を実行")
	fmt.Println("  go run cmd/test/main.go -test TestPlantUMLRepository_GenerateDiagram_BasicStruct")
	fmt.Println()
	fmt.Println("  # 特定のテストファイルを実行")
	fmt.Println("  go run cmd/test/main.go -file plantuml_test.go -pkg repository")
	fmt.Println()
	fmt.Println("  # カバレッジ付きで詳細出力")
	fmt.Println("  go run cmd/test/main.go -v -cover")
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
