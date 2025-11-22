
# 高度なガイド：マルチプロバイダルーティング

このガイドでは、単一の HTTP インターフェースを維持しながら、複数のモデルプロバイダ間でリクエストをルーティングおよびフォールバックする方法を紹介します。

目的は次のとおりです。

- クライアント側には 1 つの AgentOS ランタイムと HTTP サーフェスだけを見せる  
- サーバー側でタスク種別やモデル名に基づいてプロバイダを切り替える  
- メインプロバイダが利用できないときにフォールバックモデルに自動的に切り替え、クライアントコードは変更しない  

## 1. 典型的なユースケース

- 汎用対話にはあるプロバイダ、コストやレイテンシに敏感なワークロードには別のプロバイダを使いたい場合  
- メインプロバイダの障害時に自動的に予備プロバイダへ切り替えたい場合  
- 安定したクライアント統合の上で、新しいモデルを用いた実験や A/B テストを行いたい場合  

## 2. 高レベルな設計

ルーティングロジックはクライアントではなく、Agent の設定とサーバーサイドランタイムに置くことを推奨します：

1. `model` フィールドで「優先モデル/プロバイダ」を表現した Agent を定義する  
2. ランタイムは `model` と構成に基づいて、リクエストを具体的なプロバイダクライアントにルーティングする  
3. クライアントは常に同じ HTTP エンドポイント（`/agents`、`/sessions`、`/messages`）を呼び出す  

モデル名の例：

- `openai:gpt-4o-mini`  
- `gemini:flash-1.5`  
- `groq:llama3-70b`  

具体的なマッピングはサーバー側の設定に委ねます。

## 3. サンプルフロー

1. **ルーティング対応 Agent の作成**

   ```bash
   curl -X POST http://localhost:8080/agents \
     -H "Content-Type: application/json" \
     -d '{
       "name": "routing-agent",
       "description": "An agent that routes across providers based on task type.",
       "model": "openai:gpt-4o-mini",
       "tools": [],
       "config": {
         "fallbackModel": "gemini:flash-1.5"
       }
     }'
   ```

2. **セッションの作成**

   ```bash
   curl -X POST http://localhost:8080/agents/<agent-id>/sessions \
     -H "Content-Type: application/json" \
     -d '{
       "userId": "routing-user",
       "metadata": {
         "source": "advanced-multi-provider-routing"
       }
     }'
   ```

3. **メッセージの送信**

   ```bash
   curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
     -H "Content-Type: application/json" \
     -d '{
       "role": "user",
       "content": "小規模な社内ツールの場合、どのプロバイダ/モデルを推奨しますか？理由も教えてください。"
     }'
   ```

メインプロバイダが利用できない場合、ランタイムは設定された `fallbackModel` にフォールバックできます。クライアント側の呼び出しパターンは変わりません。

## 4. 設定上のポイント

- 各プロバイダのキーやエンドポイント、タイムアウトなどは `.env` と `config/default.yaml` で一元管理する。  
- 「プロバイダマトリクス」ページを参考に、利用するプロバイダと機能の組み合わせを決める。  
- クライアント側でプロバイダ固有のロジックをハードコードするのではなく、Agno-Go ランタイムを「唯一の統合面」として扱う。  

## 5. テストと検証

本パターンを本番で使用する前に：

- Quickstart と同様の呼び出しフローでルーティング Agent の基本動作を確認する。  
- 一時的にあるプロバイダのキーを外し、フォールバックが期待通り動作するかを確認する。  
- プロバイダごとの既知の制約（トークン数、レイテンシなど）を記録し、運用ドキュメントに明示する。  
