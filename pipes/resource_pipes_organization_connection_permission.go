package pipes

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	pipes "github.com/turbot/pipes-sdk-go"
)

func resourceOrganizationConnectionPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationConnectionPermissionCreate,
		ReadContext:   resourceOrganizationConnectionPermissionRead,
		UpdateContext: resourceOrganizationConnectionPermissionUpdate,
		DeleteContext: resourceOrganizationConnectionPermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"permission_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_id": {
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
				Required: true,
			},
			"connection_handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tenant_handle": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      false,
				ConflictsWith: []string{"identity_handle", "workspace_handle"},
			},
			"identity_handle": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      false,
				ConflictsWith: []string{"tenant_handle"},
			},
			"workspace_handle": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      false,
				RequiredWith:  []string{"identity_handle"},
				ConflictsWith: []string{"tenant_handle"},
			},
		},
	}
}

func resourceOrganizationConnectionPermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var tenantHandle, identityHandle, workspaceHandle, connectionHandle, orgHandle string
	var err error

	orgHandle = d.Get("organization").(string)

	// When attaching a workspace schema, we can pass in a connection folder id, connection handle or aggregator handle
	// Its already verified as part of schema validation rules that only one of these can be defined in configuration
	if val, ok := d.GetOk("tenant_handle"); ok {
		tenantHandle = val.(string)
	}
	if val, ok := d.GetOk("identity_handle"); ok {
		identityHandle = val.(string)
	}
	if val, ok := d.GetOk("workspace_handle"); ok {
		workspaceHandle = val.(string)
	}
	if val, ok := d.GetOk("connection_handle"); ok {
		connectionHandle = val.(string)
	}

	// Frame the request object
	req := pipes.CreatePermissionRequest{
		TenantHandle:    &tenantHandle,
		IdentityHandle:  &identityHandle,
		WorkspaceHandle: &workspaceHandle,
	}

	client := meta.(*PipesClient)
	var resp pipes.Permission
	var r *http.Response

	// Create permission for connection
	resp, r, err = client.APIClient.OrgConnections.CreatePermission(ctx, orgHandle, connectionHandle).Request(req).Execute()
	// Error check
	if err != nil {
		return diag.Errorf("error creating organization connection permission: %v", decodeResponse(r))
	}

	// Set property values
	d.Set("permission_id", resp.Id)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	if resp.Tenant != nil {
		d.Set("tenant_handle", resp.Tenant.TenantHandle)
	}
	if resp.Identity != nil {
		d.Set("identity_handle", resp.Identity.IdentityHandle)
	}
	if resp.Workspace != nil {
		d.Set("workspace_handle", resp.Workspace.WorkspaceHandle)
	}
	// ID formats
	// Tenant Connection - "OrganizationHandle/ConnectionHandle/PermissionId"
	d.Set("organization", orgHandle)
	d.Set("connection_handle", connectionHandle)
	d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, connectionHandle, resp.Id))

	return diags
}

func resourceOrganizationConnectionPermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, connectionHandle, permissionId string

	// ID formats
	// Tenant Connection - "OrganizationHandle/ConnectionHandle/PermissionId"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization-handle>/<connection-handle>/<permission-id>", d.Id())
	}

	orgHandle = idParts[0]
	connectionHandle = idParts[1]
	permissionId = idParts[2]

	var resp pipes.Permission
	var err error
	var r *http.Response

	resp, r, err = client.APIClient.OrgConnections.GetPermission(ctx, orgHandle, connectionHandle, permissionId).Execute()
	// Error check
	if err != nil {
		return diag.Errorf("error getting tenant connection permission: %v", decodeResponse(r))
	}

	// Set property values
	d.Set("permission_id", resp.Id)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	if resp.Tenant != nil {
		d.Set("tenant_handle", resp.Tenant.TenantHandle)
	}
	if resp.Identity != nil {
		d.Set("identity_handle", resp.Identity.IdentityHandle)
	}
	if resp.Workspace != nil {
		d.Set("workspace_handle", resp.Workspace.WorkspaceHandle)
	}
	// ID formats
	// Tenant Connection - "OrganizationHandle/ConnectionHandle/PermissionId"
	d.Set("organization", orgHandle)
	d.Set("connection_handle", connectionHandle)
	d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, connectionHandle, resp.Id))

	return diags
}

func resourceOrganizationConnectionPermissionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var tenantHandle, identityHandle, workspaceHandle, orgHandle, connectionHandle, permissionId string
	var err error

	// ID formats
	// Tenant Connection - "OrganizationHandle/ConnectionHandle/PermissionId"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization-handle>/<connection-handle>/<permission-id>", d.Id())
	}

	orgHandle = idParts[0]
	permissionId = idParts[2]

	// When attaching a workspace schema, we can pass in a connection folder id, connection handle or aggregator handle
	// Its already verified as part of schema validation rules that only one of these can be defined in configuration
	if val, ok := d.GetOk("tenant_handle"); ok {
		tenantHandle = val.(string)
	}
	if val, ok := d.GetOk("identity_handle"); ok {
		identityHandle = val.(string)
	}
	if val, ok := d.GetOk("workspace_handle"); ok {
		workspaceHandle = val.(string)
	}
	if val, ok := d.GetOk("connection_handle"); ok {
		connectionHandle = val.(string)
	}

	// Frame the request object
	req := pipes.UpdatePermissionRequest{
		TenantHandle:    &tenantHandle,
		IdentityHandle:  &identityHandle,
		WorkspaceHandle: &workspaceHandle,
	}

	client := meta.(*PipesClient)
	var resp pipes.Permission
	var r *http.Response

	// Create permission for connection
	resp, r, err = client.APIClient.OrgConnections.UpdatePermission(ctx, orgHandle, connectionHandle, permissionId).Request(req).Execute()
	// Error check
	if err != nil {
		return diag.Errorf("error updating tenant connection permission: %v", decodeResponse(r))
	}

	// Set property values
	d.Set("permission_id", resp.Id)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	if resp.Tenant != nil {
		d.Set("tenant_handle", resp.Tenant.TenantHandle)
	}
	if resp.Identity != nil {
		d.Set("identity_handle", resp.Identity.IdentityHandle)
	}
	if resp.Workspace != nil {
		d.Set("workspace_handle", resp.Workspace.WorkspaceHandle)
	}
	// ID formats
	// Tenant Connection - "OrganizationHandle/ConnectionHandle/PermissionId"
	d.Set("organization", orgHandle)
	d.Set("connection_handle", connectionHandle)
	d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, connectionHandle, resp.Id))

	return diags
}

func resourceOrganizationConnectionPermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, connectionHandle, permissionId string

	// ID formats
	// Tenant Connection - "OrganizationHandle/ConnectionHandle/PermissionId"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization-handle>/<connection-handle>/<permission-id>", d.Id())
	}

	orgHandle = idParts[0]
	connectionHandle = idParts[1]
	permissionId = idParts[2]

	var err error
	var r *http.Response

	_, r, err = client.APIClient.OrgConnections.DeletePermission(ctx, orgHandle, connectionHandle, permissionId).Execute()
	if err != nil {
		return diag.Errorf("error deleting permission from organization connection: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
