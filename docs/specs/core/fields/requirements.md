# Requirements: フィールド選択

## Functional Requirements

### P1 (Must have)

- `--fields` フラグでカンマ区切りのフィールド名を受け取り、指定されたフィールドのみ出力する
  - 例: `repiq --fields stars,license github:facebook/react`
  - フィールド名は JSON キー名と一致する (例: `stars`, `weekly_downloads`, `license`)
- `--fields` 未指定時は全フィールドを出力する (既存動作と同一)
- 全出力フォーマット (JSON, NDJSON, Markdown) に適用される
- `target` と `error` は `--fields` の指定に関係なく常に出力に含まれる
- あるプロバイダーに存在しないフィールドが指定された場合、そのプロバイダーの結果ではサイレントに無視する
  - 例: `--fields stars,weekly_downloads` で GitHub と npm を取得した場合、GitHub 結果には `stars` のみ、npm 結果には `weekly_downloads` のみ表示

### P2 (Should have)

- `--fields` に完全に無効なフィールド名 (どのプロバイダーにも存在しない) が指定された場合、stderr に警告を出力する

### P3 (Nice to have)

- `-f` を `--fields` のショートエイリアスとして提供する

## Non-Functional Requirements

- `--fields` 未指定時のパフォーマンスに影響を与えない (フィルタリングロジックをスキップ)
- フィルタリングは取得後・フォーマット前に適用する (プロバイダーの API コールには影響しない)

## Edge Cases

1. `--fields ""` (空文字) — 全フィールド出力 (`--fields` 未指定と同じ)
2. `--fields stars,stars` (重複) — 重複を無視し、1回だけ出力
3. `--fields` のみで値なし — フラグパースエラー
4. 全ターゲットで指定フィールドがどれも該当しない場合 — `target` と `error` のみの結果を返す

## Constraints

- 出力の絞り込みのみ。API コールの最適化 (不要なフィールドの取得スキップ) は行わない
- フィールド名はフラットな JSON キー名のみ。ネストしたパス (`github.stars`) はサポートしない
