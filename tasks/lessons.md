# Lessons

## 用語

### LRU eviction (Least Recently Used eviction)

キャッシュが容量上限に達したとき、**最も長い間使われていないエントリから順に削除する**方式。

- LRU = Least Recently Used (最も最近使われていないもの)
- eviction = 追い出し、削除
- 例: 上限 100 エントリのキャッシュで 101 個目を書き込むとき、最後にアクセスされた時刻が最も古いエントリを削除して空きを作る
- 対になる概念: LFU (Least Frequently Used) = 使用頻度が最も低いものを削除
- repiq では TTL ベースの期限切れで十分なため Out of Scope としている
