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

func resourceWorkspaceConnectionFolder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceConnectionFolderCreate,
		ReadContext:   resourceWorkspaceConnectionFolderRead,
		UpdateContext: resourceWorkspaceConnectionFolderUpdate,
		DeleteContext: resourceWorkspaceConnectionFolderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"connection_folder_id": {
				Type:     schema.TypeString,
				ForceNew: true,
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
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"parent_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"integration_resource_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_resource_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_resource_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_resource_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"managed_by_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"trunk": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
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
		},
	}
}

func resourceWorkspaceConnectionFolderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var workspaceHandle, title, parentId string
	var err error

	// Get details about the workspace where the connection folder would be created
	if val, ok := d.GetOk("workspace"); ok {
		workspaceHandle = val.(string)
	}

	// Get title and parent_id for the connection folder to be created
	if value, ok := d.GetOk("title"); ok {
		title = value.(string)
	}
	if value, ok := d.GetOk("parent_id"); ok {
		parentId = value.(string)
	}

	req := pipes.CreateConnectionFolderRequest{
		Title: title,
	}

	// Pass the parent_id if its set
	if parentId != "" {
		req.SetParentId(parentId)
	}

	client := meta.(*PipesClient)
	var resp pipes.Connection
	var r *http.Response

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionFolderCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceConnectionFolders.Create(ctx, actorHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceConnectionFolders.Create(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceConnectionFolderCreate. Create connection folder api error  %v", decodeResponse(r))
	}

	d.Set("connection_folder_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("title", resp.Title)
	d.Set("type", resp.Type)
	d.Set("parent_id", resp.ParentId)
	d.Set("integration_resource_name", resp.IntegrationResourceName)
	d.Set("integration_resource_identifier", resp.IntegrationResourceIdentifier)
	d.Set("integration_resource_type", resp.IntegrationResourceType)
	d.Set("integration_resource_path", resp.IntegrationResourcePath)
	d.Set("managed_by_id", resp.ManagedById)
	d.Set("trunk", resp.Trunk)
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
	// The id would be of format "WorkspaceHandle/ConnectionFolderId" for user workspaces
	// or "OrganizationHandle/WorkspaceHandle/ConnectionFolderId" for org workspaces
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Id))

	}

	return diags
}

func resourceWorkspaceConnectionFolderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var connectionFolderId, orgHandle, workspaceHandle string
	var isUser = false
	var err error
	var r *http.Response
	var resp pipes.Connection
	var diags diag.Diagnostics

	// Id would be of format "WorkspaceHandle/ConnectionFolderId" or "OrganizationHandle/WorkspaceHandle/ConnectionFolderId"
	ids := strings.Split(d.Id(), "/")
	if len(ids) == 3 {
		orgHandle = ids[0]
		workspaceHandle = ids[1]
		connectionFolderId = ids[2]
	} else if len(ids) == 2 {
		workspaceHandle = ids[0]
		connectionFolderId = ids[1]
		isUser = true
	}

	if workspaceHandle == "" {
		return diag.Errorf("resourceWorkspaceConnectionFolderRead. Workspace information not present.")
	}
	if connectionFolderId == "" {
		return diag.Errorf("resourceWorkspaceConnectionFolderRead. Connection folder id not present.")
	}

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionFolderRead. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceConnectionFolders.Get(context.Background(), actorHandle, workspaceHandle, connectionFolderId).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceConnectionFolders.Get(context.Background(), orgHandle, workspaceHandle, connectionFolderId).Execute()
	}
	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Connection Folder (%s) not found", connectionFolderId),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceWorkspaceConnectionFolderRead. Get connection error: %v", decodeResponse(r))
	}

	d.Set("connection_folder_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("title", resp.Title)
	d.Set("type", resp.Type)
	d.Set("parent_id", resp.ParentId)
	d.Set("integration_resource_name", resp.IntegrationResourceName)
	d.Set("integration_resource_identifier", resp.IntegrationResourceIdentifier)
	d.Set("integration_resource_type", resp.IntegrationResourceType)
	d.Set("integration_resource_path", resp.IntegrationResourcePath)
	d.Set("managed_by_id", resp.ManagedById)
	d.Set("trunk", resp.Trunk)
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
	// The id would be of format "WorkspaceHandle/ConnectionFolderId" for user workspaces
	// or "OrganizationHandle/WorkspaceHandle/ConnectionFolderId" for org workspaces
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Id))

	}

	return diags
}

func resourceWorkspaceConnectionFolderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	var connectionFolderId, orgHandle, workspaceHandle string
	var err error
	var r *http.Response
	var resp pipes.Connection
	isUser := false

	// Id would be of format "WorkspaceHandle/ConnectionFolderId" or "OrganizationHandle/WorkspaceHandle/ConnectionFolderId"
	ids := strings.Split(d.Id(), "/")
	if len(ids) == 3 {
		orgHandle = ids[0]
		workspaceHandle = ids[1]
		connectionFolderId = ids[2]
	} else if len(ids) == 2 {
		workspaceHandle = ids[0]
		connectionFolderId = ids[1]
		isUser = true
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	_, new := d.GetChange("title")
	if new.(string) == "" {
		return diag.Errorf("title must be configured for a connection folder")
	}

	newTitle := new.(string)

	req := pipes.UpdateConnectionFolderRequest{
		Title: &newTitle,
	}
	if ok := d.HasChange("parent_id"); ok {
		if value, ok := d.GetOk("parent_id"); ok {
			req.SetParentId(value.(string))
		}
	}

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionFolderRead. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceConnectionFolders.Update(context.Background(), actorHandle, workspaceHandle, connectionFolderId).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceConnectionFolders.Update(context.Background(), orgHandle, workspaceHandle, connectionFolderId).Request(req).Execute()

	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceConnectionFolderUpdate. Update connection error: %v", decodeResponse(r))
	}

	d.Set("connection_folder_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("title", resp.Title)
	d.Set("type", resp.Type)
	d.Set("parent_id", resp.ParentId)
	d.Set("integration_resource_name", resp.IntegrationResourceName)
	d.Set("integration_resource_identifier", resp.IntegrationResourceIdentifier)
	d.Set("integration_resource_type", resp.IntegrationResourceType)
	d.Set("integration_resource_path", resp.IntegrationResourcePath)
	d.Set("managed_by_id", resp.ManagedById)
	d.Set("trunk", resp.Trunk)
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
	// The id would be of format "WorkspaceHandle/ConnectionFolderId" for user workspaces
	// or "OrganizationHandle/WorkspaceHandle/ConnectionFolderId" for org workspaces
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Id))

	}

	return diags
}

func resourceWorkspaceConnectionFolderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var connectionFolderId, orgHandle, workspaceHandle string
	isUser := false

	// Id would be of format "WorkspaceHandle/ConnectionFolderId" or "OrganizationHandle/WorkspaceHandle/ConnectionFolderId"
	ids := strings.Split(d.Id(), "/")
	if len(ids) == 3 {
		orgHandle = ids[0]
		workspaceHandle = ids[1]
		connectionFolderId = ids[2]
	} else if len(ids) == 2 {
		workspaceHandle = ids[0]
		connectionFolderId = ids[1]
		isUser = true
	}

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionFolderRead. getUserHandler error  %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceConnectionFolders.Delete(ctx, actorHandle, workspaceHandle, connectionFolderId).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceConnectionFolders.Delete(ctx, orgHandle, workspaceHandle, connectionFolderId).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceConnectionFolderDelete. Delete connection error:	%v", decodeResponse(r))
	}

	// clear the id to show we have deleted
	d.SetId("")

	return diags
}
