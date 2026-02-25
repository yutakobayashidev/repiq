# Requirements: SKILL.md 作成

## Functional Requirements

### P1 (Must have)

- `skills/repiq/SKILL.md` を Agent Skills 仕様に準拠して作成する
  - frontmatter: `name`, `description` (必須フィールド)
  - `name` は `repiq` (ディレクトリ名と一致)
  - `description` はエージェントが「いつこのスキルを発動すべきか」を判断できるトリガーキーワードを含む
- SKILL.md の本文に以下を含める:
  - repiq の目的 (データのみ提供、判断はしない)
  - CLI の呼び出し形式: `repiq [flags] <scheme>:<identifier> [...]`
  - 全 5 スキームの形式と例: `github:owner/repo`, `npm:package`, `pypi:package`, `crates:crate`, `go:module/path`
  - 出力フォーマットフラグ: `--json` (デフォルト), `--ndjson`, `--markdown`
  - `--no-cache` フラグの説明
  - 代表的なユースケース (ライブラリ比較、リポジトリ健全性評価)
- `skills/repiq/references/REFERENCE.md` に以下を分離:
  - 各プロバイダーの JSON 出力フィールド一覧と意味
  - GitHubMetrics (8 fields), NPMMetrics (5), PyPIMetrics (7), CratesMetrics (7), GoMetrics (4)
  - エラー出力の形式
- SKILL.md から REFERENCE.md へ相対パスで参照する
- `skills-ref validate ./skills/repiq` でバリデーションが通過する
- `npx skills add github:yutakobayashidev/repiq` でインストール可能である (skills/ ディレクトリがリポジトリルートに存在すれば自動検出される)

### P2 (Should have)

- SKILL.md の `description` フィールドにエージェントが反応しやすいキーワードを含める:
  - "repository", "library", "package", "metrics", "stars", "downloads", "compare", "evaluate", "OSS"
- frontmatter に `license` フィールドを含める (MIT)
- frontmatter に `compatibility` フィールドを含め、repiq CLI が PATH に必要な旨を記載
- frontmatter に `metadata` (author, version) を含める
- README.md に Agent Skills セクションを追加し、インストール方法を記載

### P3 (Nice to have)

- frontmatter の `allowed-tools` に `Bash(repiq:*)` を指定 (実験的フィールド)

## Non-Functional Requirements

- SKILL.md は 500 行以内 (progressive disclosure 準拠、推奨 5000 トークン以下)
- REFERENCE.md は構造化されたテーブル形式で、エージェントがフィールド名・型・意味を正確に把握できる
- 英語で記述 (Agent Skills は国際標準フォーマットのため)

## Edge Cases

1. repiq が PATH に存在しない場合 → SKILL.md にインストール方法 (`go install` / Nix) を記載し、エージェントがユーザーに案内できるようにする
2. GitHub 認証なしで実行した場合 → SKILL.md に rate limit の注意と `gh auth login` / `GITHUB_TOKEN` 設定を記載
3. 存在しないパッケージ・リポジトリを指定した場合 → JSON 出力の `error` フィールドの解釈方法を REFERENCE.md に記載
4. 複数ターゲット混在時 (github + npm + pypi) → 並列取得されること、結果が配列で返ることを明記

## Constraints

- 既存の repiq CLI に変更を加えない (SKILL.md とドキュメントのみ)
- Agent Skills 仕様 (agentskills.io/specification) に厳密に準拠する
- `name` フィールドはディレクトリ名 `repiq` と一致させる
