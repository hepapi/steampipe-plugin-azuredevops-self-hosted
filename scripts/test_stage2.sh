#!/usr/bin/env bash
# ============================================================================
# Smoke-test the Stage 2 (Policy & Security) tables of the azuredevopssh plugin.
#
# Like scripts/test_stage1.sh, this runs plain `steampipe query` statements
# against the already-installed, running azuredevops connection. It does NOT
# build a custom binary or use an isolated install dir.
#
# Usage:
#   ./scripts/test_stage2.sh                         # run every Stage 2 query
#   ./scripts/test_stage2.sh --list                  # list + verify Stage 2 tables respond
#   ./scripts/test_stage2.sh policy_type             # run only one query set
#
# Requires `steampipe` on PATH and a configured `azuredevops` connection.
#
# policy_evaluation needs a pull request artifact, and permission needs a
# security namespace. Those two queries look up the ids first, then query
# with a small limit.
# ============================================================================
set -euo pipefail

# A smoke query per Stage 2 table that can be listed directly. Each entry is
# "<label>|<sql>".
build_queries() {
  cat <<'EOF'
policy_type|select project_id, id, display_name from azuredevopssh_policy_type order by display_name limit 5;
security_namespace|select namespace_id, name, display_name from azuredevopssh_security_namespace order by name limit 20;
servicehook_subscription|select id, publisher_id, event_type, consumer_id, status from azuredevopssh_servicehook_subscription limit 10;
EOF
}

ok()  { printf '\033[1;32m  ok\033[0m   %s\n' "$*"; }
bad() { printf '\033[1;31m  FAIL\033[0m %s\n' "$*"; }

# Run one SQL statement. Prints the result on success.
run_sql() {
  local label="$1"
  local sql="$2"
  local logfile="${TMPDIR:-/tmp}/stage2_$$.log"
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

# First cell of a one-column steampipe result, header stripped.
query_value() {
  local sql="$1"
  steampipe query --output csv --header=false "$sql" 2>/dev/null | head -n 1 | tr -d '"' | tr -d '[:space:]'
}

run_permission() {
  local namespace_id
  namespace_id="$(query_value "select namespace_id from azuredevopssh_security_namespace where name = 'Git Repositories' limit 1;" || true)"
  if [[ -z "$namespace_id" ]]; then
    echo ""
    echo "--- permission ---"
    bad "permission (no Git Repositories security namespace)"
    return 1
  fi
  run_sql "permission" "select namespace_name, token, descriptor, allow, deny from azuredevopssh_permission where namespace_id = '${namespace_id}' limit 5;"
}

run_policy_evaluation() {
  # artifact_id is often null on the pull request row. Build the Code Review
  # artifact id the policy API expects when the column is empty.
  local row project_id pull_request_id artifact_id
  row="$(steampipe query --output csv --header=false "select project_id, pull_request_id, coalesce(nullif(artifact_id, ''), 'vstfs:///CodeReview/CodeReviewId/' || project_id || '/' || pull_request_id::text) from azuredevopssh_git_pull_request limit 1;" 2>/dev/null | head -n 1 || true)"
  if [[ -z "$row" ]]; then
    echo ""
    echo "--- policy_evaluation ---"
    bad "policy_evaluation (no pull request to evaluate)"
    return 1
  fi
  IFS=',' read -r project_id pull_request_id artifact_id <<<"$row"
  project_id="${project_id//\"/}"
  pull_request_id="${pull_request_id//\"/}"
  artifact_id="${artifact_id//\"/}"
  artifact_id="${artifact_id//$'\r'/}"
  run_sql "policy_evaluation" "select id, status, policy_type_display_name, is_enabled, is_blocking from azuredevopssh_policy_evaluation where project_id = '${project_id}' and artifact_id = '${artifact_id}' limit 10;"
}

list_tables() {
  echo ""
  echo "==> Stage 2 tables registered by the fork"
  local names=(
    azuredevopssh_policy_type
    azuredevopssh_security_namespace
    azuredevopssh_servicehook_subscription
  )
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
  run_permission || true
  run_policy_evaluation || true
}

run_queries() {
  local target="$1"
  local pass=0 fail=0
  local label sql logfile
  local failed_labels=()
  logfile="${TMPDIR:-/tmp}/stage2_$$.log"
  while IFS='|' read -r label sql; do
    [[ -z "$label" ]] && continue
    if [[ "$target" == "all" || "$target" == "$label" ]]; then
      echo ""
      echo "--- $label ---"
      if steampipe query "$sql" >"$logfile" 2>&1; then
        cat "$logfile"
        ok "$label"; pass=$((pass+1))
      else
        bad "$label"
        echo "--- stderr (last lines) ---"
        tail -n 15 "$logfile" || true
        failed_labels+=("$label")
        fail=$((fail+1))
      fi
    fi
  done < <(build_queries)
  rm -f "$logfile"

  if [[ "$target" == "all" || "$target" == "permission" ]]; then
    if run_permission; then
      pass=$((pass+1))
    else
      failed_labels+=("permission")
      fail=$((fail+1))
    fi
  fi

  if [[ "$target" == "all" || "$target" == "policy_evaluation" ]]; then
    if run_policy_evaluation; then
      pass=$((pass+1))
    else
      failed_labels+=("policy_evaluation")
      fail=$((fail+1))
    fi
  fi

  echo ""
  echo "=========================================================="
  echo " Stage 2 result: ${pass} passed, ${fail} failed"
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
    all|policy_type|security_namespace|servicehook_subscription|permission|policy_evaluation) ;;
    *) echo "Unknown arg: $mode" >&2; return 1 ;;
  esac

  run_queries "$mode"
}

main "$@"
