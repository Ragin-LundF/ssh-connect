#!/bin/bash

CONFIG_FILE="ssh_connect_server.toml"
DRY_RUN=0
INIT_CONFIG=0
ADD_SERVER=0

# Arrays that hold parsed server data by index.
declare -a SERVER_KEYS SERVER_NAMES SERVER_IPS SERVER_USERS SERVER_CERTS

fail() {
  echo "Error: $1" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Usage: ./ssh_connect.sh [--dry-run] [--config <path>]

Options:
  --dry-run   Show the SSH command after selection, but do not execute it.
  --config    Use a custom TOML config file.
  --init      Create an example config file and exit.
  --add       Add a new server entry using dialog prompts.
  -h, --help  Show this help message.
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --dry-run)
        DRY_RUN=1
        ;;
      --config)
        shift
        [[ $# -gt 0 ]] || fail "Missing value for --config"
        CONFIG_FILE="$1"
        ;;
      --init)
        INIT_CONFIG=1
        ;;
      --add)
        ADD_SERVER=1
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        fail "Unknown option: $1"
        ;;
    esac
    shift
  done
}

init_config_file() {
  [[ -e "$CONFIG_FILE" ]] && fail "Config file already exists: ${CONFIG_FILE}"

  cat > "$CONFIG_FILE" <<'EOF'
[server.app_prod]
name = "App Production"
ip = "203.0.113.10"
user = "deploy"
certificate = "/Users/youruser/.ssh/app_prod.pem"

[server.db_prod]
name = "DB Production"
ip = "203.0.113.20"
user = "dbadmin"
certificate = "/Users/youruser/.ssh/db_prod.pem"

[group.production]
name = "Production"
servers = ["server.app_prod", "server.db_prod"]
group_certificate = "/Users/youruser/.ssh/prod_shared.pem"

# Optional per-server override example
[server.db_prod.override]
certificate = "/Users/youruser/.ssh/db_prod_override.pem"
EOF

  echo "Example config created: ${CONFIG_FILE}"
}

escape_toml_string() {
  local value="$1"
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  printf '%s' "$value"
}

prompt_required_input() {
  local title="$1"
  local prompt="$2"
  local input

  while true; do
    input=$(dialog \
      --clear \
      --backtitle "SSH Connect" \
      --title "$title" \
      --inputbox "$prompt" \
      10 70 \
      2>&1 >/dev/tty
    ) || return 1

    if [[ -n "$input" ]]; then
      printf '%s\n' "$input"
      return 0
    fi

    dialog \
      --clear \
      --backtitle "SSH Connect" \
      --title "$title" \
      --msgbox "This field is required." \
      8 40
  done
}

add_new_server_dialog() {
  local alias name ip user
  local safe_name safe_ip safe_user

  alias=$(prompt_required_input "Add Server" "Alias (used as [server.<alias>]):") || return 1

  if [[ ! "$alias" =~ ^[A-Za-z0-9_-]+$ ]]; then
    dialog \
      --clear \
      --backtitle "SSH Connect" \
      --title "Invalid Alias" \
      --msgbox "Alias may only contain letters, numbers, underscore and dash." \
      9 72
    return 1
  fi

  if yq -e ".server.\"$alias\"" "$CONFIG_FILE" >/dev/null 2>&1; then
    dialog \
      --clear \
      --backtitle "SSH Connect" \
      --title "Alias Exists" \
      --msgbox "A server with alias '$alias' already exists." \
      8 60
    return 1
  fi

  name=$(prompt_required_input "Add Server" "Display name:") || return 1
  ip=$(prompt_required_input "Add Server" "IP address or hostname:") || return 1
  user=$(prompt_required_input "Add Server" "SSH user:") || return 1

  safe_name=$(escape_toml_string "$name")
  safe_ip=$(escape_toml_string "$ip")
  safe_user=$(escape_toml_string "$user")

  cat >> "$CONFIG_FILE" <<EOF

[server.$alias]
name = "$safe_name"
ip = "$safe_ip"
user = "$safe_user"
EOF

  dialog \
    --clear \
    --backtitle "SSH Connect" \
    --title "Server Added" \
    --msgbox "Server '$alias' was added to ${CONFIG_FILE}." \
    8 70
}

check_dependencies() {
  command -v yq >/dev/null 2>&1 || fail "'yq' is required but was not found."
  command -v dialog >/dev/null 2>&1 || fail "'dialog' is required but was not found."
}

load_servers() {
  local key name ip user cert

  while IFS=$'\t' read -r key name ip user cert; do

    # Skip incomplete entries to avoid broken SSH commands.
    [[ -z "$key" || -z "$ip" || -z "$user" ]] && continue

    SERVER_KEYS+=("$key")
    SERVER_NAMES+=("${name:-$key}")
    SERVER_IPS+=("$ip")
    SERVER_USERS+=("$user")
    SERVER_CERTS+=("$cert")
  done < <(yq -r '.server | keys[] as $k | [$k, .[$k].name, .[$k].ip, .[$k].user, (.[$k].certificate // "")] | @tsv' "$CONFIG_FILE")

  ((${#SERVER_KEYS[@]} > 0)) || fail "No valid server entries found in ${CONFIG_FILE}."
}

select_server_index() {
  local menu_args=()
  local i choice

  for i in "${!SERVER_KEYS[@]}"; do
    menu_args+=(
      "$i"
      "${SERVER_NAMES[$i]} (${SERVER_USERS[$i]}@${SERVER_IPS[$i]})"
    )
  done

  choice=$(dialog \
    --clear \
    --backtitle "SSH Connect" \
    --title "Select Server" \
    --menu "Choose a server to connect:" \
    18 78 10 \
    "${menu_args[@]}" \
    2>&1 >/dev/tty
  ) || return 1

  printf '%s\n' "$choice"
}

run_selected_server() {
  local index="$1"
  local user ip cert
  local -a ssh_cmd

  user="${SERVER_USERS[$index]}"
  ip="${SERVER_IPS[$index]}"
  cert="${SERVER_CERTS[$index]}"

  clear
  echo "Connecting to ${SERVER_NAMES[$index]} (${user}@${ip})"

  if [[ -n "$cert" && ! -f "$cert" ]]; then
    if (( DRY_RUN )); then
      echo "Warning: Certificate file not found: ${cert}" >&2
    else
      fail "Certificate file not found: ${cert}"
    fi
  fi

  if [[ -n "$cert" ]]; then
    ssh_cmd=(ssh -o "IdentityFile=$cert" "${user}@${ip}")
  else
    ssh_cmd=(ssh "${user}@${ip}")
  fi

  if (( DRY_RUN )); then
    echo "Dry-run enabled. Command will not be executed."
    printf 'Would run:'
    printf ' %q' "${ssh_cmd[@]}"
    printf '\n'
    exit 0
  fi

  exec "${ssh_cmd[@]}"
}

main() {
  parse_args "$@"

  if (( INIT_CONFIG )); then
    init_config_file
    exit 0
  fi

  [[ -f "$CONFIG_FILE" ]] || fail "Config file not found: ${CONFIG_FILE}"

  check_dependencies

  if (( ADD_SERVER )); then
    if ! add_new_server_dialog; then
      clear
      echo "Cancelled."
      exit 0
    fi
    clear
    echo "Server added to ${CONFIG_FILE}."
    exit 0
  fi

  load_servers

  local selected_index
  if ! selected_index=$(select_server_index); then
    clear
    echo "Cancelled."
    exit 0
  fi

  run_selected_server "$selected_index"
}

main "$@"
