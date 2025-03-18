[[ -o interactive ]] || return 0

setopt NO_CHECK_JOBS NO_HUP

: ${ZSH_AI_SUGGESTIONS_BINARY:="$HOME/.local/bin/zsh-ai-suggestions"}
: ${ZSH_AI_SUGGESTIONS_TIMEOUT:=5}
: ${ZSH_AI_SUGGESTIONS_DEBUG:=false}
: ${ZSH_AI_SUGGESTIONS_LOG_RETENTION_DAYS:=1}
: ${ZSH_AI_SUGGESTIONS_CLEANUP_ON_EXIT:=false}
: ${ZSH_AI_SUGGESTIONS_LOG_TO_FILES:=false}
: ${ZSH_AI_SUGGESTIONS_LOG_LEVEL:="info"}

AI_TMP_DIR="/tmp/zsh-ai-suggestions"
AI_INPUT_FILE="$AI_TMP_DIR/zsh-ai-input-$$"
AI_OUTPUT_FILE="$AI_TMP_DIR/zsh-ai-output-$$"
mkdir -p "$AI_TMP_DIR"

function cleanup() {
  rm -f "$AI_INPUT_FILE" "$AI_OUTPUT_FILE"

  if [[ "$ZSH_AI_SUGGESTIONS_CLEANUP_ON_EXIT" == "true" ]]; then
    local terminal_count=$(pgrep -f "zsh" | wc -l)
    if [[ "$terminal_count" -eq 1 ]]; then
      log "last terminal closing, stopping backend"
      pkill -f "zsh-ai-suggestions" 2>/dev/null
    fi
  fi
}
trap cleanup EXIT

function manage_logs() {
  find "$AI_TMP_DIR" -name "backend.*.log" -type f -mtime +${ZSH_AI_SUGGESTIONS_LOG_RETENTION_DAYS} -delete 2>/dev/null
  
  local log_count=$(find "$AI_TMP_DIR" -name "backend.*.log" | wc -l)
  if [[ "$log_count" -gt 20 ]]; then
    find "$AI_TMP_DIR" -name "backend.*.log" -type f | sort | head -n $(($log_count - 20)) | xargs rm -f 2>/dev/null
  fi
}

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
  
  if [[ ! -x "$ZSH_AI_SUGGESTIONS_BINARY" ]]; then
    log "backend binary not found or not executable: $ZSH_AI_SUGGESTIONS_BINARY"
    return 1
  fi
  
  setopt local_options
  setopt no_notify no_monitor
  
  if [[ "$ZSH_AI_SUGGESTIONS_LOG_TO_FILES" == "true" ]]; then
    local timestamp=$(date +%Y%m%d%H%M%S)
    local stdout_log="$AI_TMP_DIR/backend.stdout.$timestamp.log"
    local stderr_log="$AI_TMP_DIR/backend.stderr.$timestamp.log"
    
    find "$AI_TMP_DIR" -name "backend.*.log" -type f -mtime +1 -delete 2>/dev/null
    
    { "$ZSH_AI_SUGGESTIONS_BINARY" > "$stdout_log" 2> "$stderr_log" & } 2>/dev/null
    local pid=$!
    
    disown %% 2>/dev/null
    
    log "backend started in background with PID: $pid (logging to files)"
  else
    { "$ZSH_AI_SUGGESTIONS_BINARY" > /dev/null 2> /dev/null & } 2>/dev/null
    local pid=$!

    disown %% 2>/dev/null
    
    log "backend started in background with PID: $pid (logging disabled)"
  fi
  
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