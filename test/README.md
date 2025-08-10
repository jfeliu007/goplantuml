# GoPlantUML Test Suite

このディレクトリには、GoPlantUMLプロジェクトの包括的なテストスイートが含まれています。

## ディレクトリ構造

```
test/
├── README.md                    # このファイル
├── test_helper.go               # テストユーティリティ
├── results/                     # テスト結果とカバレッジレポート
├── cmd/                         # テスト実行用メインファイル
│   └── main.go                  # カスタムテストランナー
└── pkg/                         # パッケージごとのテスト
    ├── client/
    │   ├── controller/          # コントローラーレイヤーのテスト
    │   ├── repository/          # リポジトリレイヤーのテスト
    │   └── usecase/             # ユースケースレイヤーのテスト
    └── config/                  # 設定関連のテスト

etc/test/data/                   # 外部テストデータファイル
├── yaml/                        # YAML設定ファイル
└── go/                          # Goソースファイル
```

## ✨ 美しいテスト出力

このテストランナーは、`lipgloss`と`rodaine/table`を使用して美しい出力を提供します：

### 🎨 機能
- **カラフルなテーブル**: テスト結果を見やすいテーブル形式で表示
- **リアルタイム出力**: テスト実行中のリアルタイムフィードバック
- **統計情報**: 成功率、実行時間、テスト数の詳細統計
- **カバレッジレポート**: カラー化されたカバレッジ情報
- **エラーハイライト**: 失敗したテストのハイライト表示

### 🎯 出力例
```
🧪 GoPlantUML Test Runner

📊 Test Results
 Metric        Value               
 Total Tests   5                  
 Passed        5  
 Failed        0                  
 Success Rate  100.0%             
 Duration      108.319709ms       

🎉 すべてのテストが成功しました!
```

## テストの実行

### 🎨 美しい出力（推奨）

#### 1. すべてのテストを実行
```bash
# 基本実行
make test
# または
go run test/cmd/main.go

# 詳細出力付き
make test-verbose
# または  
go run test/cmd/main.go -v

# カバレッジ付き
make test-coverage
# または
go run test/cmd/main.go -cover
```

#### 2. パッケージ別テスト実行
```bash
# リポジトリレイヤーのテスト
make test-repository
# または
go run test/cmd/main.go -pkg repository -v

# コントローラレイヤーのテスト
make test-controller
# または
go run test/cmd/main.go -pkg controller -v

# ユースケースレイヤーのテスト
make test-usecase
# または
go run test/cmd/main.go -pkg usecase -v

# 設定のテスト
make test-config
# または
go run test/cmd/main.go -pkg config -v
```

#### 3. 特定のテスト関数を実行
```bash
# 特定のテスト関数を実行
make test-func FUNC=TestPlantUMLRepository_GenerateDiagram_BasicStruct
# または
go run test/cmd/main.go -test TestPlantUMLRepository_GenerateDiagram_BasicStruct -v

# PlantUMLリポジトリのテストのみ
make test-plantuml
# または
go run test/cmd/main.go -file plantuml_test.go -pkg repository -v
```

#### 4. 特定のテストファイルを実行
```bash
# 特定のテストファイルを実行
make test-file FILE=plantuml_test.go PKG=repository
# または
go run test/cmd/main.go -file plantuml_test.go -pkg repository -v
```

### 📋 従来のシンプル出力

シンプルな出力が必要な場合は、`-pretty=false`オプションを使用します：

```bash
# シンプルな出力でテスト実行
go run test/cmd/main.go -pretty=false
```

### 🛠️ 従来のgo testコマンド
```bash
# リポジトリレイヤーのテスト
go test ./test/pkg/client/repository/

# コントローラレイヤーのテスト
go test ./test/pkg/client/controller/

# 設定のテスト
go test ./test/pkg/config/

# 全テストの実行
go test ./test/...
```

## テストの種類

### 1. ユニットテスト
- **リポジトリレイヤー**: PlantUML図生成の核となる機能をテスト
- **コントローラレイヤー**: リクエスト処理とレスポンス生成をテスト
- **設定**: YAML設定ファイルの読み込みと検証をテスト

### 2. 統合テスト
- エンドツーエンドのワークフローをテスト
- ファイルシステムとの実際の相互作用をテスト
- 複数レイヤー間の連携をテスト

## テストヘルパー

`test_helper.go`には以下のユーティリティが含まれています：

### TestHelper構造体
- メモリファイルシステムまたは実ファイルシステムでのテスト
- テスト用Goファイルの生成
- YAML設定ファイルの生成
- テストデータの管理

### 主要メソッド
- `NewTestHelper()`: メモリファイルシステムでのテストヘルパー作成
- `NewTestHelperWithRealFS()`: 実ファイルシステムでのテストヘルパー作成
- `CreateTestGoFile()`: Goソースファイルの作成
- `CreateTestStructFile()`: 構造体ファイルの作成
- `CreateTestInterfaceFile()`: インターフェースファイルの作成
- `CreateTestYamlConfig()`: YAML設定ファイルの作成

## テスト戦略

### Clean Architecture準拠
テストは以下のレイヤー構造に従って組織化されています：

1. **Controller Layer**: HTTPリクエスト/レスポンス処理
2. **Usecase Layer**: ビジネスロジックの実行
3. **Repository Layer**: データアクセスとPlantUML生成

### モック戦略
- ファイルシステム操作には`afero`のメモリファイルシステムを使用
- 依存関係の注入によりテスト可能性を向上
- 外部依存関係の分離

### テストデータ
- `etc/test/data/`ディレクトリにサンプルGoファイルを配置
- テストケースごとに動的にファイルを生成
- 実際のGoコード構造を模倣

## カバレッジレポート

テスト実行後、以下のレポートが生成されます：

- `etc/test/results/coverage.out`: カバレッジデータ
- `etc/test/results/coverage.html`: HTMLカバレッジレポート
- `etc/test/results/*_coverage.out`: パッケージ別カバレッジ

## 継続的インテグレーション

このテストスイートは以下の環境で実行されることを想定しています：

- Go 1.24+
- Linux/macOS/Windows
- GitHub Actions (将来実装予定)

## テストの拡張

新しいテストを追加する際は以下のガイドラインに従ってください：

1. **命名規則**: `Test{ComponentName}_{Method}_{Scenario}`
2. **構造**: Setup → Execute → Assert
3. **クリーンアップ**: `defer testHelper.Cleanup()`を使用
4. **エラーハンドリング**: 適切なエラーメッセージとコンテキストを提供

## 既知の制限事項

- YAML設定ファイルのテストは実ファイルシステムが必要
- 一部の統合テストには外部依存関係が必要
- パフォーマンステストは現在未実装

## トラブルシューティング

### テスト失敗時の対処

1. **ビルドエラー**: `go mod tidy`を実行
2. **パーミッションエラー**: ファイルの権限を確認
3. **カバレッジエラー**: `etc/test/results/`ディレクトリの権限を確認

### デバッグ

```bash
# 詳細なテスト出力
go test -v ./test/...

# 特定のテストのみ実行
go test -v -run TestSpecificFunction ./test/pkg/client/repository/
```
