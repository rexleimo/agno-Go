
# 高度なガイド：メモリ拡張チャット

このガイドでは、対話の複数ターンやセッションをまたいでメモリを活用するチャット体験を構築する方法を説明します。特定のストレージ製品に依存することなく、既存の HTTP API とメタデータフィールドをどのように使えばよいかに焦点を当てます。

## 1. メモリの種類

メモリは大きく 3 つの層として考えることができます：

- **会話履歴**：現在のセッション内の最近のメッセージ  
- **ユーザープロファイル**：ユーザーに関する長期的な情報（嗜好、属性など）  
- **ナレッジレコード**：ドメイン固有の事実（過去のやり取りや重要イベントなど）  

Agno-Go は Session + Message によって第一層（会話履歴）をネイティブにサポートし、残りの層は設定と自前のサービスによって接続します。

## 2. メモリ対応エージェントの作成

Go からは、通常 HTTP 経由で AgentOS ランタイムにアクセスします。Quickstart と
同じ流れを Go コードで表現すると、次のようになります。

```go
package main

import (
  "bytes"
  "encoding/json"
  "log"
  "net/http"
  "time"
)

type Agent struct {
  Name        string                 `json:"name"`
  Description string                 `json:"description"`
  Model       map[string]any         `json:"model"`
  Tools       []map[string]any       `json:"tools"`
  Config      map[string]any         `json:"config"`
}

func main() {
  client := &http.Client{Timeout: 10 * time.Second}

  agent := Agent{
    Name:        "memory-chat-agent",
    Description: "A chat agent that uses session history and external memory.",
    Model: map[string]any{
      "provider": "openai",
      "modelId":  "gpt-4o-mini",
      "stream":   true,
    },
    Tools:  nil,
    Config: map[string]any{},
  }

  body, err := json.Marshal(agent)
  if err != nil {
    log.Fatalf("marshal agent: %v", err)
  }

  resp, err := client.Post("http://localhost:8080/agents", "application/json", bytes.NewReader(body))
  if err != nil {
    log.Fatalf("create agent: %v", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusCreated {
    log.Fatalf("unexpected status: %s", resp.Status)
  }

  // 実際のアプリケーションでは、ここでレスポンスを decode して agentId を取得し、
  // その後のセクションのように Session 作成やメッセージ送信を行います。
}
```

ターミナルや API クライアントで生の HTTP を試したい場合は、同等の `curl`
コマンドも利用できます。

```bash
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "memory-chat-agent",
    "description": "A chat agent that uses session history and external memory.",
    "model": {
      "provider": "openai",
      "modelId": "gpt-4o-mini",
      "stream": true
    },
    "tools": [],
    "config": {}
  }'
```

メモリ対応エージェントかどうかを分けるポイントは、後続で説明するように
セッション構造とメタデータの渡し方にあります。

## 3. セッションとメタデータの活用

セッション作成時にユーザー固有の識別子やメタデータを付与できます。

```bash
curl -X POST http://localhost:8080/agents/<agent-id>/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-1234",
    "metadata": {
      "source": "advanced-memory-chat",
      "segment": "beta-testers"
    }
  }'
```

アプリケーション側では、`userId` や `metadata` を用いて自前ストレージからユーザープロファイルを参照・更新し、その情報を後続メッセージに反映させることができます。

## 4. プロンプトへのメモリ統合

メッセージ送信時に、既知の事実やコンテキストを `content` に組み込むことができます。

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "以前、あなたは私のために読書プランを提案してくれました。その提案内容と、短時間の読書を好むという私の好みを踏まえて、今週のプランを提案してください。"
  }'
```

バックエンド側では：

- メモリストアから過去のやり取りやメモを読み込む  
- 要約や重要な事実をプロンプトに組み込む  
- その上で標準的なメッセージエンドポイントを通してランタイムに渡す  

## 5. 設定とストレージ

メモリに依存する度合いが高いユースケースでは：

- 「Configuration & Security Practices」ドキュメントを参考に、どのメモリバックエンドを有効化するか（インメモリ vs ローカル永続化など）を決める。  
- 追加のインフラ（データベース、キャッシュ、キューなど）は内部運用ドキュメントに記録し、AgentOS ランタイムは HTTP 動作と契約に専念させる。  
- `.env` と `config/default.yaml` の設定が公式ドキュメントのガイダンス（特に保持期間やデータ所在）と整合していることを確認する。  

## 6. テストと進化

メモリ拡張チャットを検証する際には：

- 「短期」および「長期」メモリの両方の挙動をカバーするテストプランを設計する。  
- メモリ使用量が増えても `/health` と Quickstart フローでランタイムが健康であることを確認する。  
- レイテンシやリソース使用量を監視し、要約頻度やリプレイ長といったメモリ戦略を実測結果に基づき調整する。  
