create or replace view azuredevops_v_branch_policy_kind as
select
  *,
  case
    when lower(policy_type_display_name) in ('status', 'required status checks', 'status checks')
      then 'status_check'
    else 'branch_protection'
  end as policy_kind,
  case
    when repository_id is null           then 'project'
    when ref_name is null                then 'repository'
    else 'branch'
  end as scope_level
from
  azuredevops_v_branch_policy_scope
where
  is_enabled
  and not coalesce(is_deleted, false);
