[[ -o interactive ]] || return 0
setopt NO_CHECK_JOBS NO_HUP

local _plugin_script_dir="${${(%):-%x}:h}"
local _plugin_name="zsh-ai-suggestions"

: ${ZSH_AI_SUGGESTIONS_BINARY:="$_plugin_script_dir/$_plugin_name"}
: ${ZSH_AI_SUGGESTIONS_TIMEOUT:=5}
: ${ZSH_AI_SUGGESTIONS_DEBUG:=false}
: ${ZSH_AI_SUGGESTIONS_REPO:="ebaldebo/$_plugin_name"}
: ${ZSH_AI_SUGGESTIONS_SERVER_URL:="http://localhost:5555/suggest"}
: ${ZSH_AI_SUGGESTIONS_SERVER_INIT_WAIT:=1}

function _zsh_ai_log() {
  if [[ "$ZSH_AI_SUGGESTIONS_DEBUG" == "true" ]]; then
    printf "[%s %s] %s\n" "$_plugin_name" "$(date +'%H:%M:%S.%N' 2>/dev/null || date +'%H:%M:%S')" "$*" > /dev/tty
  fi
}

function download_binary() {
  local binary_path="$1"
  local binary_dir="$(dirname "$binary_path")"
  local os_name="$(uname -s)"
  local arch="$(uname -m)"

  _zsh_ai_log "detected system: ${os_name}_${arch}"
  local archive_name="${_plugin_name}_${os_name}_${arch}.tar.gz"
  local download_url="https://github.com/${ZSH_AI_SUGGESTIONS_REPO}/releases/latest/download/${archive_name}"
  local temp_archive="$binary_dir/${archive_name}"

  _zsh_ai_log "downloading ${_plugin_name} binary from ${download_url} to $temp_archive..."

  mkdir -p "$binary_dir" || {
    echo "$_plugin_name: error: could not create/access binary directory $binary_dir" >&2
    return 1
  }

  if command -v curl >/dev/null 2>&1; then
    curl -L -s -o "$temp_archive" "$download_url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$temp_archive" "$download_url"
  else
    echo "$_plugin_name: missing dependency for download: curl or wget" >&2
    return 1
  fi

  if [[ ! -f "$temp_archive" ]]; then
    echo "$_plugin_name: error: failed to download the archive from $download_url." >&2
    return 1
  fi

  _zsh_ai_log "extracting binary from archive $temp_archive into $binary_dir..."
  if ! tar -xzf "$temp_archive" -C "$binary_dir"; then
    echo "$_plugin_name: error: failed to extract the archive." >&2
    rm -f "$temp_archive"
    return 1
  fi

  rm -f "$temp_archive"

  if [[ ! -f "$binary_path" ]]; then
    echo "$_plugin_name: error: binary not found at '$binary_path' after extraction." >&2
    return 1
  fi

  chmod +x "$binary_path"

  if [[ ! -x "$binary_path" ]]; then
    echo "$_plugin_name: error: failed to make the binary '$binary_path' executable." >&2
    return 1
  fi

  _zsh_ai_log "binary successfully downloaded, extracted, and installed at $binary_path."
  return 0
}

if [[ ! -x "$ZSH_AI_SUGGESTIONS_BINARY" ]]; then
  _zsh_ai_log "binary $ZSH_AI_SUGGESTIONS_BINARY not found or not executable. attempting download..."
  download_binary "$ZSH_AI_SUGGESTIONS_BINARY" || {
    echo "$_plugin_name: error: failed to download zsh-ai-suggestions binary. plugin will not work." >&2
    return 1
  }
fi

function is_backend_running() {
  if command -v pgrep > /dev/null 2>&1; then
    pgrep -f "$ZSH_AI_SUGGESTIONS_BINARY" > /dev/null
    return $?
  else
    ps aux | grep "[$(basename "$ZSH_AI_SUGGESTIONS_BINARY")]" | grep -v grep > /dev/null
    return $?
  fi
}

