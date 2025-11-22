# 퀵스타트: 10분 안에 Agno-Go 체험하기

이 가이드는 약 10분 안에 Agno-Go를 사용해 첫 번째 요청부터 응답까지 전체 흐름을 경험하도록 도와줍니다.

1. AgentOS 런타임를 실행합니다.  
2. 최소한의 에이전트를 생성합니다.  
3. 세션을 생성합니다.  
4. 메시지를 보내고 응답을 확인합니다.  

> 모든 경로는 프로젝트 루트(예: `<your-project-root>/go/cmd/agno`, `<your-project-root>/config/default.yaml`)를 기준으로 합니다. 자신의 환경에 맞게 플레이스홀더를 바꿔 사용하세요.

프로젝트 루트에서 서비스 실행:

```bash
cd <your-project-root>
go run ./go/cmd/agno --config ./config/default.yaml
```

헬스 체크:

```bash
curl http://localhost:8080/health
```

최소 에이전트 생성:

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

해당 에이전트에 대한 세션 생성:

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

세션에 메시지 전송:

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "Agno-Go에 대해 간단히 소개해 주세요."
  }'
```

`messageId`, `content`, `usage`, `state` 필드를 포함하는 JSON 응답이 반환되는지 확인하세요.

## 다음 단계

- [구성 및 보안](./config-and-security) 문서를 읽어 각 프로바이더의 키, 엔드포인트, 런타임 옵션을 안전하게 설정하는 방법을 확인하세요.  
- [Core Features & API](./core-features-and-api) 와 [프로바이더 매트릭스](./providers/matrix) 를 살펴보며 전체 기능을 체계적으로 이해할 수 있습니다.  
- 기본 흐름에 익숙해졌다면, [고급 가이드](./advanced/multi-provider-routing) 의 사례를 따라 더 복잡한 워크플로를 구성해 보세요.  
