create or replace view azuredevopssh_v_branch_coverage as
with branch_protection as (
  select repository_id, ref_name, count(*) as policy_count
  from azuredevopssh_v_branch_policy_kind
  where policy_kind = 'branch_protection' and repository_id is not null and ref_name is not null
  group by repository_id, ref_name
),
status_checks as (
  select repository_id, ref_name, count(*) as policy_count
  from azuredevopssh_v_branch_policy_kind
  where policy_kind = 'status_check' and repository_id is not null and ref_name is not null
  group by repository_id, ref_name
)
select
  r.id                                 as repository_id,
  r.name                               as repository_name,
  r.web_url                            as repository_web_url,
  r.project_id                         as project_id,
  prj.name                             as project_name,
  b.name                               as branch_ref,
  replace(b.name, 'refs/heads/', '')   as branch,
  coalesce(bp.policy_count, 0) > 0     as has_branch_protection,
  coalesce(bp.policy_count, 0)         as branch_protection_policy_count,
  coalesce(sc.policy_count, 0) > 0     as has_status_check,
  coalesce(sc.policy_count, 0)         as status_check_policy_count
from
  azuredevopssh_git_repository r
  join azuredevopssh_git_repository_branch b on b.repository_id = r.id
  left join azuredevopssh_project prj on prj.id = r.project_id
  left join branch_protection bp
    on bp.repository_id = r.id and bp.ref_name = b.name
  left join status_checks sc
    on sc.repository_id = r.id and sc.ref_name = b.name;
