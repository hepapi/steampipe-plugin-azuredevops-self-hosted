create or replace view azuredevops_v_repository_policy_summary as
with repo_policies as (
  -- policies that explicitly target a specific repository
  select
    repository_id,
    count(*)                                          as total_policy_count,
    count(*) filter (where policy_kind = 'branch_protection') as branch_protection_policy_count,
    count(*) filter (where policy_kind = 'status_check')      as status_check_policy_count
  from
    azuredevops_v_branch_policy_kind
  where
    repository_id is not null
  group by
    repository_id
),
project_policies as (
  -- project-level policies (no repository_id) that apply to all repos in a project
  select
    project_id,
    count(*) > 0 as has_project_policy
  from
    azuredevops_v_branch_policy_kind
  where
    repository_id is null
  group by
    project_id
)
select
  r.id                                                   as repository_id,
  r.name                                                 as repository_name,
  r.project_id                                           as project_id,
  prj.name                                               as project_name,
  r.web_url                                              as repository_web_url,
  coalesce(rp.total_policy_count, 0)                     as repo_policy_count,
  coalesce(rp.branch_protection_policy_count, 0)         as branch_protection_policy_count,
  coalesce(rp.status_check_policy_count, 0)              as status_check_policy_count,
  coalesce(pp.has_project_policy, false)                 as has_project_level_policy,
  coalesce(rp.total_policy_count, 0) > 0                 as has_repo_policy,
  coalesce(rp.total_policy_count, 0) > 0
    or coalesce(pp.has_project_policy, false)            as has_any_policy
from
  azuredevops_git_repository r
  left join azuredevops_project prj on prj.id = r.project_id
  left join repo_policies rp on rp.repository_id = r.id
  left join project_policies pp on pp.project_id = r.project_id;