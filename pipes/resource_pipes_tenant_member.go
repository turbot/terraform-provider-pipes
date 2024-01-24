package pipes

import (
	"context"
	"fmt"
	"log"
	"strings"

	pipes "github.com/turbot/pipes-sdk-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTenantMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTenantMemberCreate,
		ReadContext:   resourceTenantMemberRead,
		DeleteContext: resourceTenantMemberDelete,
		UpdateContext: resourceTenantMemberUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"tenant_member_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_handle": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
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
		},
	}
}

func resourceTenantMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	client := meta.(*PipesClient)

	// Create request
	req := pipes.InviteTenantUserRequest{
		Email: d.Get("email").(string),
		Role:  d.Get("role").(string),
	}

	// Get the tenant handle
	tenantHandle := d.Get("tenant_handle").(string)

	// Invite requested member
	tenantMember, r, err := client.APIClient.TenantMembers.Invite(ctx, tenantHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error inviting member: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Member invited: %v", tenantMember)

	// Get details of the invited member
	tenantUser, r, err := client.APIClient.Identities.Get(ctx, tenantMember.UserId).Execute()
	if err != nil {
		return diag.Errorf("error getting invited member details: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Member details: %v", tenantUser)

	// Set property values
	d.SetId(fmt.Sprintf("%s/%s", tenantHandle, tenantMember.UserId))
	d.Set("tenant_member_id", tenantMember.Id)
	d.Set("tenant_id", tenantMember.TenantId)
	d.Set("user_id", tenantMember.UserId)
	d.Set("user_handle", tenantUser.Handle)
	d.Set("email", tenantMember.Email)
	d.Set("role", tenantMember.Role)
	d.Set("status", tenantMember.Status)
	d.Set("created_at", tenantMember.CreatedAt)
	d.Set("updated_at", tenantMember.UpdatedAt)
	d.Set("version_id", tenantMember.VersionId)
	if tenantMember.CreatedBy != nil {
		d.Set("created_by", tenantMember.CreatedBy.Handle)
	}
	if tenantMember.UpdatedBy != nil {
		d.Set("updated_by", tenantMember.UpdatedBy.Handle)
	}

	return diags
}

func resourceTenantMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	id := d.Id()
	separator := "/"
	idParts := strings.Split(id, separator)
	if len(idParts) < 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <tenant_handle>/<user_id>", id)
	}
	tenantHandle := idParts[0]

	if strings.Contains(idParts[1], "@") {
		return diag.Errorf("invalid user_id. Please provide valid user_id to import")
	}
	userHandle := idParts[1]

	tenantMember, r, err := client.APIClient.TenantMembers.Get(context.Background(), tenantHandle, userHandle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Member (%s) not found", userHandle)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading %s:%s.\nerr: %s", tenantHandle, userHandle, decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Tenant Member received: %s", id)

	// Get details of the invited member
	tenantUser, r, err := client.APIClient.Identities.Get(ctx, tenantMember.UserId).Execute()
	if err != nil {
		return diag.Errorf("error getting invited member details: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Member details: %v", tenantUser)

	// Set property values
	d.SetId(fmt.Sprintf("%s/%s", tenantHandle, tenantMember.UserId))
	d.Set("tenant_member_id", tenantMember.Id)
	d.Set("tenant_id", tenantMember.TenantId)
	d.Set("user_id", tenantMember.UserId)
	d.Set("user_handle", tenantUser.Handle)
	d.Set("email", tenantMember.Email)
	d.Set("role", tenantMember.Role)
	d.Set("status", tenantMember.Status)
	d.Set("created_at", tenantMember.CreatedAt)
	d.Set("updated_at", tenantMember.UpdatedAt)
	d.Set("version_id", tenantMember.VersionId)
	if tenantMember.CreatedBy != nil {
		d.Set("created_by", tenantMember.CreatedBy.Handle)
	}
	if tenantMember.UpdatedBy != nil {
		d.Set("updated_by", tenantMember.UpdatedBy.Handle)
	}

	return diags
}

func resourceTenantMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	tenantHandle := d.Get("tenant_handle").(string)
	userId := d.Get("user_id").(string)
	role := d.Get("role").(string)

	// Create request
	req := pipes.UpdateTenantUserRequest{
		Role: &role,
	}

	log.Printf("\n[DEBUG] Updating membership: '%s/%s'", tenantHandle, userId)

	tenantMember, r, err := client.APIClient.TenantMembers.Update(context.Background(), tenantHandle, userId).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error updating membership: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Membership updated: %s/%s", tenantHandle, userId)

	// Get details of the invited member
	tenantUser, r, err := client.APIClient.Identities.Get(ctx, tenantMember.UserId).Execute()
	if err != nil {
		return diag.Errorf("error getting invited member details: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Member details: %v", tenantUser)

	// Set property values
	d.SetId(fmt.Sprintf("%s/%s", tenantHandle, tenantMember.User.Handle))
	d.Set("tenant_member_id", tenantMember.Id)
	d.Set("tenant_id", tenantMember.TenantId)
	d.Set("user_id", tenantMember.UserId)
	d.Set("user_handle", tenantUser.Handle)
	d.Set("email", tenantMember.Email)
	d.Set("role", tenantMember.Role)
	d.Set("status", tenantMember.Status)
	d.Set("created_at", tenantMember.CreatedAt)
	d.Set("updated_at", tenantMember.UpdatedAt)
	d.Set("version_id", tenantMember.VersionId)
	if tenantMember.CreatedBy != nil {
		d.Set("created_by", tenantMember.CreatedBy.Handle)
	}
	if tenantMember.UpdatedBy != nil {
		d.Set("updated_by", tenantMember.UpdatedBy.Handle)
	}

	return diags
}

func resourceTenantMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	id := d.Id()
	separator := "/"
	idParts := strings.Split(id, separator)
	if len(idParts) < 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <tenant_handle>/<user_id>", id)
	}
	tenantHandle := idParts[0]
	userHandle := idParts[1]

	log.Printf("\n[DEBUG] Removing membership: %s", id)

	_, r, err := client.APIClient.TenantMembers.Delete(context.Background(), tenantHandle, userHandle).Execute()
	if err != nil {
		return diag.Errorf("error removing membership %s: %s", id, decodeResponse(r))
	}
	d.SetId("")

	return diags
}
