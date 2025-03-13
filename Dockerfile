FROM debian:bookworm-slim

ARG HOME_DIR="/home/test"

RUN apt-get update && apt-get install -y --no-install-recommends \
    zsh git bash curl ca-certificates vim

RUN addgroup --system test && adduser --system --ingroup test --shell /bin/zsh --home "$HOME_DIR" test

ENV PATH="$HOME_DIR/.local/bin:$PATH" \
    HOME="$HOME_DIR"

COPY bin/zsh-ai-suggestions "$HOME_DIR/.local/bin/zsh-ai-suggestions"
COPY zsh-ai-suggestions.zsh "$HOME_DIR/zsh-ai-suggestions.zsh"

RUN chown -R test:test "$HOME_DIR" && \
    chmod +x "$HOME_DIR/zsh-ai-suggestions.zsh" && \
    chmod +x "$HOME_DIR/.local/bin/zsh-ai-suggestions"

USER test
WORKDIR "$HOME_DIR"

CMD ["zsh", "-li"]
