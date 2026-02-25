# Implementation Plan: Agent Skills SKILL.md

Epic: #18 Agent Skills 対応: repiq を標準スキルとして公開
Feature: SKILL.md 作成 + README 更新

## Dependency Graph

```
Phase 1: [#1 REFERENCE.md] ─── [#2 SKILL.md]
                                     │
Phase 2: [#3 README.md] ────────────┘
                                     │
Phase 3: [#4 Validate] ─────────────┘
```

Parallel: #1/#2 (別ファイル、依存なし)

## Tasks

- [ ] #1 Create REFERENCE.md (詳細出力スキーマ)
  What: 5 プロバイダーの全メトリクスフィールドを Markdown テーブルで記載
  Where: `skills/repiq/references/REFERENCE.md` (新規)
  How: `internal/provider/provider.go` の struct 定義から JSON タグ・型・説明を抽出してテーブル化
  Why: SKILL.md の progressive disclosure のため、詳細スキーマを分離
  Verify: 全 31 フィールド (GitHub 8 + npm 5 + PyPI 7 + crates 7 + Go 4) がテーブルに含まれること
  Files: `skills/repiq/references/REFERENCE.md`
  Depends: (none)

- [ ] #2 Create SKILL.md (メインスキルファイル)
  What: Agent Skills 仕様準拠の SKILL.md を作成。frontmatter + 8 セクション (Overview, Quick Start, Schemes, Flags, Use Cases, Auth, Installation, Output Reference)
  Where: `skills/repiq/SKILL.md` (新規)
  How: design.md の frontmatter 定義に従い、`internal/cli/cli.go` のフラグ定義と `README.md` の使用例を参照して作成
  Why: エージェントが repiq を発見・発動・実行するための指示書
  Verify: 500 行以内であること。frontmatter に name/description/license/compatibility/metadata が含まれること
  Files: `skills/repiq/SKILL.md`
  Depends: (none)

- [ ] #3 Update README.md (Agent Skills セクション追加)
  What: README.md の `## Output Formats` と `## Development` の間に `## Agent Skills` セクションを追加
  Where: `README.md` (既存ファイル編集)
  How: design.md の README 更新仕様に従い、インストールコマンドを記載
  Why: ユーザーが Agent Skills 対応を発見できるようにする
  Verify: README.md に `## Agent Skills` セクションが存在し、`npx skills add` コマンドが記載されていること
  Files: `README.md`
  Depends: #1, #2

- [ ] #4 Validate SKILL.md (skills-ref バリデーション)
  What: skills-ref validate を実行し、仕様準拠を確認
  Where: リポジトリルート
  How: `npx skills-ref validate ./skills/repiq` または Python skills-ref CLI で実行
  Why: Acceptance Gate: skills-ref validate 通過が必須条件
  Verify: コマンドが exit 0 で終了し、エラーなし
  Files: (なし — 読み取りのみ)
  Depends: #1, #2
