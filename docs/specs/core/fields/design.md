# Design: フィールド選択

## Current State

repiq は全プロバイダーから全フィールドを取得し、全フィールドを出力する。フィールドの選択・フィルタリング機能はない。

パイプライン:

```
CLI args → Target parse → Provider fetch (parallel) → Format → stdout
```

## Proposed Changes

取得後・フォーマット前にフィールドフィルタリングステップを挿入する。

パイプライン (変更後):

```
CLI args → Target parse → Provider fetch (parallel) → Field filter → Format → stdout
```

## CLI インターフェース

```
repiq [--fields field1,field2,...] [--json|--ndjson] target...
```

- `--fields` はカンマ区切りのフィールド名リスト
- 省略時は全フィールド出力 (既存動作)
- `-f` をショートエイリアスとして提供

## フィルタリング対象のフィールド名

各プロバイダーの JSON キー名がそのままフィールド名になる:

| Provider | Available fields |
|----------|-----------------|
| GitHub | `stars`, `forks`, `open_issues`, `contributors`, `release_count`, `last_commit_days`, `commits_30d`, `issues_closed_30d`, `license` |
| npm | `weekly_downloads`, `monthly_downloads`, `latest_version`, `last_publish_days`, `dependencies_count`, `license` |
| PyPI | `weekly_downloads`, `monthly_downloads`, `latest_version`, `last_publish_days`, `dependencies_count`, `license`, `requires_python` |
| crates.io | `downloads`, `recent_downloads`, `latest_version`, `last_publish_days`, `dependencies_count`, `license`, `reverse_dependencies` |
| Go | `latest_version`, `last_publish_days`, `dependencies_count`, `license` |

`target` と `error` は予約フィールドで常に出力に含まれる。

## フィルタリングの仕組み

`--fields` が指定された場合、各 Result のメトリクス構造体に対してフィールドフィルタを適用する。

- Go の struct tag (`json:"field_name"`) を使ってフィールド名とフィールドを対応付ける
- 指定されていないフィールドはゼロ値にリセットする (int → 0, string → "")
- JSON 出力: `omitempty` がないため全フィールドが出力される → フィルタリングではゼロ値リセットではなく、JSON マーシャル時に動的にフィールドを除外する必要がある

### JSON/NDJSON 出力のフィルタリング

フィールド選択時は `json.Marshal` の代わりに `map[string]any` に変換してから出力する:

1. Result を `map[string]any` に変換
2. `target`, `error` は常に保持
3. プロバイダーメトリクス (GitHub, npm 等) も `map[string]any` に変換
4. 指定フィールド以外のキーを削除
5. 空になったプロバイダーは `omitempty` 相当で除外

### Markdown 出力のフィルタリング

テーブルヘッダーとデータ行の生成時に、指定フィールドのみ列として出力する。

## 出力例

### `repiq --fields stars,license github:facebook/react npm:react`

**JSON:**

```json
[
  {
    "target": "github:facebook/react",
    "github": {
      "stars": 215000,
      "license": "MIT"
    }
  },
  {
    "target": "npm:react",
    "npm": {
      "license": "MIT"
    }
  }
]
```

**Markdown:**

```
| target | stars | license |
|---|---|---|
| github:facebook/react | 215000 | MIT |

| target | license |
|---|---|
| npm:react | MIT |
```

### `repiq --fields stars npm:react` (npm に stars はない)

```json
[
  {
    "target": "npm:react",
    "npm": {}
  }
]
```

## 変更対象ファイル

| ファイル | 変更内容 |
|---------|---------|
| `internal/cli/cli.go` | `--fields` / `-f` フラグ追加、フィルタリング呼び出し |
| `internal/format/format.go` | フィールドリスト受け取り、動的カラム生成 |
| `internal/format/format_test.go` | フィールド選択テスト追加 |
| `internal/cli/cli_test.go` | `--fields` フラグのテスト追加 |

## Tracking

| Event Name | Properties | Trigger Condition |
|------------|------------|-------------------|
| N/A | N/A | N/A |
