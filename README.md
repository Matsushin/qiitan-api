# qiitan-api
Qiitaのコピーアプリ「[Qiitan](https://github.com/Matsushin/qiitan)」のAPI。

## 概要
記事情報関連のAPIが利用できる。
定期的にキャッシュを保管して、API呼び出し時にキャッシュを返すように。キャッシュ保管のために[Aerospike](https://www.aerospike.jp/)を使用。

# 環境構築

まず、以下の前提条件が満たされていることを確認すること。

* プラットフォームに合わせたDockerがインストール済み
* [docker-compose](https://docs.docker.com/compose/) がインストール済み

その後、以下のように実行。

```shell
$ git clone git@github.com:Matsushin/qiitan-api.git
$ cd qiitan-api

$ docker-compose build
```

## データベース
本リポジトリでは、「[Qiitan](https://github.com/Matsushin/qiitan)」リポジトリで作成されるDBを利用している。
APIで結果を返すには事前にデータベースの作成とデータ投入が必要になります。

# アプリケーション起動

以下のように実行。

```shell
$ docker-compose up
```

その後、 `docker ps` コマンドを実行し、api/mysql/nginx/aerospikeコンテナが起動していることを確認する。

APIへのアクセスは以下のように行う。

```shell
$ curl localhost:18080
```

## 各API

```shell
$ curl localhost:18080/pvt/health # ヘルスチェック
$ curl localhost:18080/v1/articles # 記事一覧
$ curl localhost:18080/v1/ranking/like # いいね記事ランキング
$ curl localhost:18080/v1/ranking/stock # ストック記事ランキング
```

## APIのライブリロードについて

Go APIは[fresh](https://github.com/pilu/fresh)を使ってライブリロードするようになっている。  
また、ホストの `qiitan-api` ディレクトリはコンテナにマウントされているため、  
ホスト側でGoのコードを編集すると即APIの動作に反映される。

## MySQLコンテナへのログイン

MySQLコンテナにログインしたい場合は以下のように実行する。

```shell
$ docker exec -it qiitan-api_mysql_1 bash
# mysql -uroot qiitan # パスワードなし
```

# その他

* ECSへのデプロイ(`deploy.sh`)は、[go-ecs-ecr](https://github.com/circleci/go-ecs-ecr) を使用してCircleCIから実行する。
