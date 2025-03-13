[[ -o interactive ]] || return 0

AI_TMP_DIR="/tmp/zsh-ai-suggestions"
AI_INPUT_FILE="$AI_TMP_DIR/zsh-ai-input-$$"
AI_OUTPUT_FILE="$AI_TMP_DIR/zsh-ai-output-$$"

mkdir -p "$AI_TMP_DIR"

DEBUG=false

function log() {
  if [[ "$DEBUG" == "true" ]]; then
    echo "$1" > /dev/tty
  fi
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

  log "writing input: $input"
  echo "$input" > "$AI_INPUT_FILE"

  start_time=$(date +%s)

  while [[ ! -f "$AI_OUTPUT_FILE" ]]; do
    sleep 0.2

    local braille_char="${braille_frames[$braille_index % ${#braille_frames}]}"
    braille_index=$((braille_index + 1))

    local buffer_before=""
    local buffer_after=""

    if [[ "$original_cursor_pos" -ge "${#input}" ]]; then
      buffer_before="$input"
      buffer_after=""
    else
      buffer_before="${input[1,$((original_cursor_pos - 1))]}"
      buffer_after="${input[$((original_cursor_pos + 1)),-1]}"
    fi

    BUFFER="${buffer_before}${braille_char}${buffer_after}"

    CURSOR="$original_cursor_pos"

    zle -R

    if (( $(date +%s) - start_time >= 5 )); then
      log "timeout waiting for AI response (Waited 5 seconds)"
      BUFFER="$input timeout"
      CURSOR=${#BUFFER}
      zle -R
      return
    fi
  done

  local end_time=$(date +%s)
  local duration=$((end_time - start_time))
  log "response file found after ${duration} seconds"

  if [[ -s "$AI_OUTPUT_FILE" ]]; then
    suggestion=$(<"$AI_OUTPUT_FILE")
  else
    log "output file is empty"
  fi

  if [[ -n "$suggestion" ]]; then
    log "applying suggestion: $suggestion"
    BUFFER="$suggestion"
    CURSOR=${#BUFFER}
    zle -R
  else
    log "no suggestion received"
  fi
}

zle -N suggest
bindkey "^@" suggest
