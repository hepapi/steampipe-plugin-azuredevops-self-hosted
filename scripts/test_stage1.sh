#!/usr/bin/env bash
# ============================================================================
# Smoke-test the Stage 1 (Git & Code) tables of the azuredevopssh plugin.
#
# Like scripts/apply_views.sh, this runs plain `steampipe query` statements
# against the already-installed, running azuredevops connection. It does NOT
# build a custom binary or use an isolated install dir.
#
# Usage:
#   ./scripts/test_stage1.sh                     # run every Stage 1 query
#   ./scripts/test_stage1.sh --list              # list + verify Stage 1 tables respond
#   ./scripts/test_stage1.sh pull_request        # run only one query set
#
# Requires `steampipe` on PATH and a configured `azuredevops` connection.
# ============================================================================
set -euo pipefail

# A smoke query per Stage 1 area. Each entry is "<label>|<sql>" where <sql>
# references the table and returns a small row set.
build_queries() {
  cat <<'EOF'
pull_request|select pull_request_id, repository_id, status, title from azuredevopssh_git_pull_request order by pull_request_id limit 5;
pull_request_thread|select pull_request_id, repository_id, thread_id, status from azuredevopssh_git_pull_request_thread order by thread_id limit 5;
pull_request_comment|select pull_request_id, repository_id, thread_id, comment_type, content from azuredevopssh_git_pull_request_comment order by thread_id limit 5;
pull_request_reviewer|select pull_request_id, repository_id, reviewer_id, display_name, vote from azuredevopssh_git_pull_request_reviewer order by pull_request_id limit 5;
commit|select commit_id, repository_id, author_name, author_email from azuredevopssh_git_commit order by commit_id limit 5;
ref|select repository_id, project_id, name, object_id, is_locked from azuredevopssh_git_ref order by name limit 5;
tag|select repository_id, project_id, name, message from azuredevopssh_git_tag order by name limit 5;
repo_enriched|select name, project_id, project_name, default_branch, is_disabled from azuredevopssh_git_repository order by name limit 100;
EOF
}

ok()  { printf '\033[1;32m  ok\033[0m   %s\n' "$*"; }
bad() { printf '\033[1;31m  FAIL\033[0m %s\n' "$*"; }

list_tables() {
  echo ""
  echo "==> Stage 1 tables registered by the fork"
  local names=(
    azuredevopssh_git_pull_request
    azuredevopssh_git_pull_request_thread
    azuredevopssh_git_pull_request_comment
    azuredevopssh_git_pull_request_reviewer
    azuredevopssh_git_commit
    azuredevopssh_git_ref
    azuredevopssh_git_tag
  )
  local t
  for t in "${names[@]}"; do
    echo ""
    echo "--- $t (responds?) ---"
    # Fetch a single row so the table is exercised but never fully scanned.
    # This is especially important for azuredevopssh_git_commit: a plain
    # count(*) would paginate every commit across every repository.
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
  logfile="${TMPDIR:-/tmp}/stage1_$$.log"
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

  echo ""
  echo "=========================================================="
  echo " Stage 1 result: ${pass} passed, ${fail} failed"
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
    all|pull_request|pull_request_thread|pull_request_comment|pull_request_reviewer|commit|ref|tag|repo_enriched) ;;
    *) echo "Unknown arg: $mode" >&2; return 1 ;;
  esac

  run_queries "$mode"
}

main "$@"