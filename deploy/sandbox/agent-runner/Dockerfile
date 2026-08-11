# agent-runner: the sandbox base image for pkg/hno/tools/run.
#
# Build with a deterministic tag and pin the base image version:
#   docker build -t agent-runner:1 deploy/sandbox/agent-runner
#   podman build -t agent-runner:1 deploy/sandbox/agent-runner
#
# The sandbox core never parses payloads; language support lives here.
# Add runtimes or system tools to this image, not to the executor.
FROM python:3.12-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    nodejs \
    bash \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Payloads arrive on stdin; entrypoints are declared by runtime templates
# in pkg/hno/tools/run (python3 -, node -, bash -s).
CMD ["python3", "-"]
