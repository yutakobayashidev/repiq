# Lessons

## 用語

### LRU eviction (Least Recently Used eviction)

キャッシュが容量上限に達したとき、**最も長い間使われていないエントリから順に削除する**方式。

- LRU = Least Recently Used (最も最近使われていないもの)
- eviction = 追い出し、削除
- 例: 上限 100 エントリのキャッシュで 101 個目を書き込むとき、最後にアクセスされた時刻が最も古いエントリを削除して空きを作る
- 対になる概念: LFU (Least Frequently Used) = 使用頻度が最も低いものを削除
- repiq では TTL ベースの期限切れで十分なため Out of Scope としている

## バグ

### deps.dev API: `:dependencies` ではなく `:requirements` を使う (Go Modules)

**何が起きたか:**
Go Modules プロバイダで deps.dev の依存関係取得に `:dependencies` エンドポイントを使用した。手動テストで 404 が返り、`dependencies_count` が常に 0 になった。

**原因:**
deps.dev v3alpha API には Go パッケージ向けに2つの類似エンドポイントがある:
- `/v3alpha/systems/go/packages/{pkg}/versions/{ver}:dependencies` — 存在しない (404)
- `/v3alpha/systems/go/packages/{pkg}/versions/{ver}:requirements` — 正しいエンドポイント

設計時に API ドキュメントを十分に検証せず、npm 等の他レジストリからの類推で `:dependencies` を使ってしまった。

**レスポンス形式の違い:**
当初想定していた形式:
```json
{"nodes": [{"relation": "DIRECT", "versionKey": {...}}]}
```
実際の `:requirements` の形式:
```json
{"go": {"directDependencies": [{"name": "...", "requirement": "..."}]}}
```

**ルール:**
1. 外部 API を使う前に、必ず `curl` で実際のレスポンスを確認する
2. 設計ドキュメントに書いた API エンドポイントは実装前に疎通確認する
3. テストのモックだけで安心しない — モックは実 API と乖離している可能性がある

### リリースは CI に任せる — 手動で `gh release create` しない

**何が起きたか:**
`git tag v0.3.0 && git push origin v0.3.0` の後、手動で `gh release create` を実行した。しかし `.github/workflows/release.yml` で goreleaser が tag push をトリガーに自動リリースを作る CI が設定されていた。結果、goreleaser のリリースと手動リリースが衝突した。

**対応:**
手動リリースを `gh release delete` で削除し、goreleaser ワークフローを `gh run rerun` で再実行して復旧した。

**ルール:**
1. リリースを作る前に `.github/workflows/` を確認し、tag push トリガーの CI がないか確認する
2. goreleaser 等のリリース CI がある場合は `git tag` + `git push --tags` のみ行い、リリース作成は CI に任せる
3. 手動で `gh release create` するのはリリース CI が存在しない場合のみ

### JSON キャッシュにはスキーマバージョンを入れる

**何が起きたか:**
`monthly_downloads` (npm) と `license` (GitHub) フィールドを追加した後、キャッシュから返されるレスポンスでこれらのフィールドが常にゼロ値 (`0` / `""`) になっていた。`--no-cache` を指定すると正しい値が返る状態。

**原因:**
キャッシュの on-disk JSON (`entry` 構造体) にスキーマバージョンがなかった。フィールド追加前に作成されたキャッシュエントリを `json.Unmarshal` すると、新フィールドは Go のゼロ値 (int=0, string="") になる。TTL 内であればそのまま有効なキャッシュヒットとして返されるため、ユーザーには「API は動いているのにデータが欠落している」ように見えた。

**修正:**
`entry` 構造体に `version` フィールドと `schemaVersion` 定数を追加。`Get` 時にバージョンが一致しなければキャッシュミスとして扱う。

**ルール:**
1. JSON でシリアライズされた構造体をキャッシュする場合、必ずスキーマバージョンを含める
2. 構造体にフィールドを追加・削除・改名したら `schemaVersion` をインクリメントする
3. キャッシュ起因のバグは `--no-cache` との比較で切り分ける — 値が変わるならキャッシュが原因
