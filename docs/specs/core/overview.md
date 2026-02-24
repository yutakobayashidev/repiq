# Core: CLI + Provider Interface + GitHub Provider

## Summary

repiq の動作基盤となる CLI スケルトン、プロバイダーインターフェース、GitHub プロバイダーを実装する。これにより `repiq github:facebook/react` で GitHub リポジトリのメトリクスを JSON/NDJSON/Markdown で取得できる最小構成が完成する。

## Background & Purpose

repiq は AI エージェント向けの OSS データ取得 CLI だが、現在コードが一切存在しない。まず動く最小構成を作り、プロバイダーパターンを確立することで、後続の npm/crate/pypi 等の拡張を容易にする。

ユーザーの課題:
- AI エージェントがライブラリ選定に必要な客観データを素早く取得する手段がない
- 既存ツール (gh CLI, Scorecard 等) は個別に叩く必要があり、統一されたスキーマがない

## Why Now

プロジェクトの Day 1。全ての Epic はこの core の上に構築される。

## Hypothesis

- Hypothesis 1: If we provide GitHub metrics in a fixed JSON schema, then AI agents can make better-informed library selection decisions without additional data wrangling
- Hypothesis 2: If we keep response time under 3 seconds, then vibe coding workflows won't be interrupted by data fetching

## Expected Outcome

- `repiq github:<owner>/<repo>` で stars, forks, contributors, open_issues, release_count, last_commit_days, commits_30d, issues_closed_30d が取得できる
- `--json`, `--ndjson`, `--markdown` の 3 フォーマットで出力できる
- 複数ターゲットを goroutine で並列取得できる (異なるスキームの混在も許容)
- `--markdown` は異なるスキーム混在時にスキーム別テーブルで出力 (kubectl 方式)
- プロバイダーインターフェースが確立され、npm 等の追加が容易な状態になる
