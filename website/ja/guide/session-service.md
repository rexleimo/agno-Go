# Go セッションサービス

Go セッションサービスは Python AgentOS の `/sessions` REST API を完全に再現し、
単体の Go バイナリとして動作します。AgentOS と並行してデプロイすることで、
PostgreSQL バックエンドの高速な会話 CRUD、運用向け HTTP ミドルウェア、そして
ローカルから Kubernetes まで網羅するデプロイ資産を利用できます。

## 機能ハイライト

- **エンドポイントの完全互換**: `/sessions` の一覧・作成、`/sessions/{id}` の詳細、
  リネーム・削除、`/sessions/{id}/runs` の履歴取得、`/healthz` のヘルスチェック。
- **Postgres ストレージ**: 型付き DTO とトランザクションセーフな処理で既存 AgentOS
  と同じ JSON 契約を保証。
- **マルチデータベース対応**: `AGNO_SESSION_DSN_MAP` で複数 DSN を登録し、
  `db_id` クエリパラメータでルーティングを切り替え。
- **運用向けミドルウェア**: Chi ルーターがリクエスト ID、構造化ログ、Real IP、
  パニックリカバリ、60 秒タイムアウトを提供。
- **充実したデプロイ資産**: 専用 Dockerfile、Postgres 付き Docker Compose、
  Helm Chart、curl によるスモークテストスクリプトを同梱。

## ローカルでのクイックスタート

```bash
export AGNO_PG_DSN="postgres://user:pass@localhost:5432/agentos?sslmode=disable"
export AGNO_SERVICE_PORT=8080
go run ./cmd/agentos-session
```

ログに `Go session service listening on :8080` が表示され、
`http://localhost:8080/healthz` で JSON のヘルスレスポンスを確認できます。

### 契約テストとの整合

- `make contract-test` — Go 実装と Python 実装のレスポンスを突き合わせる契約テストを実行。
- `./scripts/test-session-api.sh http://localhost:8080` — curl + jq で一覧/作成/リネーム/削除の基本フローを実施。

## 設定リファレンス

| 環境変数               | 説明                                                                                     | デフォルト |
|------------------------|------------------------------------------------------------------------------------------|------------|
| `AGNO_PG_DSN`          | メインの Postgres DSN。`AGNO_SESSION_DSN_MAP` が無い場合は必須。                        | なし       |
| `DATABASE_URL`         | `AGNO_PG_DSN` 未設定時のフォールバック DSN（Heroku 形式）。                              | なし       |
| `AGNO_SESSION_DSN_MAP` | `{"dbID":"dsn"}` 形式の JSON で複数データベースを登録し、`db_id` で切り替え可能。 | なし       |
| `AGNO_DEFAULT_DB_ID`   | `AGNO_SESSION_DSN_MAP` 利用時のデフォルト DB ID。                                        | 先頭のキー |
| `AGNO_SERVICE_PORT`    | HTTP のリッスンポート。                                                                   | `8080`     |

`AGNO_SESSION_DSN_MAP` を設定した場合は、
`/sessions?type=agent&db_id=analytics` のようにリクエストでターゲット DB を指定します。

## API サマリー

| Endpoint               | Method | 説明                                                                                   |
|------------------------|--------|----------------------------------------------------------------------------------------|
| `/healthz`             | GET    | `{"status":"ok"}` を返すヘルスチェック。                                               |
| `/sessions`            | GET    | ページング & フィルタ (`type`, `component_id`, `user_id`, `session_name`, `sort_by`, `db_id`) に対応。 |
| `/sessions`            | POST   | 状態・メタデータ・事前投入済み runs/summary を含むセッションを作成。                    |
| `/sessions/{id}`       | GET    | タイプと ID を元に詳細取得。`db_id` でデータベース指定も可能。                         |
| `/sessions/{id}`       | DELETE | セッションと履歴を削除。                                                                |
| `/sessions/{id}/rename`| POST   | `session_name` を更新。                                                                |
| `/sessions/{id}/runs`  | GET    | 保存された run 履歴を取得。Python AgentOS とレスポンス形式が同一。                      |

書き込み系のエンドポイントは Python フィクスチャに準拠した JSON を受け付け、契約テストで検証されます。

## Docker Compose で実行

`docker-compose.session.yml` を利用して Postgres と Go サービスを同時に立ち上げます。

```bash
docker compose -f docker-compose.session.yml up --build
```

起動後は `http://localhost:8080` にアクセスし、スクリプトで動作を確認してください。

## Helm デプロイ

`deploy/helm/agno-session/` の Chart を使うと Kubernetes へ展開できます。以下は基本コマンドです。

```bash
helm upgrade --install agno-session ./deploy/helm/agno-session \
  --set image.repository=ghcr.io/<org>/agno-session \
  --set image.tag=v1.2.9 \
  --set config.dsn="postgres://user:pass@postgres:5432/agentos?sslmode=disable"
```

`values.yaml` にはプローブ、レプリカ数、Service 形態などの詳細設定が含まれています。

## 本番運用チェックリスト

- ステージングで `make contract-test` を実行しレスポンス互換性を確認。
- 本番トラフィックの一部をミラーし、Python サービスと JSON を比較検証。
- Postgres の接続数やレイテンシを監視し、必要に応じてコネクションプールを導入。
- `/healthz` と HTTP ログを用いたダッシュボード・アラートを構成。
- 本番安定までは Python 実装を待機系として残しておくことを推奨。
