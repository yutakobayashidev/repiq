# Requirements: CLI + GitHub Provider

## Functional Requirements

### P1 (Must have)

- `repiq github:<owner>/<repo>` でリポジトリのメトリクスを取得できる
- 取得するメトリクス:
  - `stars` — スター数
  - `forks` — フォーク数
  - `open_issues` — オープン Issue 数 (GitHub API 仕様により PR を含む)
  - `contributors` — コントリビューター数 (Link ヘッダー推定)
  - `release_count` — リリース数
  - `last_commit_days` — 最終コミットからの経過日数
  - `commits_30d` — 直近 30 日のコミット数
  - `issues_closed_30d` — 直近 30 日にクローズされた Issue 数
  - `license` — SPDX ライセンス識別子 (`GET /repos` レスポンスの `license.spdx_id` から抽出、追加 API コール不要)
- 出力フォーマット:
  - `--markdown` (デフォルト): Markdown テーブル。スキーム混在時はスキーム別テーブル
  - `--json`: JSON 配列
  - `--ndjson`: 1行1レコード (改行区切り JSON)
- 複数ターゲットを引数に指定でき、並列で取得する
- 未認証でも動作する (GitHub API 60 req/hour)
- `gh auth token` があれば自動で認証モード (5,000 req/hour)
- `GITHUB_TOKEN` 環境変数でもトークンを受け付ける
- 認証の優先順位: `gh auth token` > `GITHUB_TOKEN` > 未認証

### P2 (Should have)

- `--help` でヘルプメッセージを表示
- `--version` でバージョン番号を表示
- rate limit に近づいた場合、stderr に警告を出す

### P3 (Nice to have)

- stderr にどの認証モードで動作しているかを表示 (debug 用)

## Non-Functional Requirements

- 単一リポジトリの取得が 3 秒以内に完了する
- 出力 JSON スキーマは deterministic (フィールド順固定)
- Provider インターフェースが定義され、新しいプロバイダーの追加が既存コードの変更なしに可能
- `golangci-lint run` がパスする

## Edge Cases

1. 存在しないリポジトリを指定した場合 → error フィールド付きの結果を返す
2. 複数ターゲットのうち一部が失敗 → 成功分は返し、失敗分は error フィールド付き。exit code 1
3. 全ターゲットが失敗 → 全て error フィールド付き。exit code 1
4. 全ターゲットが成功 → exit code 0
5. ターゲット引数なし → stderr にヘルプを表示。exit code 1
6. 不正な入力フォーマット (`github:` のないもの、コロンなし) → エラーメッセージ。exit code 1
7. GitHub API rate limit 超過 → error フィールドに rate limit エラーを含める
8. ネットワークエラー → error フィールドにタイムアウト/接続エラーを含める
9. `--json` と `--ndjson` と `--markdown` を同時指定 → 最後に指定されたものを採用

## Constraints

- CLI フレームワーク: 標準ライブラリのみ (flag パッケージ)
- GitHub クライアント: google/go-github
- Go 1.24
