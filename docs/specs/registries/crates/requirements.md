# Requirements: crates.io プロバイダー

## Functional Requirements

### P1 (Must have)

- `crates:<crate>` ターゲット形式で crates.io のクレートメトリクスを取得できる
- 以下のメトリクスを返す:
  - `downloads` - 総ダウンロード数
  - `recent_downloads` - 直近 90 日間のダウンロード数
  - `latest_version` - 最新安定バージョン (`max_stable_version`)
  - `last_publish_days` - 最終パブリッシュからの日数
  - `dependencies_count` - 最新バージョンの依存クレート数 (`normal` kind のみ)
  - `license` - SPDX ライセンス識別子
  - `reverse_dependencies` - 被依存クレート数
- 存在しないクレートで `"crates.io: 404 Not Found"` 形式のエラーメッセージを返す
- 不正なクレート名で `"invalid crate name \"...\""` 形式のバリデーションエラーを返す
- 既存の出力フォーマット (JSON / NDJSON / Markdown) で crates.io メトリクスが表示される
- 混合ターゲット一括取得が動作する
- 全 HTTP リクエストに `User-Agent` ヘッダーを設定する (crates.io の要件)

### P2 (Should have)

- メタデータ取得後、依存関係取得と被依存数取得を並列実行する 2 段階フローでレスポンスタイムを最小化する (dependencies エンドポイントにバージョン番号が必要なため)
- 一部の API コールが失敗しても他の結果を部分的に返す (partial results)

### P3 (Nice to have)

- なし

## Non-Functional Requirements

- 単一ターゲット取得が 3 秒以内
- crates.io のレート制限 (1 req/sec 推奨) を考慮し、単一クレートへの並列リクエストは 3 本以内に抑える
- ユニットテストは httptest mock ベースで crates.io API をモックする
- golangci-lint が pass する

## Edge Cases

1. `max_stable_version` が null (全バージョンが pre-release) → `newest_version` にフォールバック
2. 最新バージョンの `license` が空文字列 → 空文字列をそのまま返す
3. 依存関係エンドポイントが失敗 → `dependencies_count: 0` + partial error
4. reverse_dependencies エンドポイントが失敗 → `reverse_dependencies: 0` + partial error
5. yanked されたクレート → 最新の non-yanked バージョン情報を返す (crates.io の `max_stable_version` が自動的に non-yanked を返す)
6. `versions` 配列が空 → エラーとして返す

## Constraints

- crates.io API は認証不要だが、`User-Agent` ヘッダーが必須
- `User-Agent` の値は `repiq/<version> (https://github.com/yutakobayashidev/repiq)` 形式
- crates.io のレート制限は約 1 req/sec (バースト許容あり)
- 依存関係 API はバージョン指定が必要 (`/crates/{crate}/{version}/dependencies`)
