# Scope: repiq

## Current Phase

MVP (Minimum Viable Product)

## In Scope (this phase)

- CLI スケルトン (Go, `<scheme>:<identifier>` 入力パース)
- `github:<owner>/<repo>` プロバイダー (stars, forks, last_commit_days, contributors, open_issues, release_count, commits_30d, issues_closed_30d)
- `npm:<package>` プロバイダー (weekly_downloads, latest_version, last_publish_days, dependencies_count, license)
- `--json` / `--ndjson` / `--markdown` 出力フォーマット
- 複数ターゲットの一括取得
- GitHub 認証 (`gh auth token` 優先、`GITHUB_TOKEN` フォールバック)

## Out of Scope (this phase)

- OpenSSF Scorecard 統合
- crate / pypi / go modules プロバイダー
- local:// パス解析
- ローカルキャッシュ
- fast モード (500ms)
- Agents Skills
- プラグインシステム
- ランキング・推薦・スコアリング (永久に非対象)

## Recommended Epics

| Priority | Epic (slug) | One-line description | Why |
| -------- | ----------- | -------------------- | --- |
| P0 | `core` | CLI スケルトン + プロバイダーインターフェース + GitHub プロバイダー | 動く最小構成。これがないと何も始まらない |
| P1 | `npm` | npm レジストリプロバイダー | MVP スコープの後半。GitHub だけでは不十分 |
| P2 | `cache` | ローカルキャッシュレイヤー | 繰り返し実行の高速化。UX 改善に直結 |
| P3 | `registries` | crate / pypi / go modules プロバイダー追加 | レジストリ拡充。ビジョンの中核 |
| P4 | `scorecard` | OpenSSF Scorecard 統合 | セキュリティ指標の追加。エージェントの判断材料を拡充 |

## Technical Constraints

- 言語: Go 1.24
- 開発環境: Nix flake
- 外部依存: 最小限 (標準ライブラリ優先)
- API: GitHub REST/GraphQL API, npm registry API
- パフォーマンス: 通常リクエスト 3秒以内

## Success Criteria

- `repiq github:facebook/react` で GitHub メトリクスが JSON で返る
- `repiq npm:react` で npm メトリクスが JSON で返る
- `repiq github:facebook/react npm:react` で複数ターゲット一括取得できる
- `--ndjson` で 1行1レコード形式で出力できる
- `--markdown` で人間が読める Markdown テーブルで出力できる
- CI で `golangci-lint` が通る
- 基本的なユニットテストがある
