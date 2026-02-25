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
