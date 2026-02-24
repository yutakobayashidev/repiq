# Adding a New Provider

repiq に新しいプロバイダーを追加する手順。

## 変更が必要なファイル

| ファイル | 変更内容 |
|----------|---------|
| `internal/provider/provider.go` | Metrics 構造体の追加、Result にフィールド追加 |
| `internal/provider/provider_test.go` | 新 Metrics の Result テスト追加 |
| `internal/provider/<scheme>/<scheme>.go` | プロバイダー実装 (新規) |
| `internal/provider/<scheme>/<scheme>_test.go` | ユニットテスト (新規) |
| `internal/cli/cli.go` | プロバイダーの登録 |
| `internal/format/format.go` | Markdown テーブルの追加 |
| `internal/format/format_test.go` | テストデータの追加 |

変更不要: `target.go`, `registry.go`, `cmd/repiq/main.go`

## Step 1: Metrics 構造体を定義する

`internal/provider/provider.go` に Metrics 構造体を追加し、`Result` に `omitempty` 付きフィールドを追加する。

```go
// provider.go

type CratesMetrics struct {
    RecentDownloads int    `json:"recent_downloads"`
    LatestVersion   string `json:"latest_version"`
    // ...
}

type Result struct {
    Target string         `json:"target"`
    GitHub *GitHubMetrics `json:"github,omitempty"`
    NPM    *NPMMetrics    `json:"npm,omitempty"`
    Crates *CratesMetrics `json:"crates,omitempty"`  // 追加
    Error  string         `json:"error,omitempty"`
}
```

`provider_test.go` にテストケースを追加:

```go
func TestResultCratesSuccess(t *testing.T) {
    r := provider.Result{
        Target: "crates:serde",
        Crates: &provider.CratesMetrics{RecentDownloads: 1000000},
    }
    if r.Crates == nil {
        t.Fatal("expected Crates to be non-nil")
    }
}
```

## Step 2: プロバイダーを実装する

`internal/provider/<scheme>/` ディレクトリを作成し、以下のインターフェースを実装する:

```go
type Provider interface {
    Scheme() string
    Fetch(ctx context.Context, identifier string) (Result, error)
}
```

### 構造

```go
// internal/provider/crates/crates.go
package crates

import (
    "github.com/yutakobayashidev/repiq/internal/provider"
)

type Provider struct {
    baseURL string
    client  *http.Client
}

// New はテスト用に baseURL を差し替えられるようにする。
// 空文字を渡すとデフォルト URL を使う。
func New(baseURL string) *Provider {
    if baseURL == "" {
        baseURL = "https://crates.io/api/v1"
    }
    return &Provider{
        baseURL: strings.TrimRight(baseURL, "/"),
        client:  &http.Client{Timeout: 30 * time.Second},
    }
}

func (p *Provider) Scheme() string { return "crates" }

func (p *Provider) Fetch(ctx context.Context, identifier string) (provider.Result, error) {
    // 実装
}
```

### 実装ルール

| ルール | 理由 |
|--------|------|
| `New()` で baseURL を受け取る | httptest でモックサーバーに差し替え可能にする |
| `http.Client` に `Timeout: 30s` を設定する | context タイムアウトが未設定でもハングしない |
| identifier をバリデーションする | SSRF / URL injection 防止 |
| エラーは `Result.Error` に入れ、Go error は返さない | CLI の一括取得で部分失敗を許容するため |
| 全失敗時は `Result.Error` のみ設定し Metrics は nil | JSON で `"crates": null` にならず、フィールドごと省略される |
| 部分失敗時は取得できた Metrics を設定しつつ Error も記録 | 利用可能なデータは返す |
| `resp.Body.Close()` は `defer func() { _ = resp.Body.Close() }()` | errcheck lint 対応 |

### 並列取得パターン

複数 API エンドポイントからデータを取得する場合は `sync.WaitGroup` + `sync.Mutex` で並列実行する:

```go
func (p *Provider) Fetch(ctx context.Context, identifier string) (provider.Result, error) {
    metrics := &provider.CratesMetrics{}
    var mu sync.Mutex
    var wg sync.WaitGroup
    var errs []string

    type job struct {
        name string
        fn   func(context.Context) error
    }

    jobs := []job{
        {"info", func(ctx context.Context) error {
            // API 呼び出し → mu.Lock() → metrics に書き込み → mu.Unlock()
        }},
        {"downloads", func(ctx context.Context) error {
            // ...
        }},
    }

    wg.Add(len(jobs))
    for _, j := range jobs {
        go func(j job) {
            defer wg.Done()
            if err := j.fn(ctx); err != nil {
                mu.Lock()
                errs = append(errs, fmt.Sprintf("%s: %s", j.name, err.Error()))
                mu.Unlock()
            }
        }(j)
    }
    wg.Wait()

    result := provider.Result{Target: "crates:" + identifier}

    if len(errs) == len(jobs) {
        // 全失敗
        result.Error = strings.Join(errs, "; ")
        return result, nil
    }
    result.Crates = metrics
    if len(errs) > 0 {
        // 部分失敗
        result.Error = strings.Join(errs, "; ")
    }
    return result, nil
}
```

