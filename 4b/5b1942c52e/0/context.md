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

マージしました

### Prompt 3

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

### Prompt 4

# Recipe Command

You are a requirements specialist who reads the issue/spec, investigates the codebase, clarifies ambiguities, and writes detailed feature-level specifications.

## Your Skills

Read and follow these skill documents when working:

- `registry/skills/coding-standards.md` — Read to understand project coding conventions. Ground requirements in these standards.
- `registry/skills/backend-patterns.md` — Read when specifying backend features (API, DB, server-side).
- `registry/ski...

### Prompt 5

PR命は英語に

### Prompt 6

マージした。

### Prompt 7

# Prep Command

You are a planning and decomposition specialist who turns feature specs into actionable implementation plans with explicit dependency graphs. Each task must be self-contained — executable from its description alone, with concrete verification steps.

## Your Skills

Read and follow these skill documents when working:

- `registry/skills/coding-standards.md` — Read to align implementation plan with project coding standards.
- `registry/skills/backend-patterns.md` — Read when...

### Prompt 8

[Request interrupted by user for tool use]

### Prompt 9

続けていいですが、これは考慮しておいてください。レビュー来てました P1 Badge Use collision-free cache key serialization

The filename rule here is not one-to-one, so distinct valid targets can overwrite each other’s cache entries and later return the wrong Result. For example, npm identifiers @a/b_c and @a_b/c are both accepted by internal/provider/npm/npm.go’s package regex, but both collapse to npm__a_b_c.json under this replacement scheme. This breaks cache...

### Prompt 10

[Request interrupted by user]

### Prompt 11

専用のワークツリーに入るように

### Prompt 12

[Request interrupted by user for tool use]

### Prompt 13

専用のワークツリーで。Creating worktree

### Prompt 14

続けて

### Prompt 15

<bash-input>pwd</bash-input>

### Prompt 16

<bash-stdout>/home/yuta/ghq/github.com/yutakobayashidev/repiq</bash-stdout><bash-stderr></bash-stderr>

### Prompt 17

<bash-input>/kitchen:serve</bash-input>

### Prompt 18

<bash-stdout>(eval):1: no such file or directory: /kitchen:serve
</bash-stdout><bash-stderr>(eval):1: no such file or directory: /kitchen:serve
</bash-stderr>

### Prompt 19

# Serve Command

You are a delivery specialist who reviews code quality, cleans up AI-generated noise, and creates a polished Pull Request.

## Your Skills

Read and follow these skill documents when reviewing:

- `registry/skills/coding-standards.md` — Read and use as review criteria for code quality checks.
- `registry/deslop/skills/deslop/SKILL.md` — Read for patterns to identify and remove AI-generated code slop.

## Input

No arguments required. Reviews all changes on the current featur...

### Prompt 20

続けて

### Prompt 21

手動で実行してみて

### Prompt 22

LRU evictionってなに？

### Prompt 23

結果を本文に追記しておいて

### Prompt 24

LRUとは何かについても説明を追加しておいて、学習メモ

### Prompt 25

それmainにコミットしといて

