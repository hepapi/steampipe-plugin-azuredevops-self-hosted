package azuredevops

import (
	"context"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/wiki"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

const wikiPageBatchSize = 100

// wikiPageRow is shared by the batch list, which returns id and path, and the
// get, which can also return page content.
type wikiPageRow struct {
	Id           int
	Path         string
	ProjectId    string
	WikiId       string
	Order        *int
	IsParentPage *bool
	GitItemPath  *string
	Content      *string
	Url          *string
}

func tableAzureDevOpsWikiPage(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_wiki_page",
		Description: "Retrieve pages in an Azure DevOps wiki.",
		List: &plugin.ListConfig{
			Hydrate: listWikiPages,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "project_id", Require: plugin.Required},
				{Name: "wiki_id", Require: plugin.Required},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "id", Require: plugin.Required},
				{Name: "project_id", Require: plugin.Required},
				{Name: "wiki_id", Require: plugin.Required},
			},
			Hydrate: getWikiPage,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Wiki page identifier.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "wiki_id",
				Description: "Wiki that contains the page.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "Project that owns the wiki.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "path",
				Description: "Page path.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "order",
				Description: "Order of the page among its siblings.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "is_parent_page",
				Description: "True when the page has subpages.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "git_item_path",
				Description: "Path of the backing git item.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "content",
				Description: "Page content. Populated when one page is fetched by id.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "REST URL of the page.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Path"),
			},
		}),
	}
}

func listWikiPages(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	projectID := qualString(d, "project_id")
	wikiID := qualString(d, "wiki_id")
	if projectID == "" || wikiID == "" {
		return nil, fmt.Errorf("azuredevops_wiki_page requires project_id and wiki_id")
	}
	pageID, hasPageID := qualInt(d, "id")

	client, err := newWikiClient(ctx, d, "azuredevops_wiki_page.listWikiPages")
	if err != nil {
		return nil, err
	}

	top := wikiPageBatchSize
	var token string
	for {
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
		request := wiki.WikiPagesBatchRequest{Top: &top}
		if token != "" {
			request.ContinuationToken = &token
		}
		page, err := client.GetPagesBatch(ctx, wiki.GetPagesBatchArgs{
			Project:           &projectID,
			WikiIdentifier:    &wikiID,
			PagesBatchRequest: &request,
		})
		if err != nil {
			plugin.Logger(ctx).Error("azuredevops_wiki_page.listWikiPages", "api_error", err)
			return nil, err
		}
		if page == nil || len(page.Value) == 0 {
			return nil, nil
		}
		for _, item := range page.Value {
			if item.Id == nil {
				continue
			}
			if hasPageID && int64(*item.Id) != pageID {
				continue
			}
			path := ""
			if item.Path != nil {
				path = *item.Path
			}
			d.StreamListItem(ctx, wikiPageRow{
				Id:        *item.Id,
				Path:      path,
				ProjectId: projectID,
				WikiId:    wikiID,
			})
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if page.ContinuationToken == "" {
			return nil, nil
		}
		token = page.ContinuationToken
	}
}

func getWikiPage(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	pageID, ok := qualInt(d, "id")
	projectID := qualString(d, "project_id")
	wikiID := qualString(d, "wiki_id")
	if !ok || projectID == "" || wikiID == "" {
		return nil, nil
	}
	client, err := newWikiClient(ctx, d, "azuredevops_wiki_page.getWikiPage")
	if err != nil {
		return nil, err
	}
	id := int(pageID)
	response, err := client.GetPageById(ctx, wiki.GetPageByIdArgs{
		Project:        &projectID,
		WikiIdentifier: &wikiID,
		Id:             &id,
		IncludeContent: boolPtr(true),
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_wiki_page.getWikiPage", "api_error", err)
		return nil, err
	}
	if response == nil || response.Page == nil || response.Page.Id == nil {
		return nil, nil
	}
	page := response.Page
	path := ""
	if page.Path != nil {
		path = *page.Path
	}
	return wikiPageRow{
		Id:           *page.Id,
		Path:         path,
		ProjectId:    projectID,
		WikiId:       wikiID,
		Order:        page.Order,
		IsParentPage: page.IsParentPage,
		GitItemPath:  page.GitItemPath,
		Content:      page.Content,
		Url:          page.Url,
	}, nil
}
