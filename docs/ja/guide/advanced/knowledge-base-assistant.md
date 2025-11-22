
# 高度なガイド：ナレッジベースアシスタント

このガイドでは、Quickstart と同じ HTTP インターフェースを保ちながら、自分のナレッジソースに基づいて質問に回答できるアシスタントを設計する方法を説明します。

目的は次のとおりです。

- クライアント側の統合を最小限の HTTP エンドポイントに保つ  
- モデル呼び出しの前後に、ベクターストアなどの「検索ステップ」を挿入する  
- どこまでが「ナレッジ設定/検索」で、どこからが AgentOS ランタイムの責務かを明確にする  

## 1. シナリオ

たとえば、プロダクトドキュメントや社内ガイドラインに関する質問に回答するアシスタントを作りたいとします。高レベルのフローは以下の通りです。

1. オフラインでドキュメントをベクトル化し、ベクターストアに格納する（ここでは詳細は扱いません）。  
2. クエリ時に、ユーザーの質問に対して最も関連性の高いパッセージをベクターストアから取得する。  
3. 取得したコンテキストをメッセージコンテンツの一部として Agent に渡す。  

ランタイムは引き続き、Agent・Session・Message の管理を担当します。

## 2. Agent とセッション

Agent とセッションの作成は Quickstart のパターンを再利用できます。

```bash
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "kb-assistant",
    "description": "Answers questions using knowledge base context.",
    "model": "openai:gpt-4o-mini",
    "tools": [],
    "config": {}
  }'
```

```bash
curl -X POST http://localhost:8080/agents/<agent-id>/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "kb-user",
    "metadata": {
      "source": "advanced-knowledge-base-assistant"
    }
  }'
```

## 3. 取得コンテキストの渡し方

アプリケーションがナレッジストアから関連パッセージを取得したら、それをメッセージコンテンツに含めることができます。

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "次のコンテキストを使って質問に答えてください。\\n\\n[CONTEXT]\\n...取得したパッセージ...\\n\\n質問：当社の返金ポリシーはどうなっていますか？"
  }'
```

別案として、セッション作成時やアプリケーション側の状態管理の中で `metadata` フィールドに検索メタデータを格納することもできます。ランタイム API 自体は特定の検索パターンを強制しません。

## 4. 設定とプロバイダ選定

ナレッジベースアシスタントを構築する際には：

- 「プロバイダマトリクス」を参考に、長いコンテキストを扱いやすいプロバイダとモデルを選択する。  
- `.env` に必要な環境変数（`OPENAI_API_KEY` や `GEMINI_API_KEY` など）を設定し、「Configuration & Security Practices」ページでその意味を説明する。  
- ナレッジのインデックス作成と検索インフラ（ベクターストア、データベース、ストレージなど）はランタイムの外側に置き、取得結果のみをメッセージコンテンツに注入する。  

## 5. テストと改善

このパターンを検証する際には：

- 小規模で厳選されたドキュメントとテスト質問のセットから始める。  
- 検索コンテキストを与えたときにアシスタントが適切に回答できるか確認する。  
- 不完全または誤った回答を記録し、検索戦略やプロンプト設計を改善するための材料とする。  
