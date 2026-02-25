# Design: PyPI プロバイダー

## Current State

repiq は GitHub と npm の 2 プロバイダーに対応。Provider インターフェース (`Scheme() + Fetch()`) と Result 型 (`*GitHubMetrics`, `*NPMMetrics`) が確立されている。npm プロバイダーが 2 つの API エンドポイント (registry + downloads) に並列リクエストする先例がある。

## Proposed Changes

### 1. Result 型の拡張

`provider.go` の `Result` 構造体に `PyPI *PyPIMetrics` フィールドを追加する。

```
PyPIMetrics:
  weekly_downloads    int
  monthly_downloads   int
  latest_version      string
  last_publish_days   int
  dependencies_count  int
  license             string
  requires_python     string
```

### 2. PyPI プロバイダーの新規作成

`internal/provider/pypi/` パッケージを作成する。

**コンストラクタ**: PyPI JSON API URL と pypistats.org API URL を引数に取る (テスト時にモック URL を渡せるようにする)。デフォルト値はそれぞれ `https://pypi.org` と `https://pypistats.org`。

**identifier バリデーション**: PyPI パッケージ名の正規表現でバリデーション。PEP 508 に基づき、英数字・ハイフン・アンダースコア・ドットを許容する。

**並列フェッチ**: npm プロバイダーと同じ WaitGroup + Mutex パターンで 2 つの API を並列実行する。

| Job | API エンドポイント | 取得メトリクス |
|-----|-------------------|-------------|
| metadata | `GET /pypi/{package}/json` | latest_version, last_publish_days, dependencies_count, license, requires_python |
| downloads | `GET /api/packages/{package}/recent` | weekly_downloads, monthly_downloads |

**エラーハンドリング**: npm パターンに従う。metadata job が失敗したら全メトリクスが nil。downloads job のみ失敗なら partial result (ダウンロード数 0 + エラーメッセージ)。

### 3. CLI 登録

`internal/cli/cli.go` で PyPI プロバイダーを生成し、cache decorator でラップしてレジストリに登録する。

### 4. Markdown フォーマッター拡張

`internal/format/format.go` に PyPI 用の switch case と Markdown テーブルレンダリングを追加する。

## Backend Spec

### API エンドポイント

**PyPI JSON API**

```
GET https://pypi.org/pypi/{package}/json

Response (抜粋):
{
  "info": {
    "version": "2.32.3",
    "license": "Apache-2.0",
    "requires_python": ">=3.8",
    "requires_dist": ["certifi>=2017.4.17", "charset-normalizer<4,>=2", ...]
  },
  "releases": {
    "2.32.3": [{"upload_time_iso_8601": "2024-05-29T..."}]
  }
}
```

- `last_publish_days`: `releases[info.version]` の最新 `upload_time_iso_8601` から算出
- `dependencies_count`: `requires_dist` の要素数 (extras 条件 `; extra ==` を含むものを除外)
- `license`: `info.license` をそのまま使用

**pypistats.org API**

```
GET https://pypistats.org/api/packages/{package}/recent

Response:
{
  "data": {
    "last_day": 40574147,
    "last_week": 269467251,
    "last_month": 1081975948
  }
}
```

- `weekly_downloads`: `data.last_week`
- `monthly_downloads`: `data.last_month`

## Tracking

| Event Name | Properties | Trigger Condition |
|------------|------------|-------------------|
| N/A | CLI ツールのためトラッキングなし | - |
