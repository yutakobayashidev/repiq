# Scope: npm Provider

## In Scope

- `npm:<package>` ターゲット形式のパース (scoped package `@scope/name` を含む)
- npm レジストリ API からの以下のメトリクス取得:
  - `weekly_downloads` - 週間ダウンロード数
  - `latest_version` - 最新バージョン
  - `last_publish_days` - 最終パブリッシュからの日数
  - `dependencies_count` - 依存パッケージ数
  - `license` - ライセンス
- 既存の出力フォーマット (JSON / NDJSON / Markdown) での npm メトリクス表示
- `repiq github:X npm:Y` の混合ターゲット一括取得
- エラーハンドリング (存在しないパッケージ、private パッケージ等)
- ユニットテスト (httptest mock ベース)
- CI green (golangci-lint + テスト)

## Out of Scope

- `npm:<package>@<version>` 形式 (特定バージョン指定)
- ダウンロードトレンド (過去 N ヶ月の推移データ)
- deprecated 状態、type 定義有無、engines 互換性
- npm 認証 (npm token)
- ローカルキャッシュ

## Success Criteria (KPI)

### Expected to Improve

- AI エージェントが利用可能なメトリクスのソース数: 1 -> 2
- `repiq` 単体でカバーできるライブラリ評価観点: GitHub (活動指標) + npm (利用指標)

### At Risk (may decrease)

- 複数プロバイダー実行時の合計レスポンスタイム (並列実行で緩和)

## Acceptance Gates

- [ ] `repiq npm:react` で 5 メトリクスが JSON 出力される
- [ ] `repiq npm:@types/node` で scoped package が正常に取得される
- [ ] `repiq github:facebook/react npm:react` で混合ターゲットが一括取得される
- [ ] 存在しないパッケージで適切なエラーメッセージが返る
- [ ] `--ndjson` / `--markdown` フォーマットで npm メトリクスが表示される
- [ ] 単体テストが CI で pass する
- [ ] golangci-lint が pass する
- [ ] 単一ターゲット取得が 3 秒以内

## Experiment Info (if applicable)

N/A - フィーチャーフラグなし。npm プロバイダーは Provider レジストリへの登録で有効化される。
