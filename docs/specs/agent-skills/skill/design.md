# Design: SKILL.md 作成

## Current State

- repiq は Go 製 CLI として動作し、5 プロバイダー (GitHub, npm, PyPI, crates.io, Go Modules) のメトリクスを JSON/NDJSON/Markdown で出力する
- Agent Skills 対応はなく、エージェントが repiq を利用するには CLI の存在と使い方を事前に知っている必要がある
- README.md に CLI の使用方法は記載されているが、Agent Skills フォーマットではない

## Proposed Changes

### 1. ディレクトリ構成

```
skills/
└── repiq/
    ├── SKILL.md              # メインスキルファイル
    └── references/
        └── REFERENCE.md      # 詳細出力スキーマ
```

### 2. SKILL.md 構成

**Frontmatter:**

```yaml
name: repiq
description: >-
  Fetch objective metrics for OSS repositories and packages.
  Use when evaluating, comparing, or selecting libraries and repositories.
  Supports GitHub repos, npm/PyPI/crates.io packages, and Go modules.
  Returns stars, downloads, contributors, release activity, and more as structured JSON.
license: MIT
compatibility: Requires repiq CLI in PATH. Install via go install or Nix.
metadata:
  author: yutakobayashidev
  version: "1.0"
```

**Body セクション構成:**

1. **Overview** — repiq の目的 (data only, no judgments)
2. **Quick Start** — 基本的な呼び出し例
3. **Schemes** — 5 スキームの形式と具体例
4. **Flags** — `--markdown` (default), `--json`, `--ndjson`, `--no-cache`
5. **Use Cases** — ライブラリ比較、健全性評価のワークフロー例
6. **Authentication** — GitHub token 設定方法
7. **Installation** — `go install`, Nix による導入手順
8. **Output Reference** — REFERENCE.md へのリンク

### 3. REFERENCE.md 構成

各プロバイダーのメトリクスを Markdown テーブルで記載:

| Field | Type | Description |
|-------|------|-------------|
| `stars` | int | Total star count |
| ... | ... | ... |

対象:
- GitHubMetrics (8 fields)
- NPMMetrics (5 fields)
- PyPIMetrics (7 fields)
- CratesMetrics (7 fields)
- GoMetrics (4 fields)
- Error handling (Result.Error field)

### 4. README.md 更新

既存の README.md に以下のセクションを追加:

```markdown
## Agent Skills

repiq is available as an [Agent Skill](https://agentskills.io/) for AI coding agents.

### Install

npx skills add github:yutakobayashidev/repiq
```

### 5. npx skills add 対応

`npx skills add github:yutakobayashidev/repiq` は、リポジトリルートの `skills/` ディレクトリ内の `SKILL.md` を自動検出する。特別な設定ファイルは不要で、ディレクトリ構造が仕様に準拠していれば動作する。

## Backend Spec

N/A (CLI・API の変更なし)

## Tracking

| Event Name | Properties | Trigger Condition |
|------------|------------|-------------------|
| N/A | N/A | N/A |
