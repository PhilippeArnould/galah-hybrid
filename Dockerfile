FROM golang:latest
WORKDIR /galah
COPY . .

RUN go mod download && \
    MOD="$(go list -m -f '{{.Dir}}' github.com/tmc/langchaingo)" && \
    FILE="$MOD/llms/ollama/internal/ollamaclient/types.go" && \
    sed -i "/^type ChatRequest struct {/a\\	Think bool \`json:\\\"think\\\"\`" "$FILE" && \
    echo "===== CHATREQUEST PATCHED =====" && \
    sed -n "/type ChatRequest struct {/,/^}/p" "$FILE" && \
    mkdir -p bin && \
    go build -o bin/galah ./cmd/galah


ENTRYPOINT ["./bin/galah"]


