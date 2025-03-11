FROM alpine:latest

RUN apk add --no-cache zsh git bash curl && \
    adduser -D -s /bin/zsh test

#COPY bin/zsh-ai-suggestions /home/test/.local/bin/zsh-ai-suggestions
#COPY zsh-ai-suggestions.zsh /home/test/zsh-ai-suggestions.zsh
RUN chown -R test:test /home/test

USER test
WORKDIR /home/test

#RUN echo "source /home/test/zsh-ai-suggestions.zsh" >> /home/test/.zshrc

CMD ["zsh", "-i"]
