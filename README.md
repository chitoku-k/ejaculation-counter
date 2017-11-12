ぴゅっぴゅカウンター
====================

[![][dependencies-badge]][dependencies-link]

[ぴゅっぴゅカウンター](https://xn--y2wx43a.chitoku.jp) はぴゅっぴゅ回数をカウントして毎日真夜中にツイートします。

## 機能

- Twitter のプロフィールの名前部分に自動更新で前日と当日のぴゅっぴゅ回数を表示
- 日々のぴゅっぴゅ回数をデータベースに記録
- 毎日真夜中にカウンターとデータベースを更新
- 「ぴゅっ♡」を含むツイートでぴゅっぴゅカウンターを更新

## おまけ

- だいたい以下のようなツイートで診断メーカーの結果をリプライ
  - 「ぴゅっぴゅしていい？」
  - 「おふとん入っていい？」
  - 「ちんぽに勝ちたい」
  - 「ちんぽチャレンジ」
  - 「おちんぽ挿入チャレンジ」
  - 「おちんちん握って」
- ニジエ と ホルネ の連携ツイートでぴゅっぴゅカウンターを更新
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

# Twitter ユーザー ID（数値）
TWITTER_ID=

# Twitter 開発者用キー
TWITTER_CONSUMER_KEY=
TWITTER_CONSUMER_SECRET=

# Twitter ユーザー トークン
TWITTER_ACCESS_TOKEN=
TWITTER_ACCESS_TOKEN_SECRET=

# データベース 接続先
MYSQL_HOST=
MYSQL_DATABASE=

# データベース 接続情報
MYSQL_USER=
MYSQL_PASSWORD=
```

Twitter のプロフィールの更新を利用する場合は名前を変更します。  
名前の末尾に次の文字列を置いてプロフィールを設定すると次のぴゅっぴゅから反映されます。

```
（昨日: 0 / 今日: 0）
```

## 動作環境

- Node.js 7 以上
- MySQL 互換の RDBMS

## 実行

環境変数を `.env` として保存して次のコマンドを実行します。

```bash
$ npm start
```

[dependencies-link]:    https://gemnasium.com/github.com/chitoku-k/ejaculation-counter
[dependencies-badge]:   https://img.shields.io/gemnasium/chitoku-k/ejaculation-counter.svg?style=flat-square
