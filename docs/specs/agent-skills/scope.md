# Scope: Agent Skills

## In Scope

- repiq 用の SKILL.md ファイル作成 (Agent Skills 標準フォーマット準拠)
- 全 5 プロバイダー (GitHub, npm, PyPI, crates.io, Go Modules) の使い方をカバー
- `skills/repiq/` ディレクトリとして repiq リポジトリ内に配置
- `npx skills add github:yutakobayashidev/repiq` でのインストール対応
- agent-skills-nix での宣言的インストール対応 (dotnix への設定追加)
- skills-ref による SKILL.md バリデーション通過

## Out of Scope

- MCP Server としての実装 (別エピック)
- 新しいプロバイダーの追加 (既存 5 プロバイダーのみ対象)
- repiq の CLI インターフェース変更 (現行 CLI をそのまま利用)
- エージェント側の実装変更 (標準フォーマットに従うだけ)
- スキルの有料配布・ライセンス制限

## Success Criteria (KPI)

### Expected to Improve

- repiq の発見性: Agent Skills 対応エージェントからの自然な利用開始
- 導入障壁の低減: `npx skills add` / `nix run` によるワンコマンドインストール
- エージェントの repiq 利用成功率: SKILL.md の指示に従った正確なコマンド実行

### At Risk (may decrease)

- 特になし (既存 CLI に変更を加えないため副作用リスクは低い)

## Acceptance Gates

- [ ] `skills/repiq/SKILL.md` が Agent Skills 仕様に準拠している (skills-ref validate 通過)
- [ ] `npx skills add github:yutakobayashidev/repiq` でインストール可能
- [ ] Claude Code でスキルが発見・発動し、`repiq github:<repo>` 等が正しく実行される
- [ ] agent-skills-nix 経由で Nix 環境にインストール可能
- [ ] SKILL.md が 500 行以内に収まっている (progressive disclosure 準拠)

## Experiment Info (if applicable)

- N/A (実験ではなく、標準フォーマットへの対応)
