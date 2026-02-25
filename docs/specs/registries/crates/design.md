# Design: crates.io プロバイダー

## Current State

repiq は GitHub と npm の 2 プロバイダーに対応。crates.io は単一 API ベース URL で複数エンドポイントを提供しており、npm プロバイダーの 2 API 並列パターンを拡張する形で実装できる。

## Proposed Changes

### 1. Result 型の拡張

`provider.go` の `Result` 構造体に `Crates *CratesMetrics` フィールドを追加する。

```
CratesMetrics:
  downloads             int
  recent_downloads      int
  latest_version        string
  last_publish_days     int
  dependencies_count    int
  license               string
  reverse_dependencies  int
```

### 2. crates.io プロバイダーの新規作成

`internal/provider/crates/` パッケージを作成する。

**コンストラクタ**: crates.io API ベース URL を引数に取る。デフォルトは `https://crates.io`。HTTP クライアントに `User-Agent` ヘッダーを自動付与する RoundTripper を設定する。

**identifier バリデーション**: crates.io のクレート名規則 (英数字、ハイフン、アンダースコア) に基づく正規表現でバリデーション。

**並列フェッチ**: 3 つの API コールを並列実行する。

| Job | API エンドポイント | 取得メトリクス |
|-----|-------------------|-------------|
| metadata | `GET /api/v1/crates/{crate}` | downloads, recent_downloads, latest_version, last_publish_days, license |
| dependencies | `GET /api/v1/crates/{crate}/{version}/dependencies` | dependencies_count |
| reverse_deps | `GET /api/v1/crates/{crate}/reverse_dependencies?per_page=1` | reverse_dependencies (`meta.total`) |

**依存関係の順序**: metadata job の結果から `latest_version` を取得した後に dependencies job を実行する必要がある。そのため、metadata → (dependencies + reverse_deps の並列) という 2 段階の実行フローとなる。

**エラーハンドリング**: metadata が失敗したら全メトリクスが nil。dependencies/reverse_deps のみ失敗なら partial result。

### 3. User-Agent ヘッダー

GitHub プロバイダーの `tokenTransport` パターンを参考に、`User-Agent` を自動付与するカスタム `RoundTripper` を実装する。

```
User-Agent: repiq/0.0.0 (https://github.com/yutakobayashidev/repiq)
```

### 4. CLI 登録

`internal/cli/cli.go` で crates.io プロバイダーを生成し、cache decorator でラップしてレジストリに登録する。

### 5. Markdown フォーマッター拡張

`internal/format/format.go` に crates.io 用の switch case と Markdown テーブルレンダリングを追加する。

## Backend Spec

### API エンドポイント

**Crate メタデータ**

```
GET https://crates.io/api/v1/crates/{crate}

Response (抜粋):
{
  "crate": {
    "downloads": 835433978,
    "recent_downloads": 116815564,
    "max_stable_version": "1.0.228",
    "newest_version": "1.0.228",
    "updated_at": "2025-09-27T16:51:35Z"
  },
  "versions": [
    {
      "num": "1.0.228",
      "license": "MIT OR Apache-2.0",
      "created_at": "2025-09-27T..."
    }
  ]
}
```

- `downloads`: `crate.downloads`
- `recent_downloads`: `crate.recent_downloads`
- `latest_version`: `crate.max_stable_version` (null なら `crate.newest_version`)
- `last_publish_days`: `versions[0].created_at` (最新バージョン) から算出
- `license`: `versions[0].license`

**依存関係**

```
GET https://crates.io/api/v1/crates/{crate}/{version}/dependencies

Response (抜粋):
{
  "dependencies": [
    {"crate_id": "serde_derive", "kind": "normal", "optional": false},
    {"crate_id": "serde_test", "kind": "dev", "optional": false}
  ]
}
```

- `dependencies_count`: `kind == "normal"` の依存のみカウント (`dev` と `build` は除外)

**被依存クレート数**

```
GET https://crates.io/api/v1/crates/{crate}/reverse_dependencies?per_page=1

Response (抜粋):
{
  "meta": {"total": 72719}
}
```

- `reverse_dependencies`: `meta.total`
- `per_page=1` で最小限のデータ転送 (total 数だけが必要)

## Tracking

| Event Name | Properties | Trigger Condition |
|------------|------------|-------------------|
| N/A | CLI ツールのためトラッキングなし | - |
