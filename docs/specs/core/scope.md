# Scope: Core

## In Scope

- CLI エントリポイント (`main.go`)、引数パース (標準ライブラリのみ)
- `<scheme>:<identifier>` 入力フォーマットのパース
- `Provider` インターフェース定義
- GitHub プロバイダー実装 (google/go-github クライアント)
  - stars, forks, last_commit_days, contributors, open_issues, release_count
  - commits_30d, issues_closed_30d
- 出力フォーマッター: `--json`, `--ndjson`, `--markdown` (デフォルト)
- `--markdown` は異なるスキームが混在する場合、スキームごとにテーブルを分離して出力 (kubectl 方式)
- 複数ターゲットの goroutine 並列取得 (スキームの混在も制限なし)
- GitHub 認証: 未認証でも動作 (60 req/hour)。`gh auth token` → `GITHUB_TOKEN` の順でトークンを探し、あれば rate limit 緩和 (5,000 req/hour)
- エラー出力 (stderr に構造化エラー)
- ユニットテスト

## Out of Scope

- npm / crate / pypi / local プロバイダー (別 Epic)
- OpenSSF Scorecard 統合
- ローカルキャッシュ
- Agents Skills
- CI/CD パイプライン構築
- サブコマンド (現時点では不要)
- 設定ファイル

## Success Criteria (KPI)

### Expected to Improve

- AI エージェントが GitHub リポジトリの客観データを取得できるようになる (0 -> 1)
- 単一リポジトリのメトリクス取得が 3 秒以内で完了する

### At Risk (may decrease)

- 特になし (新規プロジェクトのため)

## Acceptance Gates

- [ ] `repiq github:facebook/react` が正しい JSON を返す
- [ ] `repiq github:facebook/react --ndjson` が NDJSON を返す
- [ ] `repiq github:facebook/react --markdown` が Markdown テーブルを返す
- [ ] `repiq github:facebook/react github:vercel/next.js` が並列取得で結果を返す
- [ ] トークンなしでも動作する (未認証モード)
- [ ] `gh auth token` / `GITHUB_TOKEN` があれば自動的に認証モードで動作する
- [ ] `repiq invalid-input` が適切なエラーを返す
- [ ] `go test ./...` が全てパスする
- [ ] `golangci-lint run` がパスする

## Experiment Info (if applicable)

N/A (新規プロジェクトの初期実装)
