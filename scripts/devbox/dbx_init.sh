#!/usr/bin/env bash
# --
# @version: 2.3.0
# @purpose: Devbox initialization with status matrix focusing on Claude CLI.
# --

set -euo pipefail

# --------------------------------------------------------------------
# Constants / metadata
# --------------------------------------------------------------------
# Important: paths are relative to repo root
C_DBX_INIT_PATH_ROOT="scripts/devbox"
C_DBX_INIT_ENTRYPOINT="dbx_init.sh"

# Devbox / project metadata (can be overridden from env)
C_DBX_META_TEAM="${C_DBX_META_TEAM:-MPOWR-IT}"
C_DBX_META_TEAM_ID="${C_DBX_META_TEAM_ID:-hetzner-fip-controller}"
C_DBX_META_VERSION="${C_DBX_META_VERSION:-$(cat VERSION 2>/dev/null || echo "dev")}"

# --------------------------------------------------------------------
# Minimal colors & icon themes
# --------------------------------------------------------------------
# export DBX_ICON_SET=ascii    # or: ticks / blocks
# export DBX_NO_COLOR=1        # monochrome
# --------------------------------------------------------------------

: "${DBX_ICON_SET:=ticks}"   # ascii | ticks | blocks
: "${DBX_NO_COLOR:=0}"

if [ "${DBX_NO_COLOR}" = "1" ]; then
  GREEN=""; RED=""; YELLOW=""; BLUE=""; GREY=""; BOLD=""; RESET=""
else
  GREEN="\033[32m"; RED="\033[31m"; YELLOW="\033[33m"; BLUE="\033[34m"; GREY="\033[90m"; BOLD="\033[1m"; RESET="\033[0m"
fi

declare -A STATUS

record_ok()   { STATUS["$1"]="OK|$2"; }
record_fail() { STATUS["$1"]="FAIL|$2"; }
record_warn() { STATUS["$1"]="WARN|$2"; }

_icon_for() {
  case "${DBX_ICON_SET}:${1}" in
    ascii:OK)   echo "OK"  ;;
    ascii:FAIL) echo "ERR" ;;
    ascii:WARN) echo "WRN" ;;
    ticks:OK)   echo "✓"   ;;
    ticks:FAIL) echo "✗"   ;;
    ticks:WARN) echo "!"   ;;
    blocks:OK)  echo "■"   ;;
    blocks:FAIL)echo "■"   ;;
    blocks:WARN)echo "■"   ;;
    *)          echo "${1}" ;;
  esac
}

_color_for() {
  case "$1" in
    OK)   echo "${GREEN}"  ;;
    FAIL) echo "${RED}"    ;;
    WARN) echo "${YELLOW}" ;;
    *)    echo "${GREY}"   ;;
  esac
}

# --------------------------------------------------------------------
# Core checks
# --------------------------------------------------------------------
init_check() {
  local unameOut; unameOut="$(uname -a)"
  case "${unameOut}" in
    *Microsoft*)  C_DBX_OS="WSL"     ;;
    *microsoft*)  C_DBX_OS="WSL2"    ;;
    Linux*)       C_DBX_OS="Linux"   ;;
    Darwin*)      C_DBX_OS="Mac"     ;;
    CYGWIN*)      C_DBX_OS="Cygwin"  ;;
    MINGW*|*Msys) C_DBX_OS="Windows" ;;
    *)            C_DBX_OS="Unknown" ;;
  esac

  # check that we are in repo root (or at least that the init script exists)
  if [ ! -f "$C_DBX_INIT_PATH_ROOT/$C_DBX_INIT_ENTRYPOINT" ]; then
    echo -e "${RED}${BOLD}ERROR${RESET}: Devbox init path missing. Run from repo root or fix C_DBX_INIT_PATH_ROOT."
    exit 1
  fi
}

