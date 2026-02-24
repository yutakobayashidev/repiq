# npm Provider

## Summary

npm レジストリから パッケージメトリクスを取得する npm プロバイダーを追加し、repiq を GitHub 単体のツールからマルチソース対応のデータ取得 CLI へ進化させる。

## Background & Purpose

repiq の core epic で GitHub プロバイダーと Provider インターフェースが確立された。しかし GitHub メトリクス単体では、AI エージェントがライブラリ選定を行う際に情報が不足する。

具体的には:

- **利用実態が見えない**: スター数が多くても実際に使われているかは分からない。npm の weekly downloads は実利用の直接的な指標になる
- **メンテナンス状況の別視点**: GitHub の last_commit_days はコード変更の頻度を示すが、last_publish_days はリリースの頻度を示す。両方あって初めてメンテナンス状況を正確に把握できる
- **依存関係の複雑さ**: dependencies_count はサプライチェーンリスクの簡易指標として機能する

また、2つ目のプロバイダーを実装することで、Provider インターフェースの拡張性を実証する。

## Why Now

- core epic (P0) が完了し、Provider インターフェースが確立された
- npm は MVP スコープに含まれており、GitHub の次に実装すべきプロバイダーとして P1 に位置づけられている
- Provider パターンの検証は、後続の crates.io / PyPI 等の拡張前に行うべき

## Hypothesis

- **仮説 1 (AI 判断材料の充実)**: GitHub メトリクスに npm メトリクス (weekly_downloads, last_publish_days 等) を加えることで、AI エージェントがライブラリの実利用状況とメンテナンス状況をより正確に評価でき、選定精度が向上する
- **仮説 2 (プロバイダー拡張性の検証)**: 既存の Provider インターフェースが2つ目のプロバイダー (npm) でも変更なしにスムーズに機能し、後続レジストリ追加の実用性が実証される

## Expected Outcome

- `repiq npm:react` で npm メトリクスが JSON で返る
- `repiq github:facebook/react npm:react` で GitHub + npm の一括取得ができる
- Provider インターフェースに破壊的変更なしで npm プロバイダーが統合される
- 3秒以内のレスポンスタイム目標を維持する
