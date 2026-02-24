# Session Context

## User Prompts

### Prompt 1

# Open Command

Define what this app is about and set the foundation for all future Epics.

This is the very first step: open the restaurant — decide what kind of place this is.

## When to Use

- Starting a new project from scratch
- No `docs/specs/overview.md` exists yet
- Updating the app's overall direction

## Workflow

### Step 1: Investigate Existing State

Use Grep, Glob, Read to understand:

- Does a codebase already exist? What does it do?
- Is there existing documentation (README, C...

### Prompt 2

Markdownの表を出力する機能は欲しいかも、人間用に

### Prompt 3

# Menu Command

Create a lightweight epic-level spec for the initiative described in ``.

This is the PM's first step: define WHAT to build and WHY.

## Workflow

### Step 0: Check Current Branch

Run `git branch --show-current` and present the current branch to the user via AskUserQuestion. Ask whether to proceed on the current branch or switch first. The worktree in Step 3 will branch from whatever branch you are on.

### Step 1: Investigate Codebase

Use Grep, Glob, Read to understand:

- Exi...

### Prompt 4

[Request interrupted by user]

### Prompt 5

# Menu Command

Create a lightweight epic-level spec for the initiative described in `core`.

This is the PM's first step: define WHAT to build and WHY.

## Workflow

### Step 0: Check Current Branch

Run `git branch --show-current` and present the current branch to the user via AskUserQuestion. Ask whether to proceed on the current branch or switch first. The worktree in Step 3 will branch from whatever branch you are on.

### Step 1: Investigate Codebase

Use Grep, Glob, Read to understand:

-...

### Prompt 6

[Request interrupted by user for tool use]

### Prompt 7

続けて

### Prompt 8

よくよく考えたけど、違うidentifier同士で比較された場合どすいたらいいんだろ？

### Prompt 9

ベストプラクティスはなんだろう

### Prompt 10

kubectl get pods,servicesって何？

### Prompt 11

じゃあそうしよう

### Prompt 12

じゃあして

### Prompt 13

[Request interrupted by user]

### Prompt 14

docsは一旦コミットせず、specはPRにしたい

### Prompt 15

一旦mainに戻り、commitlintでfeatとかのよくあるやつの規約を守るようにしてほしい、flakeで設定できる？

### Prompt 16

gitleaksもおねがい

### Prompt 17

github cliでgitignore初期化して、go

### Prompt 18

うん

### Prompt 19

push

### Prompt 20

a,.pre-commit-config.yaml ignore force push

### Prompt 21

さっきのPRのコミットメッセージ変えてforce push、rebase

### Prompt 22

[Request interrupted by user for tool use]

### Prompt 23

$ git config --add wt.hook 'direnv allow && eval "$(direnv export bash 2>/dev/null)" && pnpm install' これ設定するといい，gopウロジェクトに合わせていい感じに

### Prompt 24

いいね、では続けて

### Prompt 25

docs(specs)みたいにサブのやつを必須にしたいんだが、できるか？

### Prompt 26

[Request interrupted by user for tool use]

### Prompt 27

なんかもう少し綺麗にやる方法あるはず、調べて

### Prompt 28

うーんじゃあそれやめて、AGENTS.mdに英語で規約書くようにして。CLAUDE/mdmもsym link

### Prompt 29

[Request interrupted by user for tool use]

### Prompt 30

いやそれは残したままでいい

### Prompt 31

いいね、コミットして

### Prompt 32

yes

### Prompt 33

feature/coreのコメントも直して

### Prompt 34

GitHub CLIと上手く統合して、ユーザーはトークンの明示なしで機能するのが理想だな

### Prompt 35

トークンなくても一応レートリミットつよいけどうごうには動くはず

### Prompt 36

ダウンロード数は結構大事な気がする

### Prompt 37

一旦いいか

### Prompt 38

次はどうすればいいんだっけ

### Prompt 39

もうマージした /kitchen:order

### Prompt 40

[Request interrupted by user]

### Prompt 41

# Order Command

Takes a spec directory (e.g., `docs/specs/pdf-export`) or epic name via `` and creates an **Epic Issue** on GitHub.

This is the PM's second step: pass the order to the engineering team.

## Workflow

### Step 0: Check Current Branch

Run `git branch --show-current` and present the result to the user. If not on a feature branch, use AskUserQuestion to ask whether to proceed on the current branch or switch to the correct worktree first.

### Step 1: Read the Spec

1. Read `overvi...

### Prompt 42

# Recipe Command

You are a requirements specialist who reads the issue/spec, investigates the codebase, clarifies ambiguities, and writes detailed feature-level specifications.

## Your Skills

Read and follow these skill documents when working:

- `registry/skills/coding-standards.md` — Read to understand project coding conventions. Ground requirements in these standards.
- `registry/skills/backend-patterns.md` — Read when specifying backend features (API, DB, server-side).
- `registry/ski...

### Prompt 43

これも一旦PRに

### Prompt 44

スカッシュマージなので、PRタイトルもコミットメッセージの規格に対応するように. claude.md

