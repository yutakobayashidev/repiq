# Requirements: PyPI プロバイダー

## Functional Requirements

### P1 (Must have)

- `pypi:<package>` ターゲット形式で PyPI パッケージのメトリクスを取得できる
- 以下のメトリクスを返す:
  - `weekly_downloads` - 週間ダウンロード数 (pypistats.org)
  - `monthly_downloads` - 月間ダウンロード数 (pypistats.org)
  - `latest_version` - 最新バージョン
  - `last_publish_days` - 最終パブリッシュからの日数
  - `dependencies_count` - 依存パッケージ数 (`requires_dist` から算出、extras を除く)
  - `license` - ライセンス
  - `requires_python` - Python バージョン要件 (例: `>=3.9`)
- 存在しないパッケージで `"pypi registry: 404 Not Found"` 形式のエラーメッセージを返す
- 不正なパッケージ名で `"invalid PyPI package name \"...\""` 形式のバリデーションエラーを返す
- 既存の出力フォーマット (JSON / NDJSON / Markdown) で PyPI メトリクスが表示される
- 混合ターゲット一括取得が動作する (`repiq github:psf/requests pypi:requests`)

### P2 (Should have)

- PyPI JSON API と pypistats.org API への並列リクエストでレスポンスタイムを最小化する
- 一方の API が失敗しても他方の結果を部分的に返す (partial results)

### P3 (Nice to have)

- パッケージ名の正規化 (`Requests` → `requests`、`my_package` → `my-package` は PyPI 側で処理されるため特別な対応不要)

## Non-Functional Requirements

- 単一ターゲット取得が 3 秒以内
- ユニットテストは httptest mock ベースで PyPI JSON API と pypistats.org の両方をモックする
- golangci-lint が pass する
- キャッシュレイヤーとの統合は自動 (既存の cache decorator パターンを利用)

## Edge Cases

1. `requires_dist` が `null` のパッケージ → `dependencies_count: 0` を返す
2. `requires_dist` に extras 条件付き依存が含まれる (例: `socks ; extra == "socks"`) → extras 依存を除外してカウントする
3. `license` フィールドが空で `classifiers` にライセンス情報がある → `license` フィールドを優先し、空なら空文字列を返す (classifiers のパースは行わない)
4. pypistats.org API がダウンしている → ダウンロード数は 0、他のメトリクスは正常に返す (partial result + error message)
5. PyPI JSON API がダウンしている → 全メトリクスが取得不可、エラーメッセージのみ返す
6. `releases` が空 (パッケージが存在するがリリースなし) → `last_publish_days: 0`、`latest_version` は `info.version` から取得

## Constraints

- PyPI JSON API は認証不要 (レート制限は非公開だが CDN 経由)
- pypistats.org API は認証不要 (データは約 1 日遅延)
- `requires_dist` の依存は PEP 508 形式 (`package>=1.0; python_version>="3.8"`)。バージョン制約と環境マーカーを含むが、パッケージ名部分のみカウントに使用する
