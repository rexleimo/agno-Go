# Go 세션 서비스

Go 세션 서비스는 Python AgentOS `/sessions` REST API를 그대로 재현하는 독립형 Go
바이너리입니다. AgentOS와 함께 배포하면 PostgreSQL 기반의 고성능 세션 CRUD, 운영용
HTTP 미들웨어, 그리고 로컬부터 Kubernetes까지 아우르는 배포 도구를 활용할 수 있습니다.

## 기능 하이라이트

- **동일한 엔드포인트**: `/sessions` 목록/생성, `/sessions/{id}` 상세, 이름 변경, 삭제,
  `/sessions/{id}/runs` 히스토리, `/healthz` 헬스 체크 제공.
- **Postgres 스토리지**: 타입 안전 DTO와 트랜잭션 세이프 처리를 통해 기존 AgentOS JSON
  계약을 그대로 유지.
- **멀티 데이터베이스 라우팅**: `AGNO_SESSION_DSN_MAP`으로 여러 DSN을 등록하고 `db_id`
  쿼리 파라미터로 요청마다 대상 스토어를 선택.
- **운영 친화 미들웨어**: Chi 라우터가 요청 ID, 구조화 로그, Real IP, 패닉 복구,
  60초 타임아웃을 기본 제공.
- **배포 자산 완비**: 전용 Dockerfile, Postgres 포함 Docker Compose, Helm Chart,
  curl 기반 스모크 테스트 스크립트까지 포함.

## 로컬 빠른 시작

```bash
export AGNO_PG_DSN="postgres://user:pass@localhost:5432/agentos?sslmode=disable"
export AGNO_SERVICE_PORT=8080
go run ./cmd/agentos-session
```

로그에 `Go session service listening on :8080` 메시지가 출력되며
`http://localhost:8080/healthz`에서 JSON 헬스 응답을 확인할 수 있습니다.

### 계약 테스트 정합성

- `make contract-test` — Go 구현과 Python 구현의 응답을 비교하는 계약 테스트 실행.
- `./scripts/test-session-api.sh http://localhost:8080` — curl + jq로 목록/생성/이름변경/삭제 플로우를 순회.

## 설정 레퍼런스

| 환경 변수               | 설명                                                                                 | 기본값     |
|------------------------|--------------------------------------------------------------------------------------|------------|
| `AGNO_PG_DSN`          | 기본 Postgres DSN. `AGNO_SESSION_DSN_MAP`이 없을 때 필수.                           | 없음       |
| `DATABASE_URL`         | `AGNO_PG_DSN` 미설정 시 사용하는 Heroku 스타일 DSN.                                  | 없음       |
| `AGNO_SESSION_DSN_MAP` | `{"dbID":"dsn"}` 형태 JSON으로 다중 데이터베이스 라우팅을 활성화하고 `db_id` 선택 허용. | 없음       |
| `AGNO_DEFAULT_DB_ID`   | `AGNO_SESSION_DSN_MAP` 사용 시 기본 데이터베이스 ID.                                 | 첫 번째 키 |
| `AGNO_SERVICE_PORT`    | HTTP 바인딩 포트.                                                                    | `8080`     |

`AGNO_SESSION_DSN_MAP`을 설정했다면 `/sessions?type=agent&db_id=analytics`처럼 요청마다
타깃 스토어를 지정할 수 있습니다.

## API 개요

| Endpoint                | Method | 설명                                                                                   |
|-------------------------|--------|----------------------------------------------------------------------------------------|
| `/healthz`              | GET    | `{"status":"ok"}` 를 반환하는 헬스 체크.                                              |
| `/sessions`             | GET    | 페이징 및 필터(`type`, `component_id`, `user_id`, `session_name`, `sort_by`, `db_id`) 지원. |
| `/sessions`             | POST   | 상태, 메타데이터, 사전 준비된 runs/summary를 포함해 세션 생성.                          |
| `/sessions/{id}`        | GET    | 타입/ID 기준 상세 조회. `db_id`로 데이터베이스 지정 가능.                               |
| `/sessions/{id}`        | DELETE | 세션과 히스토리 삭제.                                                                   |
| `/sessions/{id}/rename` | POST   | `session_name` 업데이트.                                                                |
| `/sessions/{id}/runs`   | GET    | 저장된 run 히스토리를 반환하며 Python AgentOS와 동일한 구조 유지.                       |

모든 쓰기 요청은 Python 픽스처와 동일한 JSON 형식을 요구하며, 계약 테스트에서 자동 검증합니다.

## Docker Compose 실행

`docker-compose.session.yml`을 사용해 Postgres와 Go 세션 서비스를 함께 기동할 수 있습니다.

```bash
docker compose -f docker-compose.session.yml up --build
```

서비스가 준비되면 `http://localhost:8080`에 접속 후 스크립트를 실행해 점검하세요.

## Helm 배포

`deploy/helm/agno-session/` Chart로 Kubernetes에 손쉽게 배포할 수 있습니다. 아래는 기본 명령 예시입니다.

```bash
helm upgrade --install agno-session ./deploy/helm/agno-session \
  --set image.repository=ghcr.io/<org>/agno-session \
  --set image.tag=v1.2.9 \
  --set config.dsn="postgres://user:pass@postgres:5432/agentos?sslmode=disable"
```

프로브, 레플리카 수, 서비스 노출 방식 등의 세부 설정은 `values.yaml`에서 조정할 수 있습니다.

## 운영 체크리스트

- 스테이징 환경에서 `make contract-test`를 실행해 호환성을 검증합니다.
- 실제 트래픽 일부를 미러링해 Python 서비스와 JSON 응답을 비교합니다.
- Postgres 연결 및 지연 시간을 모니터링하며 필요 시 커넥션 풀 등을 도입합니다.
- `/healthz`와 HTTP 로그를 기반으로 대시보드 및 알림을 구성합니다.
- 프로덕션 안정화 전까지는 Python 런타임을 대비책으로 유지하는 것이 좋습니다.