# --------------------------------------------------------------------
# PATH sanity (npm global bin)
# --------------------------------------------------------------------
init_path_sanity() {
  if command -v npm >/dev/null 2>&1; then
    local npm_prefix npm_bin
    npm_prefix="$(npm config get prefix 2>/dev/null || true)"
    if [ -n "$npm_prefix" ] && [ -d "$npm_prefix/bin" ]; then
      npm_bin="${npm_prefix}/bin"
    fi
    if [ -z "${npm_bin:-}" ] && [ -d "${HOME}/.npm-global/bin" ]; then
      npm_bin="${HOME}/.npm-global/bin"
    fi
    if [ -n "${npm_bin:-}" ] && [[ ":$PATH:" != *":${npm_bin}:"* ]]; then
      export PATH="${npm_bin}:$PATH"
    fi
    record_ok "npm-path" "global bin on PATH"
  else
    record_warn "npm-path" "npm missing (Claude CLI install needs Node)"
  fi
}

# --------------------------------------------------------------------
# Claude CLI + key
# --------------------------------------------------------------------
init_claude() {
  export CLAUDE_MODEL="${CLAUDE_MODEL:-claude-3-7-sonnet}"

  if command -v claude >/dev/null 2>&1; then
    record_ok "claude-cli" "installed"
  else
    if command -v npm >/dev/null 2>&1; then
      npm config set prefix "${HOME}/.npm-global" >/dev/null 2>&1 || true
      mkdir -p "${HOME}/.npm-global/bin"
      export PATH="${HOME}/.npm-global/bin:$PATH"
      npm i -g @anthropic-ai/cli >/dev/null 2>&1 || true

      if command -v claude >/dev/null 2>&1; then
        record_ok "claude-cli" "installed"
      else
        record_fail "claude-cli" "install failed (npm i -g @anthropic-ai/cli)"
      fi
    else
      record_fail "claude-cli" "npm missing"
    fi
  fi

  if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
    record_ok "anthropic-key" "present"
  else
    record_warn "anthropic-key" "missing"
  fi
}

# --------------------------------------------------------------------
# Matrix renderer
# --------------------------------------------------------------------
print_matrix() {
  echo
  echo -e "${BOLD}Devbox — ${C_DBX_META_TEAM}/${C_DBX_META_TEAM_ID} v${C_DBX_META_VERSION}${RESET} [${C_DBX_OS}]"
  echo "──────────────────────────────────────────────────────────────"
  printf "%-14s  %-3s  %s\n" "Component" "St" "Notes"
  echo "──────────────────────────────────────────────────────────────"

  for key in npm-path claude-cli anthropic-key; do
    if [[ -n "${STATUS[$key]:-}" ]]; then
      local code note icon color
      IFS='|' read -r code note <<< "${STATUS[$key]}"
      icon="$(_icon_for "${code}")"
      color="$(_color_for "${code}")"
      if [ "${DBX_NO_COLOR}" = "1" ]; then
        printf "%-14s  %-3s  %s\n" "$key" "$icon" "$note"
      else
        printf "%-14s  %b%-3s%b  %s\n" "$key" "$color" "$icon" "$RESET" "$note"
      fi
    fi
  done

  echo "──────────────────────────────────────────────────────────────"
}

# --------------------------------------------------------------------
# Help (compact, Claude-focused)
# --------------------------------------------------------------------
print_help() {
  echo -e "${BLUE}${BOLD}Shortcuts (devbox run)<${RESET}"
  echo "  ai-chat         Claude chat (interactive)"
  echo "  ai-ask          Claude one-shot question"
  echo "  ai-chat-sys     Claude chat with CLAUDE.md as system prompt"
  echo "  ai-ask-file     Claude ask with attached file"
  echo
  echo "Adjust these names to match your devbox.json \"scripts\" if needed."
}

# --------------------------------------------------------------------
# Entry
# --------------------------------------------------------------------
init_check
init_path_sanity
init_claude

print_matrix
print_help
