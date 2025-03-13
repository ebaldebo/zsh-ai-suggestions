FROM debian:bookworm-slim

ARG HOME_DIR="/home/test"

RUN apt-get update && apt-get install -y --no-install-recommends \
    zsh git bash curl ca-certificates vim locales

RUN echo "en_US.UTF-8 UTF-8" > /etc/locale.gen && \
    locale-gen

ENV LANG="en_US.UTF-8"
ENV LC_ALL="en_US.UTF-8"

RUN addgroup --system test && adduser --system --ingroup test --shell /bin/zsh --home "$HOME_DIR" test

ENV PATH="$HOME_DIR/.local/bin:$PATH" \
    HOME="$HOME_DIR"

COPY bin/zsh-ai-suggestions "$HOME_DIR/.local/bin/zsh-ai-suggestions"
COPY zsh-ai-suggestions.zsh "$HOME_DIR/zsh-ai-suggestions.zsh"
COPY zsh-ai-suggestions.plugin.zsh "$HOME_DIR/zsh-ai-suggestions.plugin.zsh"

RUN chown -R test:test "$HOME_DIR" && \
    chmod +x "$HOME_DIR/zsh-ai-suggestions.zsh" && \
    chmod +x "$HOME_DIR/.local/bin/zsh-ai-suggestions" && \
    chmod +x "$HOME_DIR/zsh-ai-suggestions.plugin.zsh"

# Dockerfile with zinit
RUN echo 'ZINIT_HOME="${XDG_DATA_HOME:-${HOME}/.local/share}/zinit/zinit.git"' > ~/.zshrc && \
echo '' >> ~/.zshrc && \
echo '# Download Zinit if it does not exist' >> ~/.zshrc && \
echo 'if [ ! -d "$ZINIT_HOME" ]; then' >> ~/.zshrc && \
echo '   mkdir -p "$(dirname $ZINIT_HOME)"' >> ~/.zshrc && \
echo '   git clone https://github.com/zdharma-continuum/zinit.git "$ZINIT_HOME"' >> ~/.zshrc && \
echo 'fi' >> ~/.zshrc && \
echo '' >> ~/.zshrc && \
echo '# Source/Load zinit' >> ~/.zshrc && \
echo 'source "${ZINIT_HOME}/zinit.zsh"' >> ~/.zshrc && \
echo '' >> ~/.zshrc && \
echo 'zinit light $HOME' >> ~/.zshrc

USER test
WORKDIR "$HOME_DIR"

CMD ["zsh", "-li"]
