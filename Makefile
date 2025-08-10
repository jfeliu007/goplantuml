# GoPlantUML Makefile

.PHONY: help build test test-verbose test-coverage test-pkg test-func clean install dev

# デフォルトターゲット
help:
	@echo "利用可能なコマンド:"
	@echo "  make build          - アプリケーションをビルド"
	@echo "  make test           - すべてのテストを実行"
	@echo "  make test-verbose   - 詳細出力でテストを実行"
	@echo "  make test-coverage  - カバレッジ付きでテストを実行"
	@echo "  make test-pkg PKG=repository - 特定パッケージのテストを実行"
	@echo "  make test-func FUNC=TestPlantUMLRepository_GenerateDiagram_BasicStruct - 特定関数のテストを実行"
	@echo "  make clean          - ビルド成果物を削除"
	@echo "  make install        - 依存関係をインストール"
	@echo "  make dev            - 開発環境セットアップ"

# アプリケーションをビルド
build:
	go build -o bin/goplantuml cmd/goplantuml/main.go
	go build -o bin/client cmd/client/main.go

# 依存関係をインストール
install:
	go mod download
	go mod tidy

# 開発環境セットアップ
dev: install
	@echo "開発環境のセットアップが完了しました"

# すべてのテストを実行
test:
	go run test/cmd/main.go

# 詳細出力でテストを実行
test-verbose:
	go run test/cmd/main.go -v

# カバレッジ付きでテストを実行
test-coverage:
	go run test/cmd/main.go -cover
	@echo "カバレッジレポートは test/results/coverage.out に保存されました"

# 特定パッケージのテストを実行
# 使用例: make test-pkg PKG=repository
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "エラー: PKGを指定してください (例: make test-pkg PKG=repository)"; \
		echo "利用可能なパッケージ: repository, controller, usecase, config"; \
		exit 1; \
	fi
	go run test/cmd/main.go -pkg $(PKG) -v

# 特定のテスト関数を実行
# 使用例: make test-func FUNC=TestPlantUMLRepository_GenerateDiagram_BasicStruct
test-func:
	@if [ -z "$(FUNC)" ]; then \
		echo "エラー: FUNCを指定してください"; \
		echo "例: make test-func FUNC=TestPlantUMLRepository_GenerateDiagram_BasicStruct"; \
		exit 1; \
	fi
	go run test/cmd/main.go -test $(FUNC) -v

# 特定のテストファイルを実行
# 使用例: make test-file FILE=plantuml_test.go PKG=repository
test-file:
	@if [ -z "$(FILE)" ]; then \
		echo "エラー: FILEを指定してください"; \
		echo "例: make test-file FILE=plantuml_test.go PKG=repository"; \
		exit 1; \
	fi
	@if [ -z "$(PKG)" ]; then \
		go run test/cmd/main.go -file $(FILE) -v; \
	else \
		go run test/cmd/main.go -file $(FILE) -pkg $(PKG) -v; \
	fi

# repositoryパッケージのテスト
test-repository:
	go run test/cmd/main.go -pkg repository -v

# controllerパッケージのテスト
test-controller:
	go run test/cmd/main.go -pkg controller -v

# usecaseパッケージのテスト
test-usecase:
	go run test/cmd/main.go -pkg usecase -v

# configパッケージのテスト
test-config:
	go run test/cmd/main.go -pkg config -v

# 特定のPlantUMLテストを実行
test-plantuml:
	go run test/cmd/main.go -file plantuml_test.go -pkg repository -v

# ビルド成果物を削除
clean:
	rm -rf bin/
	rm -f test/results/coverage.out
	go clean

# 依存関係を更新
update:
	go get -u ./...
	go mod tidy

# linting
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

# フォーマット
format:
	go fmt ./...
	goimports -w .

# 全体チェック（テスト、lint、フォーマット）
check: format lint test

# リリースビルド
release:
	@echo "リリースビルドを実行中..."
	GOOS=linux GOARCH=amd64 go build -o bin/goplantuml-linux-amd64 cmd/goplantuml/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/goplantuml-darwin-amd64 cmd/goplantuml/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/goplantuml-windows-amd64.exe cmd/goplantuml/main.go
	@echo "リリースビルド完了"

# テスト結果ディレクトリを作成
test-results-dir:
	mkdir -p test/results
