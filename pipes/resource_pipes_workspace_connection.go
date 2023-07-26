package pipes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	pipes "github.com/turbot/pipes-sdk-go"
)

func resourceWorkspaceConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceConnectionCreate,
		ReadContext:   resourceWorkspaceConnectionRead,
		UpdateContext: resourceWorkspaceConnectionUpdate,
		DeleteContext: resourceWorkspaceConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"connection_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_]{0,37}[a-z0-9]?$`), "Handle must be between 1 and 39 characters, and may only contain alphanumeric characters or single underscores, cannot start with a number or underscore and cannot end with an underscore."),
			},
			"workspace_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
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
			"identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"connection_created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_plugin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"workspace_created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_database_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_hive": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_public_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"workspace_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"workspace_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWorkspaceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	workspaceHandle := d.Get("workspace_handle").(string)
	connHandle := d.Get("connection_handle").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var resp pipes.WorkspaceConn
	var err error
	var r *http.Response

	// Create request
	req := pipes.CreateWorkspaceConnRequest{ConnectionHandle: connHandle}

	client := meta.(*PipesClient)
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionCreate. getUserHandler error %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceConnectionAssociations.Create(ctx, actorHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceConnectionAssociations.Create(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating workspace connection association: %v", decodeResponse(r))
	}

	// Set property values
	d.Set("association_id", resp.Id)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("organization", orgHandle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("identity_id", resp.IdentityId)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("workspace_handle", workspaceHandle)
	d.Set("connection_handle", resp.Connection.Handle)
	d.Set("connection_created_at", resp.Connection.CreatedAt)
	d.Set("connection_updated_at", resp.Connection.UpdatedAt)
	d.Set("connection_identity_id", resp.Connection.IdentityId)
	d.Set("connection_plugin", resp.Connection.Plugin)
	d.Set("connection_type", resp.Connection.Type)
	d.Set("connection_version_id", resp.Connection.VersionId)

	// Get the workspace details
	workspaceResp, r, err := getWorkspaceDetails(ctx, client, d)

	// Error check
	if err != nil {
		return diag.Errorf("error getting workspace details for connection association: %v", decodeResponse(r))
	}

	d.Set("workspace_state", workspaceResp.State)
	d.Set("workspace_created_at", workspaceResp.CreatedAt)
	d.Set("workspace_database_name", workspaceResp.DatabaseName)
	d.Set("workspace_hive", workspaceResp.Hive)
	d.Set("workspace_host", workspaceResp.Host)
	d.Set("workspace_identity_id", workspaceResp.IdentityId)
	d.Set("workspace_public_key", workspaceResp.PublicKey)
	d.Set("workspace_updated_at", workspaceResp.UpdatedAt)
	d.Set("workspace_version_id", workspaceResp.VersionId)

	// If workspace connection association is created inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ConnectionHandle" otherwise "WorkspaceHandle/ConnectionHandle"
	id := fmt.Sprintf("%s/%s", workspaceHandle, resp.Connection.Handle)
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s/%s", orgHandle, id))
	} else {
		d.SetId(id)
	}

	return diags
}

func resourceWorkspaceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, connectionHandle string
	var isUser = false

	// For backward-compatibility, we see whether the id contains : or /
	separator := "/"
	if strings.Contains(d.Id(), ":") {
		separator = ":"
	}
	// If workspace connection association is created inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ConnectionHandle" otherwise "WorkspaceHandle/ConnectionHandle"
	idParts := strings.Split(d.Id(), separator)
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<connection-handle>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		connectionHandle = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		connectionHandle = idParts[1]
	}

	var resp pipes.WorkspaceConn
	var err error
	var r *http.Response

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionRead. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceConnectionAssociations.Get(ctx, actorHandle, workspaceHandle, connectionHandle).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceConnectionAssociations.Get(ctx, orgHandle, workspaceHandle, connectionHandle).Execute()
	}

	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Association (%s) not found", resp.Id),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceWorkspaceConnectionRead. Get workspace connection association error: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Association received: %s", resp.Id)

	if separator == ":" {
		d.SetId(strings.ReplaceAll(d.Id(), ":", "/"))
	}
	d.Set("association_id", resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("organization", orgHandle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("identity_id", resp.IdentityId)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("connection_handle", resp.Connection.Handle)
	d.Set("workspace_handle", workspaceHandle)
	d.Set("connection_created_at", resp.Connection.CreatedAt)
	d.Set("connection_updated_at", resp.Connection.UpdatedAt)
	d.Set("connection_identity_id", resp.Connection.IdentityId)
	d.Set("connection_plugin", resp.Connection.Plugin)
	d.Set("connection_type", resp.Connection.Type)
	d.Set("connection_version_id", resp.Connection.VersionId)

	// Get the workspace details
	workspaceResp, r, err := getWorkspaceDetails(ctx, client, d)

	// Error check
	if err != nil {
		return diag.Errorf("error getting workspace details for connection association: %v", decodeResponse(r))
	}

	d.Set("workspace_state", workspaceResp.State)
	d.Set("workspace_created_at", workspaceResp.CreatedAt)
	d.Set("workspace_database_name", workspaceResp.DatabaseName)
	d.Set("workspace_hive", workspaceResp.Hive)
	d.Set("workspace_host", workspaceResp.Host)
	d.Set("workspace_identity_id", workspaceResp.IdentityId)
	d.Set("workspace_public_key", workspaceResp.PublicKey)
	d.Set("workspace_updated_at", workspaceResp.UpdatedAt)
	d.Set("workspace_version_id", workspaceResp.VersionId)

	return diags
}

func resourceWorkspaceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	workspaceHandle := d.State().Attributes["workspace_handle"]
	connHandle := d.State().Attributes["connection_handle"]
	orgHandle := d.State().Attributes["organization"]

	if d.HasChange("workspace_handle") {
		_, newWorkspaceHandle := d.GetChange("workspace_handle")
		workspaceHandle = newWorkspaceHandle.(string)
	}
	if d.HasChange("connection_handle") {
		_, newConnHandle := d.GetChange("connection_handle")
		connHandle = newConnHandle.(string)
	}

	var id string
	if workspaceHandle != "" && connHandle != "" {
		id = fmt.Sprintf("%s/%s", workspaceHandle, connHandle)
		d.Set("workspace_handle", workspaceHandle)
		d.Set("connection_handle", connHandle)
	}
	if orgHandle != "" {
		d.SetId(fmt.Sprintf("%s/%s", orgHandle, id))
	} else {
		d.SetId(id)
	}

	return diags
}

func resourceWorkspaceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, connectionHandle string
	var isUser = false

	// For backward-compatibility, we see whether the id contains : or /
	separator := "/"
	if strings.Contains(d.Id(), ":") {
		separator = ":"
	}
	idParts := strings.Split(d.Id(), separator)
	if len(idParts) < 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<connection-handle>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		connectionHandle = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		connectionHandle = idParts[1]
	}

	log.Printf("\n[DEBUG] Deleting Workspace Connection association: %s", fmt.Sprintf("%s/%s", workspaceHandle, connectionHandle))

	var err error
	var r *http.Response

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionDelete. getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceConnectionAssociations.Delete(ctx, actorHandle, workspaceHandle, connectionHandle).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceConnectionAssociations.Delete(ctx, orgHandle, workspaceHandle, connectionHandle).Execute()
	}

	if err != nil {
		return diag.Errorf("error deleting workspace connection association: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
