
# 고급 가이드: 지식 베이스 기반 어시스턴트

이 가이드는 Quickstart와 동일한 HTTP 인터페이스를 유지하면서, 자체 지식 소스에 기반해 질문에 답할 수 있는 어시스턴트를 설계하는 방법을 설명합니다.

목표는 다음과 같습니다.

- 클라이언트 통합을 소수의 HTTP 엔드포인트로 유지할 것  
- 모델 호출 전후에 벡터 검색 등 “검색 단계”를 삽입할 것  
- 어디까지가 “지식 구성/검색”이고 어디서부터가 AgentOS 런타임의 책임인지 명확히 할 것  

## 1. 시나리오

예를 들어, 제품 문서나 사내 가이드라인에 대해 질문에 답변하는 어시스턴트를 만들고 싶다고 가정해 봅시다. 고수준 플로우는 다음과 같습니다.

1. 오프라인에서 문서를 임베딩으로 변환하고 벡터 스토어에 저장합니다(상세 구현은 본 가이드 범위 밖).  
2. 질의 시점에, 사용자 질문에 대해 가장 관련성이 높은 패시지를 벡터 스토어에서 검색합니다.  
3. 검색된 컨텍스트를 메시지 콘텐츠의 일부로 Agent에 전달합니다.  

런타임은 계속해서 Agent, Session, Message 관리 역할을 담당합니다.

## 2. Agent와 세션

Agent와 세션 생성은 Quickstart의 패턴을 그대로 재사용할 수 있습니다.

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

## 3. 검색된 컨텍스트 전달

애플리케이션이 지식 스토어에서 관련 패시지를 검색한 후, 이를 메시지 콘텐츠에 직접 포함할 수 있습니다.

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "다음 컨텍스트를 활용해 질문에 답변해 주세요.\\n\\n[CONTEXT]\\n...검색된 패시지...\\n\\n질문: 우리의 환불 정책은 어떻게 되어 있나요?"
  }'
```

또는 세션 생성 시점이나 애플리케이션 자체 상태 관리에서 `metadata` 필드에 검색 메타데이터를 실어 보낼 수도 있습니다. 런타임 API는 특정 검색 패턴을 강제하지 않습니다.

## 4. 구성과 프로바이더 선택

지식 베이스 기반 어시스턴트를 구축할 때는:

- “프로바이더 매트릭스”를 참고하여 긴 컨텍스트를 잘 처리할 수 있는 프로바이더와 모델을 선택합니다.  
- `.env`에 필요한 환경 변수(예: `OPENAI_API_KEY`, `GEMINI_API_KEY`)를 설정하고, “Configuration & Security Practices” 페이지에서 그 의미를 설명합니다.  
- 지식 인덱싱과 검색 인프라(벡터 스토어, 데이터베이스, 스토리지 등)는 런타임 외부의 별도 컴포넌트로 취급하고, 검색 결과만 메시지 콘텐츠에 주입합니다.  

## 5. 테스트와 개선

이 패턴을 검증할 때는:

- 소규모의 선별된 문서와 테스트 질문 세트로 시작합니다.  
- 검색 컨텍스트를 제공했을 때 어시스턴트가 질문에 정확히 답변하는지 확인합니다.  
- 불완전하거나 잘못된 답변을 기록하고, 검색 전략과 프롬프트 설계를 개선하는 데 활용합니다.  
