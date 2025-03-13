#!/usr/bin/env zsh
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
  local dots=""
  local max_dots=5

  log "clearing old output file: $AI_OUTPUT_FILE"
  rm -f "$AI_OUTPUT_FILE"

  log "writing input: $input"
  echo "$input" > "$AI_INPUT_FILE"

  local start_time=$(date +%s)
  while [[ ! -f "$AI_OUTPUT_FILE" ]]; do
    sleep 0.2
    if [[ ${#dots} -ge $max_dots ]]; then
      dots=""
    fi
    dots+="."

    BUFFER="$input$dots"
    CURSOR=${#BUFFER}
    zle -R

    if (( $(date +%s) - start_time >= 5 )); then
      log "timeout waiting for AI response (Waited 5 seconds)"
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
