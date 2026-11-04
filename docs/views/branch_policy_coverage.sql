
create or replace view azuredevops_v_branch_policy_scope as
select
  p.id,
  p.project_id,
  p.policy_type_id,
  p.policy_type_display_name,
  p.is_enabled,
  p.is_blocking,
  p.is_deleted,
  nullif(s ->> 'repositoryId', '')             as repository_id,
  s ->> 'refName'                               as ref_name,
  replace(s ->> 'refName', 'refs/heads/', '')   as branch,
  s ->> 'matchKind'                             as match_kind
from
  azuredevops_git_repository_branch_policy p
  left join lateral jsonb_array_elements(
    case
      when jsonb_typeof(p.scope::jsonb) = 'array'
       and jsonb_array_length(p.scope::jsonb) > 0
        then p.scope::jsonb
      else '[{}]'::jsonb
    end
  ) s on true;

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

create or replace view azuredevops_v_branch_policy_readable as
select
  k.id,
  k.project_id,
  prj.name                                        as project_name,
  k.repository_id,
  r.name                                          as repository_name,
  k.policy_type_id,
  k.policy_type_display_name,
  k.policy_kind,
  k.scope_level,
  k.is_enabled,
  k.is_blocking,
  k.ref_name,
  k.branch,
  k.match_kind,
  r.web_url                                       as repository_web_url,
  split_part(prj.url, '/_apis/', 1) || '/' || prj.name
                                                  as project_web_url,
  split_part(prj.url, '/_apis/', 1) || '/' || prj.name
    || '/_settings/repositories?repo=' || coalesce(k.repository_id, '')
    || case
         when k.ref_name is not null
           then '&refs=' || replace(k.ref_name, '/', '%2F')
         else ''
       end
    || '&_a=policiesMid'                          as policy_url
from
  azuredevops_v_branch_policy_kind k
  left join azuredevops_project prj on prj.id = k.project_id
  left join azuredevops_git_repository r on r.id = k.repository_id;

create or replace view azuredevops_v_branch_coverage as
with branch_protection as (
  select repository_id, ref_name, count(*) as policy_count
  from azuredevops_v_branch_policy_kind
  where policy_kind = 'branch_protection' and repository_id is not null and ref_name is not null
  group by repository_id, ref_name
),
status_checks as (
  select repository_id, ref_name, count(*) as policy_count
  from azuredevops_v_branch_policy_kind
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
  azuredevops_git_repository r
  join azuredevops_git_repository_branch b on b.repository_id = r.id
  left join azuredevops_project prj on prj.id = r.project_id
  left join branch_protection bp
    on bp.repository_id = r.id and bp.ref_name = b.name
  left join status_checks sc
    on sc.repository_id = r.id and sc.ref_name = b.name;





