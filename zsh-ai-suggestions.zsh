# zsh-ai-suggestions plugin
[[ -o interactive ]] || return 0

PLUGIN_DIR="${${(%):-%N}:A:h}"

_zsh_ai_suggestions_install() {
  local install_dir="$HOME/.local/bin"
  local bin_path="$install_dir/zsh-ai-suggestions"

  if [[ ! -x "$bin_path" ]]; then
    echo "zsh-ai-suggestions binary not found, installing..."
    mkdir -p "$install_dir"

    if [[ -f "$PLUGIN_DIR/install.sh" ]]; then
      bash "$PLUGIN_DIR/install.sh"
    else
      echo "Error: install.sh not found"
      return 1
    fi
  fi

  return 0
}

_zsh_ai_suggestions_install || return 1

AI_SUGGESTIONS_BIN=$(command -v zsh-ai-suggestions || echo "$HOME/.local/bin/zsh-ai-suggestions")

if [[ ! -x "$AI_SUGGESTIONS_BIN" ]]; then
  echo "Error: zsh-ai-suggestions binary not found or not executable"
  return 1
fi

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

  if ! { true >&3 2>/dev/null; }; then
    echo "Error: AI suggestions service is not running!"
    return 1
  fi

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
