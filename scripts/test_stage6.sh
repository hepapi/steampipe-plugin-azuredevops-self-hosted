#!/usr/bin/env bash
# ============================================================================
# Smoke-test the Stage 6 (Identity, Permissions & Entitlements) tables.
#
# Like scripts/test_stage5.sh, this runs plain `steampipe query` statements
# against the already-installed, running azuredevops connection. It does NOT
# build a custom binary or use an isolated install dir.
#
# Usage:
#   ./scripts/test_stage6.sh
#   ./scripts/test_stage6.sh identity
#
# identity and group_membership are expected to return rows on Azure DevOps
# Server. member_entitlement and user_entitlement are skipped when this server
# does not register those APIs.
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
  local logfile="${TMPDIR:-/tmp}/stage6_$$.log"
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

# Entitlement APIs are not registered on this Azure DevOps Server collection.
run_sql_or_unavailable() {
  local label="$1"
  local sql="$2"
  local logfile="${TMPDIR:-/tmp}/stage6_$$.log"
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

identity_sql="select id, descriptor, subject_descriptor, provider_display_name, is_container from azuredevops_identity where search_filter = 'LocalGroupName' and filter_value = 'Project Collection Administrators' limit 5;"

run_identity() {
  run_sql "identity" "$identity_sql"
}

run_identity_by_id() {
  local row
  row="$(query_row "select id from azuredevops_identity where search_filter = 'LocalGroupName' and filter_value = 'Project Collection Administrators' limit 1;" || true)"
  row="${row// /}"
  if [[ -z "$row" ]]; then
    note_skip "identity_by_id" "no identity matched filter_value"
    return 2
  fi
  run_sql "identity_by_id" "select id, provider_display_name, is_active, is_container from azuredevops_identity where id = '${row}';"
}

run_group_membership() {
  local row subject descriptor
  row="$(query_row "select subject_descriptor, descriptor from azuredevops_identity where search_filter = 'LocalGroupName' and filter_value = 'Project Collection Administrators' and is_container = true limit 1;" || true)"
  if [[ -z "$row" ]]; then
    note_skip "group_membership" "no group identity to read"
    return 2
  fi
  IFS=',' read -r subject descriptor <<<"$row"
  subject="${subject// /}"
  descriptor="${descriptor// /}"
  if [[ -z "$subject" ]]; then
    subject="$descriptor"
  fi
  if [[ -z "$subject" ]]; then
    note_skip "group_membership" "identity has no descriptor"
    return 2
  fi
  run_sql "group_membership" "select container_descriptor, member_descriptor, direction from azuredevops_group_membership where subject_descriptor = '${subject}' limit 5;"
}

run_member_entitlement() {
  run_sql_or_unavailable "member_entitlement" "select id, display_name, account_license_type, license_display_name from azuredevops_member_entitlement limit 5;"
}

run_user_entitlement() {
  run_sql_or_unavailable "user_entitlement" "select id, display_name, principal_name, account_license_type, license_display_name from azuredevops_user_entitlement limit 5;"
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

  if [[ "$target" == "all" || "$target" == "identity" ]]; then
    record "identity" run_identity
  fi
  if [[ "$target" == "all" || "$target" == "identity_by_id" ]]; then
    record "identity_by_id" run_identity_by_id
  fi
  if [[ "$target" == "all" || "$target" == "group_membership" ]]; then
    record "group_membership" run_group_membership
  fi
  if [[ "$target" == "all" || "$target" == "member_entitlement" ]]; then
    record "member_entitlement" run_member_entitlement
  fi
  if [[ "$target" == "all" || "$target" == "user_entitlement" ]]; then
    record "user_entitlement" run_user_entitlement
  fi

  echo ""
  echo "=========================================================="
  echo " Stage 6 result: ${pass} passed, ${fail} failed, ${skip_count} skipped"
  if [[ "$fail" -gt 0 ]]; then
    echo " Failed: ${failed_labels[*]}"
  fi
  echo "=========================================================="
  [[ "$fail" -eq 0 ]]
}

main() {
  local mode="${1:-all}"
  case "$mode" in
    all|identity|identity_by_id|group_membership|member_entitlement|user_entitlement) ;;
    *) echo "Unknown arg: $mode" >&2; return 1 ;;
  esac
  run_queries "$mode"
}

main "$@"
