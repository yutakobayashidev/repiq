# Scope: Registries (PyPI / crates.io / Go Modules)

## In Scope

### PyPI プロバイダー (`pypi:<package>`)

- PyPI JSON API (`pypi.org/pypi/{package}/json`) からのメタデータ取得
- pypistats.org API (`pypistats.org/api/packages/{package}/recent`) からのダウンロード数取得
- 取得メトリクス:
  - `weekly_downloads` - 週間ダウンロード数 (pypistats.org)
  - `monthly_downloads` - 月間ダウンロード数 (pypistats.org)
  - `latest_version` - 最新バージョン
  - `last_publish_days` - 最終パブリッシュからの日数
  - `dependencies_count` - 依存パッケージ数 (`requires_dist` から算出)
  - `license` - ライセンス
  - `requires_python` - Python バージョン要件

### crates.io プロバイダー (`crates:<crate>`)

- crates.io API (`crates.io/api/v1/crates/{crate}`) からのメトリクス取得
- 取得メトリクス:
  - `downloads` - 総ダウンロード数
  - `recent_downloads` - 直近 90 日間のダウンロード数
  - `latest_version` - 最新安定バージョン
  - `last_publish_days` - 最終パブリッシュからの日数
  - `dependencies_count` - 依存クレート数 (`/dependencies` エンドポイント)
  - `license` - SPDX ライセンス識別子
  - `reverse_dependencies` - 被依存クレート数 (`/reverse_dependencies` の `meta.total`)
- `User-Agent` ヘッダーの設定 (crates.io の要件)

### Go Modules プロバイダー (`go:<module>`)

- Go Module Proxy (`proxy.golang.org/{module}/@latest`) からのバージョン情報取得
- deps.dev API (`api.deps.dev/v3alpha/`) からの補足情報取得
- 取得メトリクス:
  - `latest_version` - 最新バージョン
  - `last_publish_days` - 最終パブリッシュからの日数
  - `dependencies_count` - 直接依存モジュール数 (deps.dev)
  - `license` - ライセンス (deps.dev)
- **注意**: Go エコシステムにはダウンロード数の公開 API が存在しないため、ダウンロード関連メトリクスは提供しない

### 共通

- 既存の出力フォーマット (JSON / NDJSON / Markdown) での各メトリクス表示
- 混合ターゲット一括取得 (`repiq github:X pypi:Y crates:Z go:W`)
- エラーハンドリング (存在しないパッケージ、API エラー等)
- 各プロバイダーのユニットテスト (httptest mock ベース)
- 既存キャッシュレイヤーとの自動統合
- CI green (golangci-lint + テスト)

## Out of Scope

- バージョン指定 (`pypi:requests@2.31.0`, `crates:serde@1.0`)
- ダウンロードトレンド (時系列データ)
- crates.io の `linecounts` (コードメトリクス)
- pkg.go.dev の HTML スクレイピング (importers_count 等)
- deps.dev の dependents 情報 (API 未安定)
- 各レジストリの認証トークン対応
- プロバイダーごとの TTL カスタマイズ

## Success Criteria (KPI)

### Expected to Improve

- 対応レジストリ数: 2 (GitHub, npm) -> 5 (+ PyPI, crates.io, Go Modules)
- AI エージェントが repiq でカバーできるライブラリ選定シーン: JS/TS -> Python, Rust, Go にも拡大
- Provider パターンの実証度: 2 プロバイダー -> 5 プロバイダーで汎用性を検証済み

### At Risk (may decrease)

- Result 型のフィールド数増加による JSON 出力の複雑化 (`omitempty` で緩和)
- deps.dev API (v3alpha) の不安定性による Go プロバイダーの一部メトリクス取得失敗リスク

## Acceptance Gates

- [ ] `repiq pypi:requests` で PyPI メトリクスが JSON 出力される
- [ ] `repiq crates:serde` で crates.io メトリクスが JSON 出力される
- [ ] `repiq go:golang.org/x/text` で Go Modules メトリクスが JSON 出力される
- [ ] `repiq github:psf/requests pypi:requests` で混合ターゲットが一括取得される
- [ ] 各プロバイダーで存在しないパッケージに適切なエラーメッセージが返る
- [ ] `--ndjson` / `--markdown` フォーマットで各プロバイダーのメトリクスが表示される
- [ ] 各プロバイダーのユニットテストが CI で pass する
- [ ] golangci-lint が pass する
- [ ] 各プロバイダー単体で 3 秒以内のレスポンスタイム
- [ ] キャッシュが全プロバイダーで透過的に動作する

## Experiment Info (if applicable)

N/A - フィーチャーフラグなし。各プロバイダーは Provider レジストリへの登録で有効化される。
