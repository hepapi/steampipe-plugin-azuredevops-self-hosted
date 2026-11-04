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
