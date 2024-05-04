# 郵便番号 CSV ユーティリティ

## 概要
郵便番号CSVファイルへのアクセスを容易にするためのユーティリティ

### download
郵便番号CSVファイルをダウンロードします。<br/>
zipファイルをダウンロードして展開したCSVファイルを出力します。

| パラメータ      | 短縮 | デフォルト | 説明                                                                                 | 例                            |
|:-----------|:---|:------|:-----------------------------------------------------------------------------------|:-----------------------------|
| --output   | -o | data  | 出力ディレクトリパス<br/>CSVファイルの保存先  | yubindango dl -o=data        |
| --jigyosyo | -j | https://www.post.japanpost.jp/zipcode/dl/jigyosyo/zip/jigyosyo.zip  | 事業所CSVのダウンロードURL                                                                   | yubinbango dl -j=https://... |
| --ken-all  | -k | https://www.post.japanpost.jp/zipcode/dl/utf/zip/utf_ken_all.zip | ken-allのダウンロードURL                                                                  | yubinbango dl -k=https://... |

```sh
$ yubinbango dl -o 出力ディレクトリパス -j 事業所データフラグ
```

### csv2json
郵便番号CSVファイルを読み込み、JSON形式に変換します。

| パラメータ    | 短縮 | デフォルト | 説明                           | 例                                    |
|:---------|:---|:---|:-----------------------------|:-------------------------------------|
| --path   | -p | ./data/*.csv,./data/*.CSV |  CSVファイルパス<br/>変換対象のCSVファイルパス | yubinbango c2j -p=./data/**/*.csv    |
| --output | -o | ./data/output/json | 出力ディレクトリパス<br/>JSONファイルの保存先  | yubinbango c2j -o=./data/output/json |
| --renew  | -r | false | 再作成フラグ<br/>JSONファイルを再作成する    | yubinbango c2j -r                    |

```sh
$ yubinbango c2j -p CSVファイルパス -o 出力ディレクトリパス -r 再作成フラグ
```

### json2jsonp
JSON形式のファイルを読み込み、JSONP形式に変換します。

| パラメータ | 短縮 | デフォルト                             | 説明                             | 例                                    |
|:---|:---|:----------------------------------|:-------------------------------|:-------------------------------------|
| --path | -p | ./data/output/json                | JSONファイルパス<br/>変換対象のJSONファイルパス | yubinbango j2j -p=./data/output/json |
| --output | -o | ./data/output/js |  出力ディレクトリパス<br/>JSONPファイルの保存先  | yubinbango j2j -o=./data/output/js   |


```sh
$ yubinbango j2j -p JSONファイルのディレクトリパス -o 出力ディレクトリパス
```

### server
指定したJSON、JSONP形式のファイルを読み込み、レスポンスを返すAPIサーバーを起動します。

| パラメータ | 短縮  | デフォルト               | 説明                                                            | 例                                      |
|:---|:----|:--------------------|:--------------------------------------------------------------|:---------------------------------------|
| --data | -d  | file://data/output/ | データディレクトリパス<br/>JSON、JSONPファイルのディレクトリパス                       | yubinbango server -d=./data/output     |
| --health | -h  | false               | ヘルスチェック有効フラグ<br/>ヘルスチェック用APIを有効化する                            | yubinbango server -h                   |
| --basic | -b  |                 |  ベーシック認証ユーザーパスワード<br/>`username:password` の形式でユーザー/パスワードを設定する | yubinbango server -b=username:password |
| --basic-auth | -B | false     | ベーシック認証有効化フラグ<br/>ベーシック認証を有効化する                               | yubinbango server -B                   |

#### 環境変数

| 環境変数                | デフォルト | 説明     |       
|:--------------------|:------|:--------------|
| PORT                | 8080  | ポート番号         |
| DATA_DIR_PATH       |       | データディレクトリパス   |
| HEALTH_CHECK        | false | ヘルスチェック有効フラグ  |
| BASIC_AUTH_USER     | user  | ベーシック認証ユーザー   |
| BASIC_AUTH_PASSWORD | pass  | ベーシック認証パスワード  |
| BASIC_AUTH_ENABLE   | false | ベーシック認証有効化フラグ |

#### api
./api ディレクトリにAPIの仕様書を格納しています。

```shell
$ yubinbango server -d データディレクトリパス -h ヘルスチェック有効フラグ　-b ベーシック認証ユーザーパスワード -B ベーシック認証有効化フラグ
```
