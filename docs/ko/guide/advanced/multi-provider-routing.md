
# 고급 가이드: 멀티 프로바이더 라우팅

이 가이드는 하나의 HTTP 인터페이스를 유지하면서 여러 모델 프로바이더 사이에서 요청을 라우팅하고, 필요 시 폴백(fallback)을 적용하는 방법을 설명합니다.

목표는 다음과 같습니다.

- 클라이언트에는 하나의 AgentOS 런타임과 HTTP 표면만 보이도록 할 것  
- 서버 측에서 작업 유형이나 모델 이름에 따라 프로바이더를 전환할 것  
- 기본 프로바이더가 사용 불가일 때 폴백 모델로 자동 전환하면서 클라이언트 코드는 유지할 것  

## 1. 사용 시나리오

- 일반적인 대화에는 한 프로바이더, 비용/지연에 민감한 작업에는 다른 프로바이더를 사용하고 싶을 때  
- 기본 프로바이더 장애 발생 시 자동으로 예비 프로바이더로 전환하고 싶을 때  
- 안정적인 클라이언트 통합 위에서 새로운 모델로 실험이나 A/B 테스트를 수행하고 싶을 때  

## 2. 상위 수준 설계

라우팅 로직은 클라이언트가 아니라 Agent 설정과 서버 사이드 런타임에 두는 것이 좋습니다.

1. `model` 필드로 “우선 모델/프로바이더”를 표현한 에이전트를 정의합니다.  
2. 런타임은 `model`과 구성에 따라 요청을 실제 프로바이더 클라이언트로 라우팅합니다.  
3. 클라이언트는 항상 동일한 HTTP 엔드포인트(` /agents`, `/sessions`, `/messages`)만 호출합니다.  

모델 이름 예시:

- `openai:gpt-4o-mini`  
- `gemini:flash-1.5`  
- `groq:llama3-70b`  

구체적인 매핑은 서버 구성에서 관리합니다.

## 3. 예제 플로우

1. **라우팅 지원 에이전트 생성**

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

2. **세션 생성**

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

3. **메시지 전송**

   ```bash
   curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
     -H "Content-Type: application/json" \
     -d '{
       "role": "user",
       "content": "작은 내부 도구의 경우 어떤 프로바이더/모델을 추천하나요? 이유도 설명해 주세요."
     }'
   ```

기본 프로바이더를 사용할 수 없는 경우, 런타임은 설정된 `fallbackModel`로 폴백할 수 있으며 클라이언트 호출 패턴은 그대로 유지됩니다.

## 4. 구성 포인트

- `.env`와 `config/default.yaml`에 각 프로바이더의 키, 엔드포인트, 타임아웃 등을 한 곳에서 관리합니다.  
- “프로바이더 매트릭스” 페이지를 참고해 사용할 프로바이더와 기능 조합을 선택합니다.  
- 클라이언트 코드에 프로바이더 특화 로직을 하드코딩하지 말고, Agno-Go 런타임을 유일한 통합 계층으로 취급합니다.  

## 5. 테스트와 검증

이 패턴을 실제로 사용하기 전에:

- Quickstart와 유사한 호출 플로우로 라우팅 에이전트의 기본 동작을 검증합니다.  
- 특정 프로바이더의 키를 임시로 제거하고, 폴백이 기대대로 작동하는지 확인합니다.  
- 프로바이더별 알려진 제약(토큰 수, 지연 시간 등)을 기록하여 내부 운영 문서에 남깁니다.  
