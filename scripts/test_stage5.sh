#!/usr/bin/env bash
# ============================================================================
# Smoke-test the Stage 5 (Test & Quality) tables of the azuredevopssh plugin.
#
# Like scripts/test_stage4.sh, this runs plain `steampipe query` statements
# against the already-installed, running azuredevops connection. It does NOT
# build a custom binary or use an isolated install dir.
#
# Usage:
#   ./scripts/test_stage5.sh                         # run every Stage 5 query
#   ./scripts/test_stage5.sh --list                  # list + verify Stage 5 tables respond
#   ./scripts/test_stage5.sh test_plan               # run only one query set
#
# Requires `steampipe` on PATH and a configured `azuredevops` connection.
#
# test_plan, test_suite, test_case, and test_point use the Test Plans API and
# may fail when that license is missing. test_run and test_result use the test
# run API. Child queries are skipped when the parent row does not exist.
# ============================================================================
set -euo pipefail

ok()  { printf '\033[1;32m  ok\033[0m   %s\n' "$*"; }
bad() { printf '\033[1;31m  FAIL\033[0m %s\n' "$*"; }
skip() { printf '\033[1;33m  skip\033[0m %s\n' "$*"; }

# First row of a steampipe CSV result, header stripped, quotes removed.
query_row() {
  local sql="$1"
  steampipe query --output csv --header=false "$sql" 2>/dev/null | head -n 1 | tr -d '"' | tr -d '\r'
}

run_sql() {
  local label="$1"
  local sql="$2"
  local logfile="${TMPDIR:-/tmp}/stage5_$$.log"
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

note_skip() {
  local label="$1"
  local reason="$2"
  echo ""
  echo "--- $label ---"
  skip "$label ($reason)"
  return 2
}

run_test_plan() {
  run_sql "test_plan" "select id, project_id, name, state from azuredevopssh_test_plan limit 5;"
}

run_test_run() {
  run_sql "test_run" "select id, project_id, name, state, is_automated, total_tests, passed_tests from azuredevopssh_test_run limit 5;"
}

lookup_plan() {
  local row
  row="$(query_row "select id, project_id from azuredevopssh_test_plan limit 1;" || true)"
  if [[ -z "$row" ]]; then
    return 1
  fi
  IFS=',' read -r PLAN_ID PROJECT_ID <<<"$row"
  PLAN_ID="${PLAN_ID// /}"
  PROJECT_ID="${PROJECT_ID// /}"
  [[ -n "$PLAN_ID" && -n "$PROJECT_ID" ]]
}

lookup_suite() {
  if ! lookup_plan; then
    return 1
  fi
  local row
  row="$(query_row "select id from azuredevopssh_test_suite where project_id = '${PROJECT_ID}' and plan_id = ${PLAN_ID} limit 1;" || true)"
  SUITE_ID="${row// /}"
  [[ -n "$SUITE_ID" ]]
}

lookup_run() {
  local row
  row="$(query_row "select id, project_id from azuredevopssh_test_run limit 1;" || true)"
  if [[ -z "$row" ]]; then
    return 1
  fi
  IFS=',' read -r RUN_ID RUN_PROJECT_ID <<<"$row"
  RUN_ID="${RUN_ID// /}"
  RUN_PROJECT_ID="${RUN_PROJECT_ID// /}"
  [[ -n "$RUN_ID" && -n "$RUN_PROJECT_ID" ]]
}

run_test_suite() {
  if ! lookup_plan; then
    note_skip "test_suite" "no test plan to read"
    return 2
  fi
  run_sql "test_suite" "select id, name, suite_type, parent_suite_id from azuredevopssh_test_suite where project_id = '${PROJECT_ID}' and plan_id = ${PLAN_ID} limit 5;"
}

run_test_case() {
  if ! lookup_suite; then
    note_skip "test_case" "no test suite to read"
    return 2
  fi
  run_sql "test_case" "select id, name, \"order\" from azuredevopssh_test_case where project_id = '${PROJECT_ID}' and plan_id = ${PLAN_ID} and suite_id = ${SUITE_ID} limit 5;"
}

run_test_point() {
  if ! lookup_suite; then
    note_skip "test_point" "no test suite to read"
    return 2
  fi
  run_sql "test_point" "select id, test_case_name, configuration_name, outcome from azuredevopssh_test_point where project_id = '${PROJECT_ID}' and plan_id = ${PLAN_ID} and suite_id = ${SUITE_ID} limit 5;"
}

run_test_result() {
  if ! lookup_run; then
    note_skip "test_result" "no test run to read"
    return 2
  fi
  run_sql "test_result" "select id, test_case_title, outcome, state from azuredevopssh_test_result where project_id = '${RUN_PROJECT_ID}' and run_id = ${RUN_ID} limit 5;"
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

list_tables() {
  echo ""
  echo "==> Stage 5 tables registered by the fork"
  echo "    test_plan and test_run can be listed directly. The others need parent ids."
  local names=(azuredevopssh_test_plan azuredevopssh_test_run)
  local t
  for t in "${names[@]}"; do
    echo ""
    echo "--- $t (responds?) ---"
    if steampipe query "select * from $t limit 1" >/dev/null 2>&1; then
      ok "$t"
    else
      bad "$t"
    fi
  done
}

run_queries() {
  local target="$1"
  pass=0
  fail=0
  skip_count=0
  failed_labels=()

  if [[ "$target" == "all" || "$target" == "test_plan" ]]; then
    record "test_plan" run_test_plan
  fi
  if [[ "$target" == "all" || "$target" == "test_suite" ]]; then
    record "test_suite" run_test_suite
  fi
  if [[ "$target" == "all" || "$target" == "test_case" ]]; then
    record "test_case" run_test_case
  fi
  if [[ "$target" == "all" || "$target" == "test_point" ]]; then
    record "test_point" run_test_point
  fi
  if [[ "$target" == "all" || "$target" == "test_run" ]]; then
    record "test_run" run_test_run
  fi
  if [[ "$target" == "all" || "$target" == "test_result" ]]; then
    record "test_result" run_test_result
  fi

  echo ""
  echo "=========================================================="
  echo " Stage 5 result: ${pass} passed, ${fail} failed, ${skip_count} skipped"
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
    all|test_plan|test_suite|test_case|test_point|test_run|test_result) ;;
    *) echo "Unknown arg: $mode" >&2; return 1 ;;
  esac

  run_queries "$mode"
}

main "$@"
