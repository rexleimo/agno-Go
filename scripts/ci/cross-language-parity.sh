#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH='' cd -- "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CACHE_DIR="$SCRIPT_DIR/.cache"
RESULT_FILE="$CACHE_DIR/parity_results.json"

usage() {
    cat <<'EOF'
Usage: cross-language-parity.sh --fixture <file> [--python "<cmd>"] [--go "<cmd>"]

Options:
  --fixture <file>   ParityFixture YAML/JSON，必须提供，建议使用绝对路径
  --python "<cmd>"   运行 Python 基线的命令（默认：python -m agno.tests.contracts.run）
  --go "<cmd>"       运行 Go 测试命令（默认：go test ./go/... -run TestParity -count=1）
  -h, --help         查看帮助

占位符 {{fixture}} 可以嵌入在命令中指定 fixture 路径；否则脚本通过 AGNO_FIXTURE 环境变量传递。
EOF
}

FIXTURE_PATH=""
PYTHON_CMD="python -m agno.tests.contracts.run"
GO_CMD="go test ./go/... -run TestParity -count=1"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --fixture)
            if [[ $# -lt 2 ]]; then
                echo "ERROR: --fixture 需要一个参数" >&2
                exit 1
            fi
            FIXTURE_PATH="$2"
            shift 2
            ;;
        --python)
            if [[ $# -lt 2 ]]; then
                echo "ERROR: --python 需要一个参数" >&2
                exit 1
            fi
            PYTHON_CMD="$2"
            shift 2
            ;;
        --go)
            if [[ $# -lt 2 ]]; then
                echo "ERROR: --go 需要一个参数" >&2
                exit 1
            fi
            GO_CMD="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "ERROR: 未知参数：$1" >&2
            usage
            exit 1
            ;;
    esac
done

if [[ -z "$FIXTURE_PATH" ]]; then
    echo "ERROR: 必须提供 --fixture <file>" >&2
    usage
    exit 1
fi

if [[ ! -f "$FIXTURE_PATH" ]]; then
    echo "ERROR: 找不到 fixture 文件：$FIXTURE_PATH" >&2
    exit 1
fi

ABS_FIXTURE="$(python3 - <<'PY' "$FIXTURE_PATH"
import os, sys
print(os.path.abspath(sys.argv[1]))
PY
)"

mkdir -p "$CACHE_DIR"

PYTHON_OUTPUT="$CACHE_DIR/python_output.json"
GO_OUTPUT="$CACHE_DIR/go_output.json"

substitute_fixture() {
    local cmd="$1"
    if [[ "$cmd" == *"{{fixture}}"* ]]; then
        echo "${cmd//\{\{fixture\}\}/$ABS_FIXTURE}"
    else
        echo "$cmd"
    fi
}

run_command() {
    local label="$1"
    local cmd="$2"
    local output="$3"
    echo "[parity] Running $label command: $cmd"
    if ! AGNO_FIXTURE="$ABS_FIXTURE" bash -c "$cmd" >"$output"; then
        echo "ERROR: $label command failed" >&2
        exit 1
    fi
}

PYTHON_PREPARED="$(substitute_fixture "$PYTHON_CMD")"
GO_PREPARED="$(substitute_fixture "$GO_CMD")"

run_command "python" "$PYTHON_PREPARED" "$PYTHON_OUTPUT"
run_command "go" "$GO_PREPARED" "$GO_OUTPUT"

python3 - <<'PY' "$PYTHON_OUTPUT" "$GO_OUTPUT" "$RESULT_FILE" "$ABS_FIXTURE" "$PYTHON_PREPARED" "$GO_PREPARED"
import json
import sys
from datetime import datetime, timezone

python_path, go_path, result_file, fixture_path, python_cmd, go_cmd = sys.argv[1:7]

def load_payload(path, label):
    try:
        with open(path, 'r', encoding='utf-8') as fh:
            return json.load(fh)
    except json.JSONDecodeError as exc:
        raise SystemExit(f"{label} output is not valid JSON: {exc}")

py_payload = load_payload(python_path, "Python")
go_payload = load_payload(go_path, "Go")

diffs = []

def compare(path, py_val, go_val):
    if py_val != go_val:
        diffs.append({
            "path": path,
            "python": py_val,
            "go": go_val,
        })

for section in ("outputs", "metrics", "session_record"):
    compare(section, py_payload.get(section), go_payload.get(section))

status = "pass" if not diffs else "fail"

result = {
    "generated_at": datetime.now(tz=timezone.utc).isoformat(),
    "fixture": fixture_path,
    "python_command": python_cmd,
    "go_command": go_cmd,
    "status": status,
    "diffs": diffs,
    "python_output_path": python_path,
    "go_output_path": go_path,
}

with open(result_file, 'w', encoding='utf-8') as fh:
    json.dump(result, fh, ensure_ascii=False, indent=2)

json.dump(result, sys.stdout, ensure_ascii=False, indent=2)
sys.stdout.write("\n")
PY

echo "Parity 结果已写入：$RESULT_FILE"
