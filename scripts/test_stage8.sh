#!/usr/bin/env bash
# ============================================================================
# Smoke-test the Stage 8 (Wiki, Search, Audit, Extensions, Agents) tables.
#
# Like scripts/test_stage7.sh, this runs plain `steampipe query` statements
# against the already-installed, running azuredevops connection. It does NOT
# build a custom binary or use an isolated install dir.
#
# Usage:
#   ./scripts/test_stage8.sh
#   ./scripts/test_stage8.sh wiki
#
# audit_log is skipped when this server does not expose the audit API.
# wiki_page and agent are skipped when the parent row does not exist.
# Variable values are not selected.
# ============================================================================
set -euo pipefail

ok()  { printf '\033[1;32m  ok\033[0m   %s\n' "$*"; }
bad() { printf '\033[1;31m  FAIL\033[0m %s\n' "$*"; }
skip() { printf '\033[1;33m  skip\033[0m %s\n' "$*"; }

query_row() {
  local sql="$1"
  steampipe query --output csv --header=false "$sql" 2>/dev/null | head -n 1 | tr -d '"' | tr -d '\r'
}

run_sql() {
  local label="$1"
  local sql="$2"
  local logfile="${TMPDIR:-/tmp}/stage8_$$.log"
  echo ""
  echo "--- $label ---"
  if steampipe query "$sql" >"$logfile" 2>&1; then
    cat "$logfile"
    ok "$label"
    rm -f "$logfile"
    return 0
  fi
  bad "$label"
  echo "--- stderr (last lines) ---"
  tail -n 15 "$logfile" || true
  rm -f "$logfile"
  return 1
}

run_sql_or_unavailable() {
  local label="$1"
  local sql="$2"
  local logfile="${TMPDIR:-/tmp}/stage8_$$.log"
  echo ""
  echo "--- $label ---"
  if steampipe query "$sql" >"$logfile" 2>&1; then
    cat "$logfile"
    ok "$label"
    rm -f "$logfile"
    return 0
  fi
  if grep -Eq 'is not registered|404|Page not found|could not be found' "$logfile"; then
    skip "$label (API is not registered on this server)"
    rm -f "$logfile"
    return 2
  fi
  bad "$label"
  echo "--- stderr (last lines) ---"
  tail -n 15 "$logfile" || true
  rm -f "$logfile"
  return 1
}

note_skip() {
  local label="$1"
  local reason="$2"
  echo ""
  echo "--- $label ---"
  skip "$label ($reason)"
  return 2
}

run_wiki() {
  run_sql "wiki" "select id, name, type, project_id, mapped_path from azuredevopssh_wiki limit 5;"
}

run_wiki_page() {
  local row wiki_id project_id
  row="$(query_row "select id, project_id from azuredevopssh_wiki limit 1;" || true)"
  if [[ -z "$row" ]]; then
    note_skip "wiki_page" "no wiki to read"
    return 2
  fi
  IFS=',' read -r wiki_id project_id <<<"$row"
  wiki_id="${wiki_id// /}"
  project_id="${project_id// /}"
  if [[ -z "$wiki_id" || -z "$project_id" ]]; then
    note_skip "wiki_page" "wiki has no project"
    return 2
  fi
  run_sql "wiki_page" "select id, path from azuredevopssh_wiki_page where project_id = '${project_id}' and wiki_id = '${wiki_id}' limit 5;"
}

run_code_search_result() {
  run_sql "code_search_result" "select file_name, path, project_name, repository_name from azuredevopssh_code_search_result where search_text = 'func' limit 3;"
}

run_audit_log() {
  run_sql_or_unavailable "audit_log" "select id, action_id, timestamp from azuredevopssh_audit_log limit 5;"
}

run_extension() {
  run_sql "extension" "select publisher_id, extension_id, extension_name, version from azuredevopssh_extension limit 5;"
}

run_agent_pool() {
  run_sql "agent_pool" "select id, name, pool_type, size, is_hosted from azuredevopssh_agent_pool limit 5;"
}

run_agent() {
  local pool_id
  pool_id="$(query_row "select id from azuredevopssh_agent_pool limit 1;" || true)"
  pool_id="${pool_id// /}"
  if [[ -z "$pool_id" ]]; then
    note_skip "agent" "no agent pool to read"
    return 2
  fi
  run_sql "agent" "select id, pool_id, name, version, status, enabled from azuredevopssh_agent where pool_id = ${pool_id} limit 5;"
}

run_variable_group() {
  run_sql "variable_group" "select id, project_id, name, type, is_shared from azuredevopssh_variable_group limit 5;"
}

record() {
  local label="$1"
  if "$2"; then
    pass=$((pass+1))
  else
    local status=$?
    if [[ "$status" -eq 2 ]]; then
      skip_count=$((skip_count+1))
    else
      failed_labels+=("$label")
      fail=$((fail+1))
    fi
  fi
}

run_queries() {
  local target="$1"
  pass=0
  fail=0
  skip_count=0
  failed_labels=()

  if [[ "$target" == "all" || "$target" == "wiki" ]]; then
    record "wiki" run_wiki
  fi
  if [[ "$target" == "all" || "$target" == "wiki_page" ]]; then
    record "wiki_page" run_wiki_page
  fi
  if [[ "$target" == "all" || "$target" == "code_search_result" ]]; then
    record "code_search_result" run_code_search_result
  fi
  if [[ "$target" == "all" || "$target" == "audit_log" ]]; then
    record "audit_log" run_audit_log
  fi
  if [[ "$target" == "all" || "$target" == "extension" ]]; then
    record "extension" run_extension
  fi
  if [[ "$target" == "all" || "$target" == "agent_pool" ]]; then
    record "agent_pool" run_agent_pool
  fi
  if [[ "$target" == "all" || "$target" == "agent" ]]; then
    record "agent" run_agent
  fi
  if [[ "$target" == "all" || "$target" == "variable_group" ]]; then
    record "variable_group" run_variable_group
  fi

  echo ""
  echo "=========================================================="
  echo " Stage 8 result: ${pass} passed, ${fail} failed, ${skip_count} skipped"
  if [[ "$fail" -gt 0 ]]; then
    echo " Failed: ${failed_labels[*]}"
  fi
  echo "=========================================================="
  [[ "$fail" -eq 0 ]]
}

main() {
  local mode="${1:-all}"
  case "$mode" in
    all|wiki|wiki_page|code_search_result|audit_log|extension|agent_pool|agent|variable_group) ;;
    *) echo "Unknown arg: $mode" >&2; return 1 ;;
  esac
  run_queries "$mode"
}

main "$@"
