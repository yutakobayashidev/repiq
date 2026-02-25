# Design: CLI + GitHub Provider

## Current State

コードが存在しない。Go 1.24 + Nix flake の開発環境のみ。

## Proposed Changes

repiq の全コードを新規作成する。以下の 4 レイヤーで構成:

### 1. CLI レイヤー

- エントリポイント: `main.go`
- 引数パース: `<scheme>:<identifier>` を解析し、scheme と identifier に分離
- フラグ: `--json` / `--ndjson` / `--markdown` (default) / `--help` / `--version`
- 複数ターゲットを受け取り、Provider に渡して並列実行
- 結果を指定フォーマットで stdout に出力。エラーは stderr

### 2. Provider インターフェース

- `Provider` インターフェースを定義
- scheme 名でプロバイダーを解決するレジストリ (map)
- 各プロバイダーは `Fetch(ctx, identifier) -> Result` を実装
- Result は scheme 固有のフィールドを持つ構造体

### 3. GitHub プロバイダー

- `github` scheme を処理
- google/go-github クライアントを使用
- 認証トークン解決: `gh auth token` (exec) → `GITHUB_TOKEN` (env) → 未認証
- 取得データ:
  - `GET /repos/{owner}/{repo}` → stars, forks, open_issues, license
  - `GET /repos/{owner}/{repo}/contributors?per_page=1` → Link ヘッダーから contributors 推定
  - `GET /repos/{owner}/{repo}/releases?per_page=1` → Link ヘッダーから release_count 推定
  - `GET /repos/{owner}/{repo}/commits?per_page=1` → 最新コミット日時から last_commit_days 算出
  - Search API → commits_30d, issues_closed_30d
- 各 API コールは goroutine で並列実行

### 4. 出力フォーマッター

- JSON: `encoding/json` で配列出力
- NDJSON: 1レコードずつ JSON エンコードして改行
- Markdown: scheme ごとにグループ化してテーブル出力

## 出力スキーマ

### 成功時

```json
{
  "target": "github:facebook/react",
  "github": {
    "stars": 215000,
    "forks": 45000,
    "open_issues": 980,
    "contributors": 1623,
    "release_count": 210,
    "last_commit_days": 2,
    "commits_30d": 120,
    "issues_closed_30d": 340,
    "license": "MIT"
  }
}
```

### エラー時

```json
{
  "target": "github:nonexistent/repo",
  "error": "GitHub API: 404 Not Found"
}
```

### Markdown 出力

```
| target | stars | forks | open_issues | contributors | release_count | last_commit_days | commits_30d | issues_closed_30d | license |
|--------|-------|-------|-------------|--------------|---------------|------------------|-------------|---------------------|---------|
| github:facebook/react | 215000 | 45000 | 980 | 1623 | 210 | 2 | 120 | 340 | MIT |
```

## パッケージ構成

```
cmd/repiq/main.go          # エントリポイント
internal/cli/cli.go        # 引数パース、実行制御
internal/provider/provider.go  # Provider インターフェース、レジストリ
internal/provider/github/github.go  # GitHub プロバイダー
internal/auth/auth.go      # トークン解決 (gh auth token / GITHUB_TOKEN)
internal/format/format.go  # 出力フォーマッター (JSON/NDJSON/Markdown)
```

## Tracking

N/A (CLI ツール、トラッキングイベントなし)
