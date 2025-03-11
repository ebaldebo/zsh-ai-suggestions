FROM alpine:latest

RUN apk add --no-cache zsh git bash curl && \
    adduser -D -s /bin/zsh test

COPY install.sh home/test/install.sh
COPY zsh-ai-suggestions.zsh /home/test/zsh-ai-suggestions.zsh
RUN chown -R test:test /home/test
RUN chmod +x /home/test/install.sh

USER test
WORKDIR /home/test

CMD ["zsh", "-i"]
