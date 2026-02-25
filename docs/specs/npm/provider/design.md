# Design: npm Provider

## Current State

- Provider インターフェース (`Scheme()`, `Fetch()`) と Registry が確立済み
- GitHub プロバイダーが唯一の実装
- `Result` 構造体は `GitHub *GitHubMetrics` フィールドのみ保持
- Markdown フォーマッターは GitHub テーブルのみをレンダリング
- Target パーサーは `npm:<package>` 形式をすでにパース可能

## Proposed Changes

### 新規ファイル

- `internal/provider/npm/npm.go` — npm プロバイダー実装
- `internal/provider/npm/npm_test.go` — ユニットテスト

### 変更ファイル

- `internal/provider/provider.go` — `NPMMetrics` 構造体追加、`Result` に `NPM` フィールド追加
- `internal/cli/cli.go` — npm プロバイダーを Registry に登録
- `internal/format/format.go` — Markdown フォーマッターに npm テーブル追加

### 変更不要

- `internal/provider/target.go` — すでに `npm:` スキームをパース可能
- `internal/provider/registry.go` — スキームベースのルックアップは汎用
- `cmd/repiq/main.go` — エントリポイントは変更不要
- `internal/auth/auth.go` — npm は認証不要

## Backend Spec

### データモデル

```
NPMMetrics {
  weekly_downloads   int
  monthly_downloads  int
  latest_version     string
  last_publish_days  int
  dependencies_count int
  license            string
}
```

`Result.NPM *NPMMetrics` として追加 (`json:"npm,omitempty"`)。

### API エンドポイント

| # | Endpoint | 取得データ | Host |
|---|----------|-----------|------|
| 1 | `GET /{package}/latest` | version, dependencies, license | registry.npmjs.org |
| 2 | `GET /{package}` (abbreviated) | modified (last_publish_days 計算用) | registry.npmjs.org |
| 3 | `GET /downloads/point/last-week/{package}` | weekly_downloads | api.npmjs.org |
| 4 | `GET /downloads/point/last-month/{package}` | monthly_downloads | api.npmjs.org |

- エンドポイント 2 は `Accept: application/vnd.npm.install-v1+json` ヘッダーで abbreviated metadata を取得
- scoped package は downloads API で `@scope%2Fname` にエンコード
- エンドポイント 3 と 4 は同一レスポンス形式 (`{downloads: int}`)、共通の `fetchDownloads(ctx, pkg, period)` で実装
- 4 リクエストを goroutine で並列実行

### 処理フロー

```
Fetch(ctx, identifier)
  ├── validate identifier (空文字チェック)
  ├── goroutine 1: GET /{pkg}/latest
  │     → latest_version, dependencies_count, license
  ├── goroutine 2: GET /{pkg} (abbreviated)
  │     → modified → last_publish_days 計算
  ├── goroutine 3: GET /downloads/point/last-week/{pkg}
  │     → weekly_downloads
  └── goroutine 4: GET /downloads/point/last-month/{pkg}
        → monthly_downloads
  ├── sync.WaitGroup.Wait()
  └── return Result{Target: "npm:{identifier}", NPM: &NPMMetrics{...}}
```

### エラーハンドリング

- 各 goroutine のエラーは個別に収集
- 全リクエスト失敗: `Result.Error` にエラーメッセージを設定、`NPM` は nil
- 部分失敗: 取得できたメトリクスは設定、エラーメッセージも error に記録
- 404: 「package not found: {identifier}」
- タイムアウト: 各 goroutine は context.Context の 30 秒タイムアウト (CLI グローバル設定) を個別に参照。タイムアウトした goroutine のメトリクスのみゼロ値とし、成功した goroutine のメトリクスは返却する

### 出力スキーマ

成功時:
```json
{
  "target": "npm:react",
  "npm": {
    "weekly_downloads": 25000000,
    "monthly_downloads": 100000000,
    "latest_version": "19.1.0",
    "last_publish_days": 15,
    "dependencies_count": 2,
    "license": "MIT"
  }
}
```

混合取得時:
```json
[
  {
    "target": "github:facebook/react",
    "github": { "stars": 215000, ... }
  },
  {
    "target": "npm:react",
    "npm": { "weekly_downloads": 25000000, ... }
  }
]
```

エラー時:
```json
{
  "target": "npm:nonexistent-pkg-xyz",
  "error": "npm registry: package not found: nonexistent-pkg-xyz"
}
```

### Markdown 出力

npm 結果を GitHub とは別テーブルにレンダリング:

```markdown
## GitHub

| target | stars | forks | ... |
|--------|-------|-------|-----|
| github:facebook/react | 215000 | 45000 | ... |

## npm

| target | weekly_downloads | monthly_downloads | latest_version | last_publish_days | dependencies_count | license |
|--------|-----------------|-------------------|----------------|-------------------|-------------------|---------|
| npm:react | 25000000 | 100000000 | 19.1.0 | 15 | 2 | MIT |
```

## Tracking

| Event Name | Properties | Trigger Condition |
|------------|-----------|-------------------|
| N/A | N/A | CLI ツールのためトラッキングなし |
