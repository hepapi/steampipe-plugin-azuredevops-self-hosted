#!/usr/bin/env bash
# ============================================================================
# Apply and test the Azure DevOps branch policy coverage views.
#
# Usage:
#   ./scripts/apply_views.sh            # apply views + run all test queries
#   ./scripts/apply_views.sh apply      # only apply (create/replace) the views
#   ./scripts/apply_views.sh test       # only run the test queries
#
# Requires `steampipe` on PATH and a configured `azuredevops` connection.
# ============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VIEWS_DIR="$SCRIPT_DIR/../docs/views"

# Views must be applied in dependency order.
VIEWS=(
  azuredevopssh_v_branch_policy_scope
  azuredevopssh_v_branch_policy_kind
  azuredevopssh_v_branch_policy_readable
  azuredevopssh_v_branch_coverage
  azuredevopssh_v_repository_policy_summary
)

apply_views() {
  echo "==> Applying views..."
  for v in "${VIEWS[@]}"; do
    local file="$VIEWS_DIR/$v.sql"
    if [[ ! -f "$file" ]]; then
      echo "ERROR: missing $file" >&2
      exit 1
    fi
    echo "    - $v"
    steampipe query "$(cat "$file")"
  done
  echo "==> Views applied."
}

run_tests() {
  echo ""
  echo "==> Running test queries..."

  echo ""
  echo "--- Readable policy list (names + links) ---"
  steampipe query "select project_name, repository_name, policy_type_display_name, policy_kind, scope_level, branch, policy_url from azuredevopssh_v_branch_policy_readable order by project_name, repository_name, branch"

  echo ""
  echo "--- Branch coverage (per repo+branch) ---"
  steampipe query "select project_name, repository_name, branch, has_branch_protection, has_status_check, repository_web_url from azuredevopssh_v_branch_coverage order by project_name, repository_name, branch"

  echo ""
  echo "--- Repository policy summary (per repo) ---"
  steampipe query "select project_name, repository_name, has_any_policy, has_repo_policy, has_project_level_policy, branch_protection_policy_count, status_check_policy_count, repository_web_url from azuredevopssh_v_repository_policy_summary order by has_any_policy, project_name, repository_name"

  echo ""
  echo "--- Repos with NO policies ---"
  steampipe query "select project_name, repository_name, repository_web_url from azuredevopssh_v_repository_policy_summary where not has_any_policy order by project_name, repository_name"

  echo ""
  echo "--- Repos missing status checks on default branch ---"
  steampipe query "select project_name, repository_name, repository_web_url from azuredevopssh_v_branch_coverage where not has_status_check and branch in ('master', 'main') order by project_name, repository_name"

  echo ""
  echo "==> Tests complete."
}

case "${1:-}" in
  apply) apply_views ;;
  test)  run_tests ;;
  *)     apply_views; run_tests ;;
esac
