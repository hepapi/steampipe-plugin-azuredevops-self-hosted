#!/usr/bin/env bash
# ============================================================================
# Smoke-test the Stage 4 (Work Item Tracking) tables of the azuredevopssh plugin.
#
# Like scripts/test_stage3.sh, this runs plain `steampipe query` statements
# against the already-installed, running azuredevops connection. It does NOT
# build a custom binary or use an isolated install dir.
#
# Usage:
#   ./scripts/test_stage4.sh                         # run every Stage 4 query
#   ./scripts/test_stage4.sh --list                  # list + verify Stage 4 tables respond
#   ./scripts/test_stage4.sh work_item               # run only one query set
#
# Requires `steampipe` on PATH and a configured `azuredevops` connection.
#
# azuredevopssh_work_item is the only Stage 4 table implemented so far.
# Queries are scoped to one project. work_item_by_id is skipped when that
# project has no work items.
# ============================================================================
set -euo pipefail

ok()  { printf '\033[1;32m  ok\033[0m   %s\n' "$*"; }
bad() { printf '\033[1;31m  FAIL\033[0m %s\n' "$*"; }

# First row of a steampipe CSV result, header stripped, quotes removed.
query_row() {
  local sql="$1"
  steampipe query --output csv --header=false "$sql" 2>/dev/null | head -n 1 | tr -d '"' | tr -d '\r'
}

run_sql() {
  local label="$1"
  local sql="$2"
  local logfile="${TMPDIR:-/tmp}/stage4_$$.log"
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

lookup_project() {
  local row
  row="$(query_row "select id from azuredevopssh_project limit 1;" || true)"
  if [[ -z "$row" ]]; then
    return 1
  fi
  PROJECT_ID="${row// /}"
}

run_work_item() {
  if ! lookup_project; then
    echo ""
    echo "--- work_item ---"
    bad "work_item (no project to read)"
    return 1
  fi
  run_sql "work_item" "select id, title, work_item_type, state, assigned_to_name, changed_date from azuredevopssh_work_item where project_id = '${PROJECT_ID}' limit 5;"
}

skip() { printf '\033[1;33m  skip\033[0m %s\n' "$*"; }

run_work_item_by_id() {
  if ! lookup_project; then
    echo ""
    echo "--- work_item_by_id ---"
    bad "work_item_by_id (no project to read)"
    return 1
  fi
  local work_item_id
  work_item_id="$(query_row "select id from azuredevopssh_work_item where project_id = '${PROJECT_ID}' limit 1;" || true)"
  work_item_id="${work_item_id// /}"
  if [[ -z "$work_item_id" ]]; then
    echo ""
    echo "--- work_item_by_id ---"
    skip "work_item_by_id (no work items in project ${PROJECT_ID})"
    return 2
  fi
  run_sql "work_item_by_id" "select id, rev, project_id, title, work_item_type, state, url from azuredevopssh_work_item where project_id = '${PROJECT_ID}' and id = ${work_item_id};"
}

list_tables() {
  echo ""
  echo "==> Stage 4 tables registered by the fork"
  echo "    Queries are scoped to one project so they do not scan the collection."
  if ! lookup_project; then
    echo ""
    echo "--- azuredevopssh_work_item (responds?) ---"
    bad "azuredevopssh_work_item (no project to read)"
    return 1
  fi
  echo ""
  echo "--- azuredevopssh_work_item (responds?) ---"
  if steampipe query "select * from azuredevopssh_work_item where project_id = '${PROJECT_ID}' limit 1" >/dev/null 2>&1; then
    ok "azuredevopssh_work_item"
  else
    bad "azuredevopssh_work_item"
    return 1
  fi
}

run_queries() {
  local target="$1"
  local pass=0 fail=0 skip=0
  local -a failed_labels=()

  if [[ "$target" == "all" || "$target" == "work_item" ]]; then
    if run_work_item; then
      pass=$((pass+1))
    else
      failed_labels+=("work_item")
      fail=$((fail+1))
    fi
  fi

  if [[ "$target" == "all" || "$target" == "work_item_by_id" ]]; then
    if run_work_item_by_id; then
      pass=$((pass+1))
    else
      local status=$?
      if [[ "$status" -eq 2 ]]; then
        skip=$((skip+1))
      else
        failed_labels+=("work_item_by_id")
        fail=$((fail+1))
      fi
    fi
  fi

  echo ""
  echo "=========================================================="
  echo " Stage 4 result: ${pass} passed, ${fail} failed, ${skip} skipped"
  if [[ "$fail" -gt 0 ]]; then
    echo " Failed: ${failed_labels[*]}"
  fi
  echo "=========================================================="
  [[ "$fail" -eq 0 ]]
}

main() {
  local mode="${1:-all}"

  if [[ "$mode" == "--list" ]]; then
    list_tables
    return 0
  fi

  case "$mode" in
    all|work_item|work_item_by_id) ;;
    *) echo "Unknown arg: $mode" >&2; return 1 ;;
  esac

  run_queries "$mode"
}

main "$@"
