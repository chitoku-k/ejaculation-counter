ぴゅっぴゅカウンター
====================

[ぴゅっぴゅカウンター](https://xn--y2wx43a.chitoku.jp) はぴゅっぴゅ回数をカウントして毎日真夜中にトゥートします。

## 機能

- 日々のぴゅっぴゅ回数をデータベースに記録
- 毎日真夜中にカウンターとデータベースを更新
- 「ぴゅっ♡」を含むトゥートでぴゅっぴゅカウンターを更新

## おまけ

- だいたい以下のようなトゥートで診断メーカーの結果をリプライ
  - 「ぴゅっぴゅしていい？」
  - 「おふとん入っていい？」
  - 「ちんぽに勝ちたい」
  - 「ちんぽチャレンジ」
  - 「おちんぽ挿入チャレンジ」
  - 「おちんちん握って」
  - 「○○のAV」
- 「through ガチャ」

## 設定方法

データベースの作成とテーブルの設定を行います。

```sql
CREATE TABLE `counts` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `user` int(11) DEFAULT NULL,
    `date` date DEFAULT NULL,
    `count` int(11) DEFAULT NULL,
    PRIMARY KEY (`id`)
);
CREATE TABLE `users` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `screen_name` varchar(20) DEFAULT NULL,
    PRIMARY KEY (`id`)
);
```

環境変数に値の設定を行います。

```bash
# ぴゅっぴゅユーザー ID（数値）
SHIKO_USER=

# Mastodon ユーザー ID（数値、スペース区切り）
MASTODON_ID=

# Mastodon アプリケーション名
MASTODON_APP=

# Mastodon ユーザー トークン
MASTODON_ACCESS_TOKEN=

# データベース 接続先
MYSQL_HOST=
MYSQL_DATABASE=

# データベース 接続情報
MYSQL_USER=
MYSQL_PASSWORD=
```

## 動作環境

- Node.js 7 以上
- MySQL 互換の RDBMS

## 実行

環境変数を `.env` として保存して次のコマンドを実行します。

```bash
$ npm start
```
