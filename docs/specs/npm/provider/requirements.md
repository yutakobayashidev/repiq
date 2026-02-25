# Requirements: npm Provider

## Functional Requirements

### P1 (Must have)

- `npm:<package>` ターゲット形式をサポートする
- scoped package (`npm:@types/node`, `npm:@babel/core`) をサポートする
- 以下の 6 メトリクスを取得・返却する:
  - `weekly_downloads` (int) — 過去 7 日間のダウンロード数
  - `monthly_downloads` (int) — 過去 30 日間のダウンロード数 (`/downloads/point/last-month` エンドポイント)
  - `latest_version` (string) — 最新バージョン文字列
  - `last_publish_days` (int) — 最終パブリッシュからの経過日数
  - `dependencies_count` (int) — latest バージョンの依存パッケージ数
  - `license` (string) — ライセンス識別子
- 存在しないパッケージに対して `error` フィールドにメッセージを返す (Go error ではなく Result.Error)
- 既存の 3 つの出力フォーマット (JSON / NDJSON / Markdown) で npm メトリクスを表示する
- `repiq github:facebook/react npm:react` のように GitHub ターゲットとの混合一括取得が動作する
- Provider インターフェースに破壊的変更を加えない

### P2 (Should have)

- Markdown 出力で npm 結果を GitHub とは別テーブルにグルーピングする (既存の scheme 別グルーピングパターンに従う)
- 部分エラー時に取得できたメトリクスは返却し、失敗したメトリクスのみ error に記録する

### P3 (Nice to have)

- `--verbose` フラグで API リクエストの詳細ログを出力する (将来対応可)

## Non-Functional Requirements

- 単一ターゲットの取得を 3 秒以内に完了する
- 外部依存を追加しない (net/http + encoding/json のみ)
- httptest ベースのユニットテストで主要パスをカバーする
- golangci-lint を pass する

## Edge Cases

1. **scoped package の URL エンコーディング**: `@scope/name` を API リクエスト URL に含める際、`@scope%2Fname` のようにスラッシュをエンコードする必要がある (downloads API)
2. **license フィールドの型バリエーション**: npm パッケージの license は string の場合と object `{"type":"MIT"}` の場合がある。両方をハンドリングする
3. **dependencies が null/未定義**: 依存がないパッケージは `dependencies` フィールド自体が存在しない。この場合 `dependencies_count: 0` を返す
4. **unpublished パッケージ**: パッケージが unpublish された場合、レジストリ API は特殊なレスポンスを返す。404 同様にエラーハンドリングする
5. **ダウンロード数が 0**: 新規パッケージや private パッケージで downloads API が 0 を返す場合、そのまま 0 を返す
6. **ネットワークタイムアウト**: 3 並列リクエストは独立して完了を待つ。タイムアウトしたリクエストのメトリクスのみゼロ値とし、成功したリクエストのメトリクスは返却する (P2 の部分エラー要件と整合)

## Constraints

- Provider インターフェース (`Scheme()`, `Fetch()`) に変更を加えてはならない
- Result 構造体に `NPM *NPMMetrics` フィールドを追加する形で拡張する
- npm レジストリ API は認証不要 (auth モジュールとの統合は不要)
- GitHub provider の実装パターン (goroutine + sync.WaitGroup) に倣う
