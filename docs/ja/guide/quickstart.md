# クイックスタート：10 分で体験する Agno-Go

このガイドでは、Agno-Go を使った最小限のエンドツーエンドフローを 10 分程度で体験します。

1. AgentOS ランタイムを起動する  
2. 最小限のエージェントを作成する  
3. セッションを作成する  
4. メッセージを送信してレスポンスを確認する  

> すべてのパスはプロジェクトルート（例：`<your-project-root>/go/cmd/agno`、`<your-project-root>/config/default.yaml`）からの相対パスとして記載されています。ご自身の環境に合わせて置き換えてください。

プロジェクトルートでサービスを起動します：

```bash
cd <your-project-root>
go run ./go/cmd/agno --config ./config/default.yaml
```

サービスのヘルスチェック：

```bash
curl http://localhost:8080/health
```

最小限のエージェントを作成します：

```bash
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "quickstart-agent",
    "description": "A minimal agent created from the docs quickstart.",
    "model": "openai:gpt-4o-mini",
    "tools": [],
    "config": {}
  }'
```

エージェントのセッションを作成します：

```bash
curl -X POST http://localhost:8080/agents/<agent-id>/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "quickstart-user",
    "metadata": {
      "source": "docs-quickstart"
    }
  }'
```

セッション内でメッセージを送信します：

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "Agno-Go を簡単に紹介してください。"
  }'
```

`messageId`、`content`、`usage`、`state` フィールドを含む JSON レスポンスが返ってくることを確認してください。

## 次のステップ

- [設定とセキュリティ](./config-and-security) を読み、プロバイダのキーやエンドポイント、ランタイム設定を安全に扱う方法を確認してください。  
- [Core Features & API](./core-features-and-api) や [プロバイダマトリクス](./providers/matrix) に進み、利用可能な機能を体系的に把握します。  
- 基本フローに慣れてきたら、[高度なガイド](./advanced/multi-provider-routing) のケースを試して、より複雑なワークフローを構築できます。  
