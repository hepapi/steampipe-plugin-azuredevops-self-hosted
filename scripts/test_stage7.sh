#!/usr/bin/env bash
# ============================================================================
# Smoke-test the Stage 7 (Artifacts / Packages) tables.
#
# Like scripts/test_stage6.sh, this runs plain `steampipe query` statements
# against the already-installed, running azuredevops connection. It does NOT
# build a custom binary or use an isolated install dir.
#
# Usage:
#   ./scripts/test_stage7.sh
#   ./scripts/test_stage7.sh feed
#
# package and package_version are skipped when the collection has no package
# to read. An empty package list is still a pass.
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
  local logfile="${TMPDIR:-/tmp}/stage7_$$.log"
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

run_feed() {
  run_sql "feed" "select id, name, project_id, upstream_enabled from azuredevops_feed limit 5;"
}

run_feed_by_id() {
  local row
  row="$(query_row "select id from azuredevops_feed limit 1;" || true)"
  row="${row// /}"
  if [[ -z "$row" ]]; then
    note_skip "feed_by_id" "no feed to read"
    return 2
  fi
  run_sql "feed_by_id" "select id, name, fully_qualified_name, upstream_enabled from azuredevops_feed where id = '${row}';"
}

lookup_feed() {
  local row
  row="$(query_row "select id from azuredevops_feed limit 1;" || true)"
  FEED_ID="${row// /}"
  [[ -n "$FEED_ID" ]]
}

run_package() {
  if ! lookup_feed; then
    note_skip "package" "no feed to read"
    return 2
  fi
  run_sql "package" "select id, name, protocol_type from azuredevops_package where feed_id = '${FEED_ID}' limit 5;"
}

run_package_version() {
  if ! lookup_feed; then
    note_skip "package_version" "no feed to read"
    return 2
  fi
  local package_id
  package_id="$(query_row "select id from azuredevops_package where feed_id = '${FEED_ID}' limit 1;" || true)"
  package_id="${package_id// /}"
  if [[ -z "$package_id" ]]; then
    note_skip "package_version" "no package in feed ${FEED_ID}"
    return 2
  fi
  run_sql "package_version" "select id, version, is_latest, is_deleted, publish_date from azuredevops_package_version where feed_id = '${FEED_ID}' and package_id = '${package_id}' limit 5;"
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

  if [[ "$target" == "all" || "$target" == "feed" ]]; then
    record "feed" run_feed
  fi
  if [[ "$target" == "all" || "$target" == "feed_by_id" ]]; then
    record "feed_by_id" run_feed_by_id
  fi
  if [[ "$target" == "all" || "$target" == "package" ]]; then
    record "package" run_package
  fi
  if [[ "$target" == "all" || "$target" == "package_version" ]]; then
    record "package_version" run_package_version
  fi

  echo ""
  echo "=========================================================="
  echo " Stage 7 result: ${pass} passed, ${fail} failed, ${skip_count} skipped"
  if [[ "$fail" -gt 0 ]]; then
    echo " Failed: ${failed_labels[*]}"
  fi
  echo "=========================================================="
  [[ "$fail" -eq 0 ]]
}

main() {
  local mode="${1:-all}"
  case "$mode" in
    all|feed|feed_by_id|package|package_version) ;;
    *) echo "Unknown arg: $mode" >&2; return 1 ;;
  esac
  run_queries "$mode"
}

main "$@"
