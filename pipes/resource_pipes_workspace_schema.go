package pipes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	pipes "github.com/turbot/pipes-sdk-go"
)

func resourceWorkspaceSchema() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceSchemaCreate,
		ReadContext:   resourceWorkspaceSchemaRead,
		DeleteContext: resourceWorkspaceSchemaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"workspace_schema_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aggregator_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"updated_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
			},
			"connection_folder_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     false,
				ExactlyOneOf: []string{"connection_folder_id", "connection_handle", "aggregator_handle"},
			},
			"connection_handle": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     false,
				ExactlyOneOf: []string{"connection_folder_id", "connection_handle", "aggregator_handle"},
			},
			"aggregator_handle": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     false,
				ExactlyOneOf: []string{"connection_folder_id", "connection_handle", "aggregator_handle"},
			},
		},
	}
}

func resourceWorkspaceSchemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var workspaceHandle, connectionFolderId, connectionHandle, aggregatorHandle string
	var err error

	// Get details about the workspace where the connection folder would be created
	if val, ok := d.GetOk("workspace"); ok {
		workspaceHandle = val.(string)
	}
	// When attaching a workspace schema, we can pass in a connection folder id, connection handle or aggregator handle
	// Its already verified as part of schema validation rules that only one of these can be defined in configuration
	if val, ok := d.GetOk("connection_folder_id"); ok {
		connectionFolderId = val.(string)
	}
	if val, ok := d.GetOk("connection_handle"); ok {
		connectionHandle = val.(string)
	}
	if val, ok := d.GetOk("aggregator_handle"); ok {
		aggregatorHandle = val.(string)
	}

	// Create request
	req := pipes.AttachWorkspaceSchemaRequest{}
	if connectionFolderId != "" {
		req.SetConnectionFolder(connectionFolderId)
	} else if connectionHandle != "" {
		req.SetConnectionHandle(connectionHandle)
	} else if aggregatorHandle != "" {
		req.SetAggregatorHandle(aggregatorHandle)
	}

	client := meta.(*PipesClient)
	var resp pipes.WorkspaceSchema
	var r *http.Response

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceSchemaCreate. getUserHandler error %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceSchemas.Attach(ctx, actorHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceSchemas.Attach(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}
	// Error check
	if err != nil {
		return diag.Errorf("error creating workspace connection association: %v", decodeResponse(r))
	}

	// Set property values
	d.Set("workspace_schema_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("aggregator_id", resp.AggregatorId)
	d.Set("name", resp.Name)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	d.Set("workspace", workspaceHandle)
	// ID formats
	// User workspace schema - "WorkspaceHandle/{schema_type}/SchemaHandle"
	// Org workspace schema - "OrganizationHandle/WorkspaceHandle/{schema_type}/SchemaHandle"
	var id string
	if connectionFolderId != "" {
		d.Set("connection_folder_id", connectionFolderId)
		id = fmt.Sprintf("%s/connection-folder/%s", workspaceHandle, connectionFolderId)
	} else if connectionHandle != "" {
		d.Set("connection_handle", connectionHandle)
		id = fmt.Sprintf("%s/connection/%s", workspaceHandle, connectionHandle)
	} else if aggregatorHandle != "" {
		d.Set("aggregator_handle", aggregatorHandle)
		id = fmt.Sprintf("%s/aggregator/%s", workspaceHandle, aggregatorHandle)
	}
	if !isUser {
		d.SetId(fmt.Sprintf("%s/%s", orgHandle, id))
	} else {
		d.SetId(id)
	}

	return diags
}

func resourceWorkspaceSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, schemaType, schemaHandle string
	var isUser = false

	// ID formats
	// User workspace schema - "WorkspaceHandle/{schema_type}/SchemaHandle"
	// Org workspace schema - "OrganizationHandle/WorkspaceHandle/{schema_type}/SchemaHandle"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 3 && len(idParts) > 4 {
		return diag.Errorf("unexpected format of ID (%q), expected <org-handle>/<workspace-handle>/<schema-type>/<schema-handle>", d.Id())
	}

	if len(idParts) == 4 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		schemaType = idParts[2]
		schemaHandle = idParts[2]
	} else if len(idParts) == 3 {
		isUser = true
		workspaceHandle = idParts[0]
		schemaType = idParts[1]
		schemaHandle = idParts[2]
	}

	var respSchema pipes.WorkspaceSchema
	var respAssociation pipes.WorkspaceConn
	var err error
	var r *http.Response

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionRead. getUserHandler error  %v", decodeResponse(r))
		}
		if schemaType == "connection-folder" {
			respAssociation, r, err = client.APIClient.UserWorkspaceConnectionAssociations.Get(ctx, actorHandle, workspaceHandle, schemaHandle).Execute()
		} else {
			respSchema, r, err = client.APIClient.UserWorkspaceSchemas.Get(ctx, actorHandle, workspaceHandle, schemaHandle).Execute()
		}
	} else {
		if schemaType == "connection-folder" {
			respAssociation, r, err = client.APIClient.OrgWorkspaceConnectionAssociations.Get(ctx, orgHandle, workspaceHandle, schemaHandle).Execute()
		} else {
			respSchema, r, err = client.APIClient.OrgWorkspaceSchemas.Get(ctx, orgHandle, workspaceHandle, schemaHandle).Execute()
		}
	}

	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Schema (%s) not found", schemaHandle),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceWorkspaceSchemaRead. Get workspace connection association error: %v", decodeResponse(r))
	}

	if schemaType == "connection-folder" {
		d.Set("workspace_schema_id", respAssociation.Id)
		d.Set("identity_id", respAssociation.IdentityId)
		d.Set("workspace_id", respAssociation.WorkspaceId)
		d.Set("connection_id", respAssociation.ConnectionId)
		d.Set("created_at", respAssociation.CreatedAt)
		d.Set("updated_at", respAssociation.UpdatedAt)
		if respAssociation.CreatedBy != nil {
			d.Set("created_by", respAssociation.CreatedBy.Handle)
		}
		if respAssociation.UpdatedBy != nil {
			d.Set("updated_by", respAssociation.UpdatedBy.Handle)
		}
		d.Set("version_id", respAssociation.VersionId)
		d.Set("organization", orgHandle)
		d.Set("workspace", workspaceHandle)
		d.Set("connection_folder_id", schemaHandle)
		id := fmt.Sprintf("%s/connection-folder/%s", workspaceHandle, schemaHandle)
		if !isUser {
			d.SetId(fmt.Sprintf("%s/%s", orgHandle, id))
		} else {
			d.SetId(id)
		}
	} else {
		d.Set("workspace_schema_id", respSchema.Id)
		d.Set("identity_id", respSchema.IdentityId)
		d.Set("workspace_id", respSchema.WorkspaceId)
		d.Set("connection_id", respSchema.ConnectionId)
		d.Set("aggregator_id", respSchema.AggregatorId)
		d.Set("name", respSchema.Name)
		d.Set("type", respSchema.Type)
		d.Set("created_at", respSchema.CreatedAt)
		d.Set("updated_at", respSchema.UpdatedAt)
		if respSchema.CreatedBy != nil {
			d.Set("created_by", respSchema.CreatedBy.Handle)
		}
		if respSchema.UpdatedBy != nil {
			d.Set("updated_by", respSchema.UpdatedBy.Handle)
		}
		d.Set("version_id", respSchema.VersionId)
		d.Set("organization", orgHandle)
		d.Set("workspace", workspaceHandle)
		var id string
		if schemaType == "connection" {
			d.Set("connection_handle", schemaHandle)
			id = fmt.Sprintf("%s/connection/%s", workspaceHandle, schemaHandle)
		} else if schemaType == "aggregator" {
			d.Set("aggregator_handle", schemaHandle)
			id = fmt.Sprintf("%s/aggregator/%s", workspaceHandle, schemaHandle)
		}
		if !isUser {
			d.SetId(fmt.Sprintf("%s/%s", orgHandle, id))
		} else {
			d.SetId(id)
		}
	}

	return diags
}

func resourceWorkspaceSchemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, schemaHandle string
	var isUser = false

	// ID formats
	// User workspace schema - "WorkspaceHandle/{schema_type}/SchemaHandle"
	// Org workspace schema - "OrganizationHandle/WorkspaceHandle/{schema_type}/SchemaHandle"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 3 && len(idParts) > 4 {
		return diag.Errorf("unexpected format of ID (%q), expected <org-handle>/<workspace-handle>/<schema-type>/<schema-handle>", d.Id())
	}

	if len(idParts) == 4 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		schemaHandle = idParts[2]
	} else if len(idParts) == 3 {
		isUser = true
		workspaceHandle = idParts[0]
		schemaHandle = idParts[2]
	}

	log.Printf("\n[DEBUG] Detaching Workspace schema: %s", fmt.Sprintf("%s/%s", workspaceHandle, schemaHandle))

	var err error
	var r *http.Response

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceSchemaDelete. getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceSchemas.Detach(ctx, actorHandle, workspaceHandle, schemaHandle).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceConnectionAssociations.Delete(ctx, orgHandle, workspaceHandle, schemaHandle).Execute()
	}

	if err != nil {
		return diag.Errorf("error detaching schema from workspace: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
