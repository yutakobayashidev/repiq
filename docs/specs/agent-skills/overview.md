# Agent Skills

## Summary

repiq を Agent Skills 標準フォーマットで公開し、Claude Code・Cursor・Gemini CLI 等のスキル対応エージェントから直接利用可能にする。

## Background & Purpose

repiq は「AI エージェントが OSS リポジトリ・ライブラリを選定するための客観メトリクス CLI」として設計されている。現在エージェントが repiq を使うには CLI をサブプロセスとして実行する必要があり、以下の課題がある:

- エージェントが repiq の存在や使い方を知らない (発見性の問題)
- CLI のインストール・セットアップが必要 (導入障壁)
- コマンド引数やスキーム形式をエージェントが正しく組み立てる保証がない (信頼性)

Agent Skills はエージェントに「いつ・どう使うか」を宣言的に伝える標準フォーマットであり、これを採用することで上記の課題をすべて解決できる。

## Why Now

- Agent Skills が Anthropic 発の標準として確立され、Claude Code・Cursor・Gemini CLI・GitHub Copilot 等 30+ のエージェント製品が対応済み
- repiq は 5 プロバイダー (GitHub, npm, PyPI, crates.io, Go Modules) の実装が完了し、スキルとして提供するに十分な機能が揃った
- agent-skills-nix による Nix ベースの宣言的配布も成熟し、Nix ユーザーへのリーチも可能

## Hypothesis

- H1: If we publish repiq as an Agent Skill, then agents (Claude Code, Cursor etc.) will automatically discover and use repiq when users ask about library selection or repository evaluation
- H2: If we provide clear instructions in SKILL.md, then agents will correctly construct repiq commands and interpret the JSON output without user intervention
- H3: If we distribute via both `npx skills add` and agent-skills-nix, then adoption friction drops to near zero for both npm and Nix users

## Expected Outcome

- repiq が Agent Skills 対応エージェントのエコシステムに参加し、ユーザーが意識せずとも repiq のデータをエージェント経由で利用可能になる
- ライブラリ選定タスクで repiq のメトリクスが自動的に参照され、データに基づく意思決定が促進される
- repiq の GitHub リポジトリへのトラフィック・Star が増加する
