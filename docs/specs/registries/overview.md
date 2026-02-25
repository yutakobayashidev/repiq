# Registries (PyPI / crates.io / Go Modules)

## Summary

PyPI・crates.io・Go Modules の 3 プロバイダーを追加し、repiq を Python・Rust・Go エコシステムに対応させる。

## Background & Purpose

repiq は現在 GitHub と npm の 2 プロバイダーに対応しているが、AI エージェントがライブラリ選定を行う際、対象言語のパッケージレジストリのメトリクスがなければ判断材料として不十分である。

具体的な課題:

- **言語カバレッジの不足**: Python / Rust / Go は主要な OSS 言語だが、repiq で対応するパッケージレジストリがない。エージェントは GitHub メトリクスだけで判断するか、repiq 以外のツールを使う必要がある
- **npm 偏重の利用指標**: weekly_downloads は npm にしかなく、他言語のパッケージ利用実態を把握できない
- **Provider パターンの未検証**: 3 つ目以降のプロバイダーが既存の Provider インターフェースにスムーズに追加できるか実証されていない

各レジストリの API 特性:

| レジストリ | ダウンロード数 | 依存関係 | ライセンス | 認証 |
|-----------|-------------|---------|----------|------|
| **PyPI** | pypistats.org (週/月) | `requires_dist` | あり | 不要 |
| **crates.io** | total + 90 日間 | `/dependencies` | SPDX | 不要 (User-Agent 必須) |
| **Go Modules** | **取得不可** | `go.mod` / deps.dev | deps.dev | 不要 |

## Why Now

- cache epic が完了し、新規プロバイダーが自動的にキャッシュの恩恵を受けられる基盤が整った
- GitHub + npm の 2 プロバイダーで Provider パターンが確立されたが、3 つ目以降のスケーラビリティは未検証
- Python・Rust・Go は AI コーディングエージェントが頻繁に扱う言語であり、これらのレジストリ対応はユーザー価値に直結する

## Hypothesis

- **仮説 1 (マルチ言語対応の価値)**: PyPI / crates.io / Go Modules のメトリクスを追加することで、repiq が対応できるライブラリ選定シーンが JavaScript/TypeScript 以外にも拡大し、ツールとしての有用性が大幅に向上する
- **仮説 2 (Provider パターンのスケーラビリティ)**: 既存の Provider インターフェースと Result 型に PyPI / Crates / Go の 3 つを追加しても、インターフェースの破壊的変更なしに統合でき、今後のレジストリ追加も同パターンで実装可能であることが実証される

## Expected Outcome

- `repiq pypi:requests` で PyPI メトリクスが返る
- `repiq crates:serde` で crates.io メトリクスが返る
- `repiq go:golang.org/x/text` で Go Modules メトリクスが返る
- `repiq github:psf/requests pypi:requests` のような混合ターゲット一括取得ができる
- Provider インターフェースに破壊的変更なしで 3 プロバイダーが統合される
- 各プロバイダー単体で 3 秒以内のレスポンスタイム目標を維持する
