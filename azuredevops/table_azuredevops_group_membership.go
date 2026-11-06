package azuredevops

import (
	"context"
	"fmt"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

type groupMembershipRow struct {
	SubjectDescriptor   string
	Direction           string
	ContainerDescriptor string
	MemberDescriptor    string
}

type membershipEdge struct {
	Container string
	Member    string
}

func tableAzureDevOpsGroupMembership(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_group_membership",
		Description: "Retrieve direct group membership for an Azure DevOps subject.",
		List: &plugin.ListConfig{
			Hydrate: listGroupMemberships,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "subject_descriptor", Require: plugin.Required},
				{Name: "direction", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "subject_descriptor",
				Description: "Subject the membership was queried for. A graph subject descriptor or a legacy identity descriptor.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "direction",
				Description: "Traversal direction. down lists members of the subject. up lists groups the subject belongs to. Defaults to down.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "container_descriptor",
				Description: "Descriptor of the group in the membership.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "member_descriptor",
				Description: "Descriptor of the member in the membership.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("MemberDescriptor"),
			},
		}),
	}
}

func listGroupMemberships(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	subject := qualString(d, "subject_descriptor")
	if subject == "" {
		return nil, fmt.Errorf("azuredevops_group_membership requires subject_descriptor")
	}
	direction, err := membershipDirection(qualString(d, "direction"))
	if err != nil {
		return nil, err
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_group_membership.listGroupMemberships", "connection_error", err)
		return nil, err
	}

	client, err := graph.NewClient(ctx, connection)
	if err != nil && !apiNotRegistered(err) {
		plugin.Logger(ctx).Error("azuredevops_group_membership.listGroupMemberships", "client_error", err)
		return nil, err
	}
	if err == nil {
		graphDirection := graph.GraphTraversalDirection(direction)
		depth := 1
		memberships, listErr := client.ListMemberships(ctx, graph.ListMembershipsArgs{
			SubjectDescriptor: &subject,
			Direction:         &graphDirection,
			Depth:             &depth,
		})
		if listErr == nil {
			return nil, streamGraphMemberships(ctx, d, subject, direction, memberships)
		}
		if !apiNotRegistered(listErr) {
			plugin.Logger(ctx).Error("azuredevops_group_membership.listGroupMemberships", "api_error", listErr)
			return nil, listErr
		}
	}

	return nil, listGroupMembershipsFromIdentity(ctx, d, connection, subject, direction)
}

func streamGraphMemberships(ctx context.Context, d *plugin.QueryData, subject, direction string, memberships *[]graph.GraphMembership) error {
	if memberships == nil {
		return nil
	}
	for _, membership := range *memberships {
		d.StreamListItem(ctx, groupMembershipRow{
			SubjectDescriptor:   subject,
			Direction:           direction,
			ContainerDescriptor: stringValue(membership.ContainerDescriptor),
			MemberDescriptor:    stringValue(membership.MemberDescriptor),
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil
		}
	}
	return nil
}

func listGroupMembershipsFromIdentity(ctx context.Context, d *plugin.QueryData, connection *azuredevops.Connection, subject, direction string) error {
	item, err := identityForMembership(ctx, connection, subject)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_group_membership.listGroupMemberships", "api_error", err)
		return err
	}
	if item == nil {
		return nil
	}

	self := stringValue(item.Descriptor)
	if self == "" {
		self = subject
	}
	var members, memberOf []string
	if item.Members != nil {
		members = *item.Members
	}
	if item.MemberOf != nil {
		memberOf = *item.MemberOf
	}

	for _, edge := range membershipEdges(direction, self, members, memberOf) {
		d.StreamListItem(ctx, groupMembershipRow{
			SubjectDescriptor:   subject,
			Direction:           direction,
			ContainerDescriptor: edge.Container,
			MemberDescriptor:    edge.Member,
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil
		}
	}
	return nil
}

func identityForMembership(ctx context.Context, connection *azuredevops.Connection, subject string) (*identity.Identity, error) {
	direct := identity.QueryMembershipValues.Direct
	bySubject, err := readIdentities(ctx, connection, identity.ReadIdentitiesArgs{
		SubjectDescriptors: &subject,
		QueryMembership:    &direct,
	})
	if err != nil {
		return nil, err
	}
	if bySubject != nil && len(*bySubject) > 0 {
		item := (*bySubject)[0]
		return &item, nil
	}

	byDescriptor, err := readIdentities(ctx, connection, identity.ReadIdentitiesArgs{
		Descriptors:     &subject,
		QueryMembership: &direct,
	})
	if err != nil {
		return nil, err
	}
	if byDescriptor == nil || len(*byDescriptor) == 0 {
		return nil, nil
	}
	item := (*byDescriptor)[0]
	return &item, nil
}

func membershipDirection(value string) (string, error) {
	switch strings.ToLower(value) {
	case "", "down":
		return "down", nil
	case "up":
		return "up", nil
	default:
		return "", fmt.Errorf("azuredevops_group_membership direction must be up or down")
	}
}

func membershipEdges(direction, self string, members, memberOf []string) []membershipEdge {
	var edges []membershipEdge
	switch direction {
	case "up":
		for _, container := range memberOf {
			edges = append(edges, membershipEdge{Container: container, Member: self})
		}
	default:
		for _, member := range members {
			edges = append(edges, membershipEdge{Container: self, Member: member})
		}
	}
	return edges
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