function start_backend() {
  _zsh_ai_log "starting backend server: $ZSH_AI_SUGGESTIONS_BINARY"

  if [[ ! -x "$ZSH_AI_SUGGESTIONS_BINARY" ]]; then
    _zsh_ai_log "backend binary not found or not executable: $ZSH_AI_SUGGESTIONS_BINARY"
    return 1
  fi

  setopt local_options no_notify no_monitor
  { "$ZSH_AI_SUGGESTIONS_BINARY" > /dev/null 2>&1 & } 2>/dev/null
  local server_pid=$!
  disown %% 2>/dev/null
  _zsh_ai_log "backend server process started in background with pid: $server_pid"

  _zsh_ai_log "waiting $ZSH_AI_SUGGESTIONS_SERVER_INIT_WAIT second(s) for server to initialize..."
  sleep "$ZSH_AI_SUGGESTIONS_SERVER_INIT_WAIT"

  if ! kill -0 $server_pid 2>/dev/null; then
    _zsh_ai_log "backend server (pid $server_pid) failed to start or exited quickly after launch."
    return 1
  fi
  _zsh_ai_log "backend server (pid $server_pid) presumed to be initializing or running."
  return 0
}

function suggest() {
  local input="$BUFFER"
  local suggestion=""
  local curl_exit_status

  if ! command -v curl >/dev/null 2>&1; then
    _zsh_ai_log "error: curl command is not available. cannot make http requests."
    return 1
  fi

  if [[ -z "$input" ]]; then
    _zsh_ai_log "empty input, skipping."
    return 0
  fi

  if ! is_backend_running; then
    _zsh_ai_log "backend server not running. attempting to start..."
    if ! start_backend; then
      _zsh_ai_log "failed to start backend server."
      return 1
    fi
  else
    _zsh_ai_log "backend server appears to be running."
  fi

  _zsh_ai_log "requesting suggestion for: '$input' from $ZSH_AI_SUGGESTIONS_SERVER_URL"
  _zsh_ai_log "sending raw input as body: '$input'. waiting for response..."

  local old_buffer="$BUFFER"
  local old_cursor="$CURSOR"
 
  BUFFER="${old_buffer} Loading..."
  CURSOR=${#BUFFER}
  zle -R
 
  suggestion=$(echo -n "$input" | curl --silent --show-error --fail --location \
    --max-time "$ZSH_AI_SUGGESTIONS_TIMEOUT" \
    --request POST \
    --header "Content-Type: text/plain" \
    --header "Accept: text/plain" \
    --data-binary @- \
    "$ZSH_AI_SUGGESTIONS_SERVER_URL" 2>/dev/null)
  curl_exit_status=$?
 
  BUFFER="$old_buffer"
  CURSOR="$old_cursor"
  zle -R

  _zsh_ai_log "curl command finished with exit status: $curl_exit_status"

  if [[ "$curl_exit_status" -ne 0 ]]; then
    _zsh_ai_log "curl failed with status $curl_exit_status."
    if [[ -n "$suggestion" ]]; then
        _zsh_ai_log "curl output: '$suggestion'"
    fi
    case "$curl_exit_status" in
      6) _zsh_ai_log "could not resolve host. check url ($ZSH_AI_SUGGESTIONS_SERVER_URL) or network.";;
      7) _zsh_ai_log "failed to connect. is the server at $ZSH_AI_SUGGESTIONS_SERVER_URL ready and accessible?";;
      22) _zsh_ai_log "http error from server";;
      28) _zsh_ai_log "curl explicit timeout after $ZSH_AI_SUGGESTIONS_TIMEOUT seconds.";;
      *)  _zsh_ai_log "curl failed with an unhandled error code: $curl_exit_status. see curl man page for details.";;
    esac
    return 1
  fi

  suggestion="${suggestion%"${suggestion##*[![:space:]]}"}"

  if [[ -n "$suggestion" ]]; then
    _zsh_ai_log "applying suggestion: '$suggestion'"
    BUFFER="$suggestion"
    CURSOR=${#BUFFER}
    zle -R
  else
    _zsh_ai_log "empty suggestion received from server or curl output was empty."
  fi

  return 0
}

zle -N suggest
bindkey "^_" suggest
