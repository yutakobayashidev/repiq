# Requirements: Go Modules プロバイダー

## Functional Requirements

### P1 (Must have)

- `go:<module>` ターゲット形式で Go モジュールのメトリクスを取得できる
- 以下のメトリクスを返す:
  - `latest_version` - 最新バージョン (proxy.golang.org)
  - `last_publish_days` - 最終パブリッシュからの日数 (proxy.golang.org)
  - `dependencies_count` - 直接依存モジュール数 (deps.dev)
  - `license` - ライセンス (deps.dev)
- 存在しないモジュールで `"go proxy: 404 Not Found"` 形式のエラーメッセージを返す
- 不正なモジュールパスで `"invalid Go module path \"...\""` 形式のバリデーションエラーを返す
- 既存の出力フォーマット (JSON / NDJSON / Markdown) で Go Modules メトリクスが表示される
- 混合ターゲット一括取得が動作する

### P2 (Should have)

- proxy.golang.org でバージョン取得後、deps.dev API の 2 エンドポイントを並列リクエストする 2 段階フローでレスポンスタイムを最小化する (deps.dev エンドポイントにバージョン番号が必要なため)
- deps.dev API が失敗しても proxy.golang.org の結果を部分的に返す (partial results)

### P3 (Nice to have)

- なし

## Non-Functional Requirements

- 単一ターゲット取得が 3 秒以内
- ユニットテストは httptest mock ベースで proxy.golang.org と deps.dev の両方をモックする
- golangci-lint が pass する
- deps.dev API (v3alpha) の変更に対してグレースフルに劣化する (エラー時は該当メトリクスを空値で返す)

## Edge Cases

1. モジュールパスにメジャーバージョンサフィックスがある (例: `github.com/user/repo/v2`) → そのまま proxy に渡す
2. proxy.golang.org が `410 Gone` を返す (モジュールが取り下げられた) → エラーメッセージとして返す
3. deps.dev API がタイムアウトまたは 5xx → `dependencies_count: 0`, `license: ""` + partial error
4. deps.dev にモジュールが登録されていない → `dependencies_count: 0`, `license: ""` + partial error
5. バージョンが 1 つもない (proxy の `@latest` が 404) → エラーメッセージのみ返す
6. `latest_version` が `v0.0.0-...` 形式の疑似バージョン → そのまま返す (フィルタリングしない)

## Constraints

- proxy.golang.org は認証不要 (Google CDN 経由)
- deps.dev API は v3alpha (不安定) — API スキーマが変更される可能性がある
- Go モジュールパスはドメインを含む (例: `golang.org/x/text`, `github.com/gorilla/mux`)
- Go エコシステムにはダウンロード数の公開 API が存在しない
- deps.dev の依存関係データはバージョン指定が必要 (`/versions/{version}:dependencies`)
