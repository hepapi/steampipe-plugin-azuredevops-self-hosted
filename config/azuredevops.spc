connection "azuredevops" {
  plugin = "azuredevops"

  # `organization_url`: Azure DevOps Organization URL (SaaS) or Collection URL (self-hosted). (Required)
  # For SaaS, this is the organization URL, e.g. https://dev.azure.com/test
  # For self-hosted Azure DevOps Server, this is a specific collection URL, e.g. https://code.digiturk.net/SCMPOCCollection
  # For more information on the Organization URL, please see https://learn.microsoft.com/en-us/azure/devops/extend/develop/work-with-urls?view=azure-devops&tabs=http.
  # Can also be set with the AZDO_ORG_SERVICE_URL environment variable.
  # organization_url = "https://dev.azure.com/test"

  # `server_url`: Azure DevOps Server root URL (self-hosted only, optional).
  # Used only by the `azuredevopssh_project_collection` table to enumerate collections.
  # When unset, it is derived from `organization_url` by stripping the last path segment.
  # For self-hosted, e.g. https://code.digiturk.net/
  # server_url = "https://code.digiturk.net/"

  # `personal_access_token`: Azure DevOps Personal Access Token. (Required)
  # For more information on the Personal Access Token, please see https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows.
  # Can also be set with the AZDO_PERSONAL_ACCESS_TOKEN environment variable.
  # personal_access_token = "......................"
}
