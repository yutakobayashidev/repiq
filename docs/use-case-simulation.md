# ユースケースシミュレーション

repiq を AI エージェントのツールとして実際のライブラリ選定シナリオで使用し、有用性を検証した記録。

## シナリオ1: Node.js Web フレームワーク選定

「Express / Fastify / Hono のどれを使うべきか？」という質問を想定。

```bash
repiq github:expressjs/express github:fastify/fastify github:honojs/hono npm:express npm:fastify npm:hono
```

### 結果 (2026-02-25 時点)

**GitHub メトリクス:**

| target | stars | commits_30d | issues_closed_30d | contributors |
|---|---|---|---|---|
| expressjs/express | 68,811 | 11 | 8 | 320 |
| fastify/fastify | 35,676 | 24 | 15 | 436 |
| honojs/hono | 29,008 | 42 | 19 | 289 |

**npm メトリクス:**

| target | weekly_downloads | dependencies_count |
|---|---|---|
| npm:express | 71,783,894 | 28 |
| npm:fastify | 5,117,902 | 15 |
| npm:hono | 19,961,929 | 0 |

### 考察

- Hono は stars では Express の半分以下だが、commits_30d が42で最も活発。依存ゼロ。週間DL 2,000万は Fastify を抜いている
- Express は週間7,100万DLだが、commits_30d は11と少なめ。成熟して安定期
- Fastify はコントリビューター数436で最多。堅実な開発体制
- repiq なしだと「stars 多い = 良い」で Express に流れがち。commits_30d と依存数を並べるだけで判断の質が変わる

---

## シナリオ2: ORM のメンテナンス状況確認

「TypeORM ってまだメンテされてる？Prisma / Drizzle と比べてどう？」という質問を想定。

```bash
repiq github:prisma/prisma github:drizzle-team/drizzle-orm github:typeorm/typeorm npm:prisma npm:drizzle-orm npm:typeorm
```

### 結果 (2026-02-25 時点)

**GitHub メトリクス:**

| target | open_issues | release_count | commits_30d | issues_closed_30d |
|---|---|---|---|---|
| prisma/prisma | 2,488 | 241 | 53 | 42 |
| drizzle-team/drizzle-orm | 1,575 | 169 | 1 | 32 |
| typeorm/typeorm | 498 | 40 | 28 | 29 |

**npm メトリクス:**

| target | weekly_downloads | latest_version | dependencies_count |
|---|---|---|---|
| npm:prisma | 8,575,823 | 7.4.1 | 6 |
| npm:drizzle-orm | 5,010,846 | 0.45.1 | 0 |
| npm:typeorm | 3,354,650 | 0.3.28 | 15 |

### 考察

- TypeORM: open_issues 498 は一見少ないが、release_count がたった40。バージョンもまだ 0.3.x
- Prisma: open_issues 2,488 は多いが、issues_closed_30d も42で回している。ちゃんとトリアージされている
- Drizzle: commits_30d が1で直近は静か。ただ issues_closed_30d は32あるので、コードは落ち着いてるがサポートは動いている
- `open_issues` の絶対数ではなく `issues_closed_30d` との比率で判断できるのが良い

---

## シナリオ3: Rust HTTP クライアント選定

「reqwest 一択？他にある？」という質問を想定。

```bash
repiq github:seanmonstar/reqwest github:hyperium/hyper crates:reqwest crates:hyper crates:ureq
```

### 結果 (2026-02-25 時点)

**GitHub メトリクス:**

| target | stars | open_issues | commits_30d | contributors |
|---|---|---|---|---|
| seanmonstar/reqwest | 11,442 | 454 | 6 | 371 |
| hyperium/hyper | 15,951 | 245 | 7 | 405 |

**crates.io メトリクス:**

| target | downloads | recent_downloads | dependencies_count | reverse_dependencies |
|---|---|---|---|---|
| crates:reqwest | 378,971,846 | 62,091,703 | 47 | 18,030 |
| crates:hyper | 527,503,701 | 85,200,428 | 18 | 4,417 |
| crates:ureq | 96,090,536 | 19,602,392 | 23 | 1,457 |

### 考察

- reqwest の依存47個は多い。軽量が欲しいなら ureq (23個) が候補
- reverse_dependencies で reqwest 18,030 vs ureq 1,457。エコシステムの中心はやはり reqwest
- hyper はダウンロード数では最多だが、reqwest の内部依存として引っ張られている分が大きい

---

## 総合評価

### repiq が活きるポイント

- ライブラリ選定で主観でなく数値ベースの回答を組み立てられる
- `commits_30d` / `issues_closed_30d` / `dependencies_count` は GitHub ページをパッと見ただけでは読み取りにくい情報で、差別化になっている
- 1コマンドで GitHub + パッケージレジストリを横断できる

### 改善の余地

- **解釈はエージェント任せ**: commits_30d が1なのは「停滞」なのか「安定」なのかは文脈次第。設計思想として正しいが、単体ツールとしての訴求力は弱くなる
- **ライセンス比較の強調**: 企業での採用判断では最重要ファクターの一つ
- **トレンド（時系列変化）の欠如**: 「半年前と比べて DL 数が伸びてるか」がわかると、新興ライブラリの評価精度が上がる
- **PyPI Stats API のレートリミット**: 複数パッケージを短時間に叩くと 429 が出やすい

### 結論

人間が手で使うというより、AI エージェントのツールチェーンに組み込んだときに真価を発揮するツール。MCP サーバー化するとさらに自然に使えるようになる。
