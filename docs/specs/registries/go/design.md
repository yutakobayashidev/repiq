# Design: Go Modules プロバイダー

## Current State

repiq は GitHub と npm の 2 プロバイダーに対応。Go Modules は他のプロバイダーと異なり、2 つの独立したサービス (proxy.golang.org + deps.dev) からデータを取得する。npm プロバイダーの 2 API 並列パターンと同じ構造で実装できる。

## Proposed Changes

### 1. Result 型の拡張

`provider.go` の `Result` 構造体に `Go *GoMetrics` フィールドを追加する。

```
GoMetrics:
  latest_version      string
  last_publish_days   int
  dependencies_count  int
  license             string
```

### 2. Go Modules プロバイダーの新規作成

`internal/provider/golang/` パッケージを作成する (パッケージ名は `go` が予約語のため `golang` を使用)。

**コンストラクタ**: proxy.golang.org URL と deps.dev API URL を引数に取る。デフォルトはそれぞれ `https://proxy.golang.org` と `https://api.deps.dev`。

**identifier バリデーション**: Go モジュールパスの基本的な形式チェック。ドメインを含むパス形式 (`domain.com/path` 形式で少なくとも 1 つの `/` を含む) を検証する。

**並列フェッチ**: 2 段階の実行フローで 2 つの API を利用する。

| Job | API エンドポイント | 取得メトリクス |
|-----|-------------------|-------------|
| proxy | `GET /{module}/@latest` | latest_version, last_publish_days |
| deps_info | `GET /v3alpha/systems/go/packages/{module}/versions/{version}` | license |
| deps_graph | `GET /v3alpha/systems/go/packages/{module}/versions/{version}:dependencies` | dependencies_count |

**実行順序**: proxy job で `latest_version` を取得 → deps_info と deps_graph を並列実行 (バージョン番号が必要なため)。

**deps.dev フォールバック**: deps.dev API が失敗した場合、`dependencies_count: 0`, `license: ""` を返し、エラーメッセージに deps.dev の失敗を含める。proxy の結果は正常に返す。

### 3. モジュールパスのエスケープ

Go module proxy はモジュールパスに大文字が含まれる場合、大文字を `!` + 小文字にエスケープする規則がある (例: `GitHub.com` → `!github.com`)。このエスケープ処理を実装する。

### 4. CLI 登録

`internal/cli/cli.go` で Go Modules プロバイダーを生成し、cache decorator でラップしてレジストリに登録する。scheme は `"go"` を使用する。

### 5. Markdown フォーマッター拡張

`internal/format/format.go` に Go Modules 用の switch case と Markdown テーブルレンダリングを追加する。

## Backend Spec

### API エンドポイント

**Go Module Proxy (バージョン情報)**

```
GET https://proxy.golang.org/{escaped_module}/@latest

Response:
{
  "Version": "v0.34.0",
  "Time": "2026-02-09T16:14:29Z",
  "Origin": {
    "VCS": "git",
    "URL": "https://go.googlesource.com/text",
    "Hash": "817fba9abd337b4d9097b10c61a540c74feaaeff",
    "Ref": "refs/tags/v0.34.0"
  }
}
```

- `latest_version`: `Version` フィールド
- `last_publish_days`: `Time` フィールドから算出

**deps.dev バージョン情報 (ライセンス)**

```
GET https://api.deps.dev/v3alpha/systems/go/packages/{url_encoded_module}/versions/{version}

Response (抜粋):
{
  "licenses": ["BSD-3-Clause"],
  "links": [{"label": "SOURCE_REPO", "url": "https://..."}]
}
```

- `license`: `licenses[0]` (複数ある場合は ` OR ` で結合)

**deps.dev 依存関係グラフ**

```
GET https://api.deps.dev/v3alpha/systems/go/packages/{url_encoded_module}/versions/{version}:dependencies

Response (抜粋):
{
  "nodes": [
    {"versionKey": {"system": "GO", "name": "...", "version": "..."}, "relation": "SELF"},
    {"versionKey": {"system": "GO", "name": "...", "version": "..."}, "relation": "DIRECT"}
  ]
}
```

- `dependencies_count`: `relation == "DIRECT"` のノード数

## Tracking

| Event Name | Properties | Trigger Condition |
|------------|------------|-------------------|
| N/A | CLI ツールのためトラッキングなし | - |
