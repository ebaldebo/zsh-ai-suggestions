AI_SUGGESTIONS_BIN=$(command -v zsh-ai-suggestions || echo "$HOME/.local/bin/zsh-ai-suggestions")

coproc "$AI_SUGGESTIONS_BIN"
if [[ $? -ne 0 ]]; then
  echo "Failed to start AI suggestions service"
  return 1
fi

exec 3>&p
exec 4<&p

function cleanup() {
  [[ -n $COPROC_PID ]] && kill "$COPROC_PID" 2>/dev/null
  exec 3>&- 4<&-
}

trap cleanup EXIT SIGTERM SIGINT

function suggest() {
  local input="$BUFFER"
  local suggestion=""
  local old_tmout=$TMOUT

  TMOUT=1
  print -n -- "$input\n" >&3
  read -u4 suggestion || suggestion=""
  TMOUT=$old_tmout

  if [[ -n "$suggestion" ]]; then
    BUFFER="$suggestion"
    zle end-of-line
  fi
}

zle -N suggest
bindkey "^@" suggest
