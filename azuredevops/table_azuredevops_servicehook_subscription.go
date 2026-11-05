package azuredevops

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/servicehooks"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsServiceHookSubscription(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_servicehook_subscription",
		Description: "Retrieve service hook subscriptions for the organization or collection.",
		List: &plugin.ListConfig{
			Hydrate: listServiceHookSubscriptions,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "publisher_id", Require: plugin.Optional},
				{Name: "event_type", Require: plugin.Optional},
				{Name: "consumer_id", Require: plugin.Optional},
				{Name: "consumer_action_id", Require: plugin.Optional},
				{Name: "status", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getServiceHookSubscription,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The subscription ID.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "publisher_id",
				Description: "The publisher that raises the event, such as tfs.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "event_type",
				Description: "The event type this subscription listens for, such as git.pullrequest.created.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "event_description",
				Description: "A description of the event this subscription listens for.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "consumer_id",
				Description: "The consumer that receives the event, such as webHooks.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "consumer_action_id",
				Description: "The consumer action invoked when the event is raised.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "action_description",
				Description: "A description of the action performed when the event is raised.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "Subscription status (enabled, onProbation, disabledByUser, disabledBySystem, or disabledByInactiveIdentity).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "resource_version",
				Description: "The resource version of the payload delivered to the consumer.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "The URL of the subscription.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "created_date",
				Description: "The date and time when the subscription was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "created_by",
				Description: "The identity that created the subscription.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "modified_date",
				Description: "The date and time when the subscription was last modified.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("ModifiedDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "modified_by",
				Description: "The identity that last modified the subscription.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "subscriber",
				Description: "The identity that owns the subscription.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "publisher_inputs",
				Description: "Publisher-specific input values, such as the project or repository filter.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "consumer_inputs",
				Description: "Consumer-specific input values, such as a webhook URL.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "probation_retries",
				Description: "Number of retries while the subscription is on probation.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("ProbationRetries").Transform(bytePtrToInt),
			},
			{
				Name:        "last_probation_retry_date",
				Description: "The date and time of the last probation retry.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("LastProbationRetryDate").Transform(azureDevOpsToTime),
			},
			{
				Name:        "links",
				Description: "Reference links for the subscription.",
				Type:        proto.ColumnType_JSON,
			},

			/// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("EventType"),
			},
		}),
	}
}

func listServiceHookSubscriptions(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_servicehook_subscription.listServiceHookSubscriptions", "connection_error", err)
		return nil, err
	}
	client := servicehooks.NewClient(ctx, connection)

	input := servicehooks.ListSubscriptionsArgs{}
	if publisherID := qualString(d, "publisher_id"); publisherID != "" {
		input.PublisherId = types.String(publisherID)
	}
	if eventType := qualString(d, "event_type"); eventType != "" {
		input.EventType = types.String(eventType)
	}
	if consumerID := qualString(d, "consumer_id"); consumerID != "" {
		input.ConsumerId = types.String(consumerID)
	}
	if consumerActionID := qualString(d, "consumer_action_id"); consumerActionID != "" {
		input.ConsumerActionId = types.String(consumerActionID)
	}

	subscriptions, err := client.ListSubscriptions(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_servicehook_subscription.listServiceHookSubscriptions", "api_error", err)
		return nil, err
	}
	if subscriptions == nil {
		return nil, nil
	}

	statusQual := qualString(d, "status")
	for _, subscription := range *subscriptions {
		if statusQual != "" && (subscription.Status == nil || !strings.EqualFold(string(*subscription.Status), statusQual)) {
			continue
		}
		d.StreamListItem(ctx, subscription)
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}

func getServiceHookSubscription(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	if h.Item != nil {
		if subscription, ok := h.Item.(servicehooks.Subscription); ok {
			return subscription, nil
		}
	}

	id := qualString(d, "id")
	if id == "" {
		return nil, nil
	}
	subscriptionID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid service hook subscription id %q: %w", id, err)
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_servicehook_subscription.getServiceHookSubscription", "connection_error", err)
		return nil, err
	}
	client := servicehooks.NewClient(ctx, connection)

	subscription, err := client.GetSubscription(ctx, servicehooks.GetSubscriptionArgs{
		SubscriptionId: &subscriptionID,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_servicehook_subscription.getServiceHookSubscription", "id", id, "api_error", err)
		return nil, err
	}
	return subscription, nil
}

func bytePtrToInt(_ context.Context, d *transform.TransformData) (interface{}, error) {
	switch v := d.Value.(type) {
	case byte:
		return int64(v), nil
	case *byte:
		if v == nil {
			return nil, nil
		}
		return int64(*v), nil
	default:
		if d.Value == nil {
			return nil, nil
		}
		return nil, nil
	}
}
