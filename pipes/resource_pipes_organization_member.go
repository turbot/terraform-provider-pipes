package pipes

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/turbot/go-kit/types"
	pipes "github.com/turbot/pipes-sdk-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOrganizationMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationMemberCreate,
		ReadContext:   resourceOrganizationMemberRead,
		DeleteContext: resourceOrganizationMemberDelete,
		UpdateContext: resourceOrganizationMemberUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_handle": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"email"},
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"user_handle"},
			},
			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organization_member_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"display_name": {
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
		},
	}
}

// CRUD functions

func resourceOrganizationMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	client := meta.(*PipesClient)

	// Get the organization
	orgIdentifier := d.Get("organization").(string)

	// get details of the organization
	org, r, err := client.APIClient.Orgs.Get(context.Background(), orgIdentifier).Execute()
	if err != nil {
		return diag.Errorf("error reading organization %s: %s", orgIdentifier, decodeResponse(r))
	}

	var orgMember pipes.OrgUser

	// If the organization belongs to the primary tenant, we procced with logic to invite a user to the organization
	if org.TenantId == PipesTenantId {
		// Create request
		req := pipes.InviteOrgUserRequest{
			Role: d.Get("role").(string),
		}

		if value, ok := d.GetOk("user_handle"); ok {
			req.Handle = types.String(value.(string))
		}
		if value, ok := d.GetOk("email"); ok {
			req.Email = types.String(value.(string))
		}

		// Return if both handle and email are empty
		if req.Handle == nil && req.Email == nil {
			return diag.Errorf("either 'user_handle' or 'email' must be set in resource config")
		}

		// Invite requested member
		orgMember, r, err = client.APIClient.OrgMembers.Invite(ctx, org.Handle).Request(req).Execute()
		if err != nil {
			return diag.Errorf("error inviting member: %s", decodeResponse(r))
		}
		log.Printf("\n[DEBUG] Member invited: %v", orgMember)
	} else {
		// else the organization belongs to a custom tenant and we need to invoke logic to simply add the user to the organization
		// Create request
		req := pipes.CreateOrgUserRequest{
			Role: d.Get("role").(string),
		}

		if value, ok := d.GetOk("user_handle"); ok {
			req.Handle = value.(string)
		}

		// Return if user_handle is empty
		if req.Handle == "" {
			return diag.Errorf("'user_handle' must be set in resource config")
		}

		// Add requested member to the organization
		orgMember, r, err = client.APIClient.OrgMembers.Create(ctx, org.Handle).Request(req).Execute()
		if err != nil {
			return diag.Errorf("error inviting member: %s", decodeResponse(r))
		}
		log.Printf("\n[DEBUG] Member invited: %v", orgMember)
	}

	// Set property values
	d.SetId(fmt.Sprintf("%s/%s", org.Handle, orgMember.UserHandle))
	d.Set("user_handle", orgMember.UserHandle)
	d.Set("created_at", orgMember.CreatedAt)
	d.Set("organization_member_id", orgMember.Id)
	d.Set("organization_id", orgMember.OrgId)
	d.Set("role", orgMember.Role)
	d.Set("scope", orgMember.Scope)
	d.Set("status", orgMember.Status)
	d.Set("updated_at", orgMember.UpdatedAt)
	d.Set("user_id", orgMember.UserId)
	d.Set("version_id", orgMember.VersionId)
	if orgMember.CreatedBy != nil {
		d.Set("created_by", orgMember.CreatedBy.Handle)
	}
	if orgMember.UpdatedBy != nil {
		d.Set("updated_by", orgMember.UpdatedBy.Handle)
	}

	if orgMember.User != nil {
		d.Set("display_name", orgMember.User.DisplayName)
	}

	return diags
}

func resourceOrganizationMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	id := d.Id()
	// For backward-compatibility, we see whether the id contains : or /
	separator := "/"
	if strings.Contains(id, ":") {
		separator = ":"
	}
	idParts := strings.Split(id, separator)
	if len(idParts) < 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization_handle>/<user_handle>", id)
	}
	org := idParts[0]

	if strings.Contains(idParts[1], "@") {
		return diag.Errorf("invalid user_handle. Please provide valid user_handle to import")
	}
	userHandle := idParts[1]

	resp, r, err := client.APIClient.OrgMembers.Get(context.Background(), org, userHandle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			log.Printf("\n[WARN] Member (%s) not found", userHandle)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading %s:%s.\nerr: %s", org, userHandle, decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Organization Member received: %s", id)

	if separator == ":" {
		d.SetId(strings.ReplaceAll(id, ":", "/"))
	}
	d.Set("user_handle", resp.UserHandle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("organization_member_id", resp.Id)
	d.Set("organization_id", resp.OrgId)
	d.Set("role", resp.Role)
	d.Set("scope", resp.Scope)
	d.Set("status", resp.Status)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("user_id", resp.UserId)
	d.Set("version_id", resp.VersionId)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	if resp.User != nil {
		d.Set("display_name", resp.User.DisplayName)
	}

	return diags
}

func resourceOrganizationMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Get the organization
	org := d.Get("organization").(string)

	userHandle := d.Get("user_handle").(string)
	role := d.Get("role").(string)

	// Create request
	req := pipes.UpdateOrgUserRequest{
		Role: role,
	}

	log.Printf("\n[DEBUG] Updating membership: '%s/%s'", org, userHandle)

	resp, r, err := client.APIClient.OrgMembers.Update(context.Background(), org, userHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error updating membership: %s", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Membership updated: %s/%s", org, resp.UserHandle)

	// Update state file
	id := fmt.Sprintf("%s/%s", org, resp.UserHandle)
	d.SetId(id)
	d.Set("user_handle", resp.UserHandle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("organization_member_id", resp.Id)
	d.Set("organization_id", resp.OrgId)
	d.Set("role", resp.Role)
	d.Set("scope", resp.Scope)
	d.Set("status", resp.Status)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("user_id", resp.UserId)
	d.Set("version_id", resp.VersionId)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	if resp.User != nil {
		d.Set("display_name", resp.User.DisplayName)
	}

	return diags
}

func resourceOrganizationMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	id := d.Id()
	// For backward-compatibility, we see whether the id contains : or /
	separator := "/"
	if strings.Contains(id, ":") {
		separator = ":"
	}
	idParts := strings.Split(id, separator)
	if len(idParts) < 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <organization_handle>/<user_handle>", id)
	}
	org := idParts[0]

	log.Printf("\n[DEBUG] Removing membership: %s", id)

	_, r, err := client.APIClient.OrgMembers.Delete(context.Background(), org, idParts[1]).Execute()
	if err != nil {
		return diag.Errorf("error removing membership %s: %s", id, decodeResponse(r))
	}
	d.SetId("")

	return diags
}
