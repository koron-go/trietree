# koron-go/trietree

[![PkgGoDev](https://pkg.go.dev/badge/github.com/koron-go/trietree)](https://pkg.go.dev/github.com/koron-go/trietree)
[![Actions/Go](https://github.com/koron-go/trietree/workflows/Go/badge.svg)](https://github.com/koron-go/trietree/actions?query=workflow%3AGo)
[![Go Report Card](https://goreportcard.com/badge/github.com/koron-go/trietree)](https://goreportcard.com/report/github.com/koron-go/trietree)

trietree implement trie-tree and Aho-Corasick algorithm.

## How to install or update

```console
$ go get github.com/koron-go/trietree@latest
```

## Desription

The `trietree` package provides two trie-tree implementations.
One is `DTree` which allows you to dynamically add elements one by one.
Another is `STree`, which is static and cannot add elements to it, but can be serialized and deserialized and is compact.
`STree` can be constructed from `DTree`.
Both trie-trees implement an efficient search based on the Aho–Corasick algorithm.

### Japanese

`trietree` パッケージは2つのトライ木の実装を提供します。
1つは1個ずつ動的に要素を追加できる `DTree` です。
もう1つは静的で要素の追加はできませんが、シリアライズ・デシリアライズが可能でコンパクトな `STree` です。
`STree` は `DTree` から構築できます。
どちらのトライ木もエイホ–コラシック法に基づく効率の良い探索を実装しています。
