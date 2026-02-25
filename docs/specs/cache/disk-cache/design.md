# Design: ディスクキャッシュ

## Current State

repiq は `provider.Provider` インターフェースを介して GitHub / npm の API を叩き、`provider.Result` を返す。CLI (`internal/cli/cli.go`) がプロバイダーを `Registry` に登録し、各ターゲットを並列にフェッチしてフォーマッターに渡す。

現在キャッシュは存在せず、毎回フル API フェッチが発生する。

```
cli.Run
  → Registry.Lookup(scheme) → Provider.Fetch(ctx, id) → Result
  → format.JSON / NDJSON / Markdown
```

## Proposed Changes

プロバイダーと CLI の間にキャッシュレイヤーをデコレーターパターンで挿入する。既存のプロバイダーコードは変更しない。

```
cli.Run
  → Registry.Lookup(scheme) → CachingProvider.Fetch(ctx, id)
      → cache hit?  → return cached Result
      → cache miss? → underlying.Fetch(ctx, id) → store Result → return
  → format.JSON / NDJSON / Markdown
```

### 新規パッケージ: `internal/cache/`

キャッシュストアとデコレータープロバイダーを提供する。

### キャッシュストア

- **保存先**: `os.UserCacheDir()/repiq/`
- **形式**: 1 エントリ = 1 JSON ファイル
- **ファイル名**: `scheme:identifier` の SHA-256 ハッシュ (hex) + `.json`。単純な文字置換では `npm:@a/b_c` と `npm:@a_b/c` が衝突するため、暗号学的ハッシュで一意性を保証する (例: `sha256("github:facebook/react").json`)
- **ファイル内容**:
  ```json
  {
    "version": 2,
    "cached_at": "2026-02-25T09:00:00Z",
    "result": {
      "target": "github:facebook/react",
      "github": { ... }
    }
  }
  ```
- **スキーマバージョン判定**: `version` がコード側の `schemaVersion` 定数と一致しなければキャッシュミス。メトリクス構造体にフィールドが追加・削除・改名された場合、`schemaVersion` をインクリメントすることで古いキャッシュを自動無効化する
- **TTL 判定**: `cached_at` + 24h > 現在時刻 → ヒット
- **書き込み**: 一時ファイル (`*.tmp`) に書き込み後 `os.Rename` でアトミックに配置
- **エラー結果の扱い**: `Result.Error` が空でない場合はキャッシュに書き込まない

### デコレータープロバイダー

- `provider.Provider` インターフェースを実装
- 内部に `underlying Provider` と `cache Store` を保持
- `noCache` フラグで読み取りバイパスを制御

### CLI の変更

- `--no-cache` フラグを追加 (bool、デフォルト false)
- プロバイダー登録時にキャッシュデコレーターでラップ
- `os.UserCacheDir()` 取得失敗時はデコレーターをスキップし、生のプロバイダーを登録

## Backend Spec

### データフロー

1. CLI がフラグをパースし、`--no-cache` を取得
2. `os.UserCacheDir()` でキャッシュディレクトリパスを解決
3. `cache.NewStore(cacheDir)` でストアを生成
4. 各プロバイダーを `cache.NewProvider(underlying, store, noCache)` でラップし Registry に登録
5. 並列フェッチ時、`CachingProvider.Fetch()` が呼ばれる:
   - `noCache == false` の場合、ストアからキーで検索
   - バージョン不一致 → キャッシュミスとして扱う
   - ヒット & TTL 内 → キャッシュの Result を返却
   - ミスまたは TTL 切れ → underlying.Fetch() を呼び出し
   - 結果に Error がなければストアに書き込み
   - 結果を返却

### キャッシュキー

`scheme:identifier` をそのままキーとして使用。ファイル名への変換はストア内部で `crypto/sha256` ハッシュにより行う。衝突のない可逆でないが一意なマッピングを保証する。

| Target | Cache Key | File Name |
|--------|-----------|-----------|
| `github:facebook/react` | `github:facebook/react` | `a1b2c3...64hex.json` |
| `npm:@types/node` | `npm:@types/node` | `d4e5f6...64hex.json` |
| `npm:@a/b_c` | `npm:@a/b_c` | `(unique hash).json` |
| `npm:@a_b/c` | `npm:@a_b/c` | `(different hash).json` |

## Tracking

| Event Name | Properties | Trigger Condition |
|------------|------------|-------------------|
| 該当なし | - | キャッシュは透過的に動作し、ユーザー向けのイベントは発生しない |
