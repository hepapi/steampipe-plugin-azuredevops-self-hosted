#!/usr/bin/env bash
# ============================================================================
# Smoke-test the Stage 3 (Build & Release) tables of the azuredevopssh plugin.
#
# Like scripts/test_stage2.sh, this runs plain `steampipe query` statements
# against the already-installed, running azuredevops connection. It does NOT
# build a custom binary or use an isolated install dir.
#
# Usage:
#   ./scripts/test_stage3.sh                         # run every Stage 3 query
#   ./scripts/test_stage3.sh --list                  # list + verify Stage 3 tables respond
#   ./scripts/test_stage3.sh release_definition      # run only one query set
#
# Requires `steampipe` on PATH and a configured `azuredevops` connection.
#
# build_timeline, build_log, and build_artifact need a build. pipeline_run
# needs a pipeline. pipeline_check needs a service connection.
# pipeline_approval needs an approval id; Azure DevOps Server 2022 rejects
# a list that does not pass one. Those queries look the ids up first.
# ============================================================================
set -euo pipefail

# A smoke query per Stage 3 table that can be listed directly. Each entry is
# "<label>|<sql>".
build_queries() {
  cat <<'EOF'
release_definition|select project_id, id, name, path from azuredevopssh_release_definition limit 5;
release_environment|select project_id, release_id, id, name, status from azuredevopssh_release_environment limit 5;
release_approval|select project_id, id, status, approval_type, release_id from azuredevopssh_release_approval limit 5;
EOF
}

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
  local logfile="${TMPDIR:-/tmp}/stage3_$$.log"
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

lookup_build() {
  local row
  row="$(query_row "select id, project_id from azuredevopssh_build limit 1;" || true)"
  if [[ -z "$row" ]]; then
    return 1
  fi
  IFS=',' read -r BUILD_ID BUILD_PROJECT_ID <<<"$row"
  BUILD_PROJECT_ID="${BUILD_PROJECT_ID// /}"
}

run_build_child() {
  local label="$1"
  if ! lookup_build; then
    echo ""
    echo "--- $label ---"
    bad "$label (no build to read)"
    return 1
  fi
  case "$label" in
    build_timeline)
      run_sql "$label" "select id, name, type, state, result from azuredevopssh_build_timeline where project_id = '${BUILD_PROJECT_ID}' and build_id = ${BUILD_ID} limit 5;"
      ;;
    build_log)
      run_sql "$label" "select id, type, line_count, url from azuredevopssh_build_log where project_id = '${BUILD_PROJECT_ID}' and build_id = ${BUILD_ID} limit 5;"
      ;;
    build_artifact)
      run_sql "$label" "select id, name, source from azuredevopssh_build_artifact where project_id = '${BUILD_PROJECT_ID}' and build_id = ${BUILD_ID} limit 5;"
      ;;
  esac
}

run_pipeline_run() {
  local row pipeline_id project_id
  row="$(query_row "select id, project_id from azuredevopssh_pipeline limit 1;" || true)"
  if [[ -z "$row" ]]; then
    echo ""
    echo "--- pipeline_run ---"
    bad "pipeline_run (no pipeline to read)"
    return 1
  fi
  IFS=',' read -r pipeline_id project_id <<<"$row"
  project_id="${project_id// /}"
  run_sql "pipeline_run" "select id, name, state, result from azuredevopssh_pipeline_run where project_id = '${project_id}' and pipeline_id = ${pipeline_id} limit 5;"
}

run_pipeline_approval() {
  # Server 2022 requires approvalIds. A Checkpoint.Approval timeline record
  # id is that approval id. When the sampled build has none, a Checkpoint
  # record id still exercises the call; a missing approval is an empty result.
  if ! lookup_build; then
    echo ""
    echo "--- pipeline_approval ---"
    bad "pipeline_approval (no build to read an approval id from)"
    return 1
  fi
  local approval_id
  approval_id="$(query_row "select id from azuredevopssh_build_timeline where project_id = '${BUILD_PROJECT_ID}' and build_id = ${BUILD_ID} and type = 'Checkpoint.Approval' limit 1;" || true)"
  approval_id="${approval_id// /}"
  if [[ -z "$approval_id" ]]; then
    approval_id="$(query_row "select id from azuredevopssh_build_timeline where project_id = '${BUILD_PROJECT_ID}' and build_id = ${BUILD_ID} and type = 'Checkpoint' limit 1;" || true)"
    approval_id="${approval_id// /}"
  fi
  if [[ -z "$approval_id" ]]; then
    echo ""
    echo "--- pipeline_approval ---"
    bad "pipeline_approval (no timeline record to use as an approval id)"
    return 1
  fi
  run_sql "pipeline_approval" "select id, status, min_required_approvers from azuredevopssh_pipeline_approval where project_id = '${BUILD_PROJECT_ID}' and id = '${approval_id}' limit 5;"
}

run_pipeline_check() {
  local row resource_id project_id
  row="$(query_row "select id, project_id from azuredevopssh_serviceendpoint limit 1;" || true)"
  if [[ -z "$row" ]]; then
    echo ""
    echo "--- pipeline_check ---"
    bad "pipeline_check (no service endpoint to use as a resource)"
    return 1
  fi
  IFS=',' read -r resource_id project_id <<<"$row"
  project_id="${project_id// /}"
  run_sql "pipeline_check" "select id, type_name, resource_name, timeout from azuredevopssh_pipeline_check where project_id = '${project_id}' and resource_type = 'endpoint' and resource_id = '${resource_id}' limit 5;"
}

list_tables() {
  echo ""
  echo "==> Stage 3 tables registered by the fork"
  echo "    Tables that require ids are checked with the lookup queries in a full run."
  local names=(
    azuredevopssh_release_definition
    azuredevopssh_release_environment
    azuredevopssh_release_approval
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
}

run_queries() {
  local target="$1"
  local pass=0 fail=0
  local label sql logfile
  local failed_labels=()
  logfile="${TMPDIR:-/tmp}/stage3_$$.log"
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

  local build_label
  for build_label in build_timeline build_log build_artifact; do
    if [[ "$target" == "all" || "$target" == "$build_label" ]]; then
      if run_build_child "$build_label"; then
        pass=$((pass+1))
      else
        failed_labels+=("$build_label")
        fail=$((fail+1))
      fi
    fi
  done

  if [[ "$target" == "all" || "$target" == "pipeline_approval" ]]; then
    if run_pipeline_approval; then
      pass=$((pass+1))
    else
      failed_labels+=("pipeline_approval")
      fail=$((fail+1))
    fi
  fi

  if [[ "$target" == "all" || "$target" == "pipeline_run" ]]; then
    if run_pipeline_run; then
      pass=$((pass+1))
    else
      failed_labels+=("pipeline_run")
      fail=$((fail+1))
    fi
  fi

  if [[ "$target" == "all" || "$target" == "pipeline_check" ]]; then
    if run_pipeline_check; then
      pass=$((pass+1))
    else
      failed_labels+=("pipeline_check")
      fail=$((fail+1))
    fi
  fi

  echo ""
  echo "=========================================================="
  echo " Stage 3 result: ${pass} passed, ${fail} failed"
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
    all|release_definition|release_environment|release_approval|pipeline_approval|build_timeline|build_log|build_artifact|pipeline_run|pipeline_check) ;;
    *) echo "Unknown arg: $mode" >&2; return 1 ;;
  esac

  run_queries "$mode"
}

main "$@"
