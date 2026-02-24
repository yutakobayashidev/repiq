# Session Context

## User Prompts

### Prompt 1

# Menu Command

Create a lightweight epic-level spec for the initiative described in ``.

This is the PM's first step: define WHAT to build and WHY.

## Workflow

### Step 1: Investigate Codebase

Use Grep, Glob, Read to understand:

- Existing features related to this initiative
- Current architecture and patterns
- Potential impact areas

### Step 2: Clarify Requirements

Use AskUserQuestion iteratively:

- 2-4 questions per round
- 2-4 concrete options per question
- Focus categories:
  - **W...

### Prompt 2

# Order Command

Takes a spec directory (e.g., `docs/specs/pdf-export`) or epic name via `` and creates an **Epic Issue** on GitHub.

This is the PM's second step: pass the order to the engineering team.

## Workflow

### Step 1: Read the Spec

1. Read `overview.md` and `scope.md` under `docs/specs/<epic>/`
2. If spec not found, suggest running `/kitchen:menu` first

### Step 2: Create Epic Issue

Compose the issue body:

- **Summarize** the spec (do not duplicate full content)
- **Link** to the...

### Prompt 3

# Recipe Command

You are a requirements specialist who reads the issue/spec, investigates the codebase, clarifies ambiguities, and writes detailed feature-level specifications.

## Your Skills

Read and follow these skill documents when working:

- `registry/skills/coding-standards.md` — Read to understand project coding conventions. Ground requirements in these standards.
- `registry/skills/backend-patterns.md` — Read when specifying backend features (API, DB, server-side).
- `registry/ski...

### Prompt 4

pr名英語に

### Prompt 5

レビューチェック

### Prompt 6

次のステップ

### Prompt 7

# Prep Command

You are a planning and decomposition specialist who turns feature specs into actionable implementation plans with explicit dependency graphs. Each task must be self-contained — executable from its description alone, with concrete verification steps.

## Your Skills

Read and follow these skill documents when working:

- `registry/skills/coding-standards.md` — Read to align implementation plan with project coding standards.
- `registry/skills/backend-patterns.md` — Read when...

### Prompt 8

続けて

### Prompt 9

# Cook Command

You are an implementation specialist who writes code following strict TDD methodology: write tests first, then implement, then refactor.

## Your Skills

Read and follow these skill documents when writing code:

- `registry/skills/coding-standards.md` — Read and follow for all code you write.
- `registry/skills/backend-patterns.md` — Read and follow when implementing backend code (API, DB, server-side).
- `registry/skills/frontend-patterns.md` — Read and follow when impleme...

### Prompt 10

# Serve Command

You are a delivery specialist who reviews code quality, cleans up AI-generated noise, and creates a polished Pull Request.

## Your Skills

Read and follow these skill documents when reviewing:

- `registry/skills/coding-standards.md` — Read and use as review criteria for code quality checks.
- `registry/deslop/skills/deslop/SKILL.md` — Read for patterns to identify and remove AI-generated code slop.

## Input

No arguments required. Reviews all changes on the current featur...

### Prompt 11

他のプロバイダー向けに、追加のためのドキュメントを書いてください

### Prompt 12

試しにビルドして動作をテストしてください

### Prompt 13

PRをready for reviewにして

### Prompt 14

レビューきた

### Prompt 15

This session is being continued from a previous conversation that ran out of context. The summary below covers the earlier portion of the conversation.

Analysis:
Let me chronologically analyze the conversation:

1. **`/kitchen:menu` - Epic Spec Creation**: User invoked the menu command to create an epic-level spec. Through investigation and Q&A, we determined it was for the npm provider. Created `docs/specs/npm/overview.md` and `docs/specs/npm/scope.md` in a worktree, committed, and created PR ...

### Prompt 16

もう1回手動実行してテストして

### Prompt 17

マージしました。次のステップは？

### Prompt 18

次はP2では？

