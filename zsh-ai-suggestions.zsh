[[ -o interactive ]] || return 0
setopt NO_CHECK_JOBS NO_HUP

: ${ZSH_AI_SUGGESTIONS_BINARY:="$HOME/.local/bin/zsh-ai-suggestions"}
: ${ZSH_AI_SUGGESTIONS_TIMEOUT:=5}
: ${ZSH_AI_SUGGESTIONS_DEBUG:=false}
: ${ZSH_AI_SUGGESTIONS_TMPDIR:="/tmp/zsh-ai-suggestions"}

mkdir -p "$ZSH_AI_SUGGESTIONS_TMPDIR"
AI_INPUT_FILE="$ZSH_AI_SUGGESTIONS_TMPDIR/zsh-ai-input-$$"
AI_OUTPUT_FILE="$ZSH_AI_SUGGESTIONS_TMPDIR/zsh-ai-output-$$"

function cleanup() {
  rm -f "$AI_INPUT_FILE" "$AI_OUTPUT_FILE"
}
trap cleanup EXIT

function log() {
  if [[ "$ZSH_AI_SUGGESTIONS_DEBUG" == "true" ]]; then
    echo "$1" > /dev/tty
  fi
}

function is_backend_running() {
  if command -v pgrep > /dev/null 2>&1; then
    pgrep -f "zsh-ai-suggestions" > /dev/null
    return $?
  else
    ps aux | grep "[z]sh-ai-suggestions" | grep -v grep > /dev/null
    return $?
  fi
}

function start_backend() {
  log "starting backend: $ZSH_AI_SUGGESTIONS_BINARY"
  log "using tmp directory: $ZSH_AI_SUGGESTIONS_TMPDIR"

  export ZSH_AI_SUGGESTIONS_TMPDIR
  
  if [[ ! -x "$ZSH_AI_SUGGESTIONS_BINARY" ]]; then
    log "backend binary not found or not executable: $ZSH_AI_SUGGESTIONS_BINARY"
    return 1
  fi
  
  setopt local_options no_notify no_monitor
  { "$ZSH_AI_SUGGESTIONS_BINARY" > /dev/null 2> /dev/null & } 2>/dev/null
  local pid=$!
  disown %% 2>/dev/null
  log "backend started in background with PID: $pid"
  sleep 0.2
  return 0
}

function suggest() {
  local input="$BUFFER"
  local suggestion=""
  local braille_frames=("⠋" "⠙" "⠹" "⠸" "⠼" "⠴")
  local braille_index=0
  local start_time
  local original_cursor_pos="$CURSOR"

  if [[ -z "$input" ]]; then
    log "empty input, skipping"
    return
  fi

  log "clearing old output file: $AI_OUTPUT_FILE"
  rm -f "$AI_OUTPUT_FILE"

  if ! is_backend_running; then
    start_backend || return 1
  else
    log "backend already running"
  fi
  
  log "writing input: $input"
  echo "$input" > "$AI_INPUT_FILE"

  start_time=$(date +%s)
  while [[ ! -f "$AI_OUTPUT_FILE" ]]; do
    sleep 0.1
    local braille_char="${braille_frames[$braille_index % ${#braille_frames}]}"
    ((braille_index++))

    local buffer_before=""
    local buffer_after=""
    if (( original_cursor_pos >= ${#input} )); then
      buffer_before="$input"
      buffer_after=""
    else
      buffer_before="${input:0:$original_cursor_pos}"
      buffer_after="${input:$original_cursor_pos}"
    fi

    BUFFER="${buffer_before}${braille_char}${buffer_after}"
    CURSOR="$original_cursor_pos"
    zle -R

    if (( $(date +%s) - start_time >= ZSH_AI_SUGGESTIONS_TIMEOUT )); then
      log "timeout waiting for ai response (waited $ZSH_AI_SUGGESTIONS_TIMEOUT seconds)"
      BUFFER="$input"
      CURSOR="$original_cursor_pos"
      zle -R
      return 1
    fi
  done

  local end_time=$(date +%s)
  local duration=$((end_time - start_time))
  log "response file found after ${duration} seconds"

  if [[ -s "$AI_OUTPUT_FILE" ]]; then
    suggestion=$(<"$AI_OUTPUT_FILE")
    if [[ -n "$suggestion" ]]; then
      log "applying suggestion: $suggestion"
      BUFFER="$suggestion"
      CURSOR=${#BUFFER}
      zle -R
    else
      log "empty suggestion received"
    fi
  else
    log "output file is empty"
  fi

  rm -f "$AI_INPUT_FILE" "$AI_OUTPUT_FILE"
}

zle -N suggest
bindkey "^@" suggest