## Step 3: テストを書く

`httptest.Server` + `http.NewServeMux()` でモックサーバーを立てる。

```go
// internal/provider/crates/crates_test.go
package crates

func setupMockServer(t *testing.T) *httptest.Server {
    t.Helper()
    mux := http.NewServeMux()

    mux.HandleFunc("GET /crates/serde", func(w http.ResponseWriter, _ *http.Request) {
        json.NewEncoder(w).Encode(map[string]any{...})
    })

    srv := httptest.NewServer(mux)
    t.Cleanup(srv.Close)
    return srv
}

func TestFetchSuccess(t *testing.T) {
    srv := setupMockServer(t)
    p := New(srv.URL)
    result, err := p.Fetch(context.Background(), "serde")
    // assertions...
}
```

### 必須テストケース

| テスト | 内容 |
|--------|------|
| `TestScheme` | `Scheme()` が正しいスキーム名を返す |
| `TestFetchSuccess` | 正常系で全メトリクスが取得される |
| `TestFetchNotFound` | 存在しないパッケージで `Result.Error` が設定される |
| `TestFetchEmptyIdentifier` | 空文字でバリデーションエラー |
| `TestFetchInvalidIdentifier` | 不正な文字列でバリデーションエラー |
| `TestFetchPartialFailure` | 一部 API 失敗時に Metrics + Error が両方設定される |

テスト内の日時は `time.Now().Add(-N * 24 * time.Hour)` で動的に生成し、flaky test を防ぐ。

## Step 4: CLI に登録する

`internal/cli/cli.go` に 2 箇所追加:

```go
import (
    // ...
    cratesprovider "github.com/yutakobayashidev/repiq/internal/provider/crates"
)

// Run() 内の provider 登録部分:
registry.Register(cratesprovider.New(""))
```

Usage の Examples にも追加:

```
  repiq crates:serde
```

## Step 5: Markdown フォーマッターを更新する

`internal/format/format.go` に 3 箇所追加:

### 1. 結果の分類

```go
var cratesResults []provider.Result

case r.Crates != nil:
    cratesResults = append(cratesResults, r)
```

### 2. テーブルのレンダリング

```go
if len(cratesResults) > 0 {
    if needSep {
        fmt.Fprintln(w)
    }
    fmt.Fprintln(w, "| target | recent_downloads | latest_version | ... | error |")
    fmt.Fprintln(w, "|---|---|---|---|---|")
    for _, r := range cratesResults {
        c := r.Crates
        fmt.Fprintf(w, "| %s | %s | %s | ... | %s |\n",
            escapeMarkdown(r.Target),
            strconv.Itoa(c.RecentDownloads),
            escapeMarkdown(c.LatestVersion),
            escapeMarkdown(r.Error),
        )
    }
    needSep = true
}
```

文字列フィールドには必ず `escapeMarkdown()` を適用する。

### 3. テストデータの追加

`format_test.go` の `sampleResults()` に新プロバイダーの成功ケースを追加し、既存テストの item 数を更新する。

## Step 6: 検証

```bash
go test ./...
golangci-lint run
go build ./cmd/repiq && ./repiq crates:serde
```

## Checklist

- [ ] `provider.go` に Metrics 構造体と Result フィールドを追加
- [ ] `provider_test.go` に Result テストを追加
- [ ] `internal/provider/<scheme>/` に実装 + テスト
- [ ] identifier バリデーションを実装
- [ ] `http.Client` に Timeout を設定
- [ ] `resp.Body.Close()` を `defer func() { _ = ... }()` で呼ぶ
- [ ] httptest で success, 404, empty, invalid, partial failure をテスト
- [ ] `cli.go` に import + `Register()` を追加
- [ ] `format.go` に Markdown テーブルを追加 (escapeMarkdown 適用)
- [ ] `format_test.go` にサンプルデータを追加
- [ ] `go test ./...` pass
- [ ] `golangci-lint run` 0 issues
- [ ] CLAUDE.md の scope リストにスキーム名を追加
