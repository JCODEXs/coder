package codersdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"github.com/coder/coder/coderd/database"
)

// Organization is the JSON representation of a Coder organization.
type Organization struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	Name      string    `json:"name" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type OrganizationMember struct {
	UserID         uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateTemplateVersionRequest enables callers to create a new Template Version.
type CreateTemplateVersionRequest struct {
	// TemplateID optionally associates a version with a template.
	TemplateID uuid.UUID `json:"template_id"`

	StorageMethod database.ProvisionerStorageMethod `json:"storage_method" validate:"oneof=file,required"`
	StorageSource string                            `json:"storage_source" validate:"required"`
	Provisioner   database.ProvisionerType          `json:"provisioner" validate:"oneof=terraform echo,required"`
	// ParameterValues allows for additional parameters to be provided
	// during the dry-run provision stage.
	ParameterValues []CreateParameterRequest `json:"parameter_values"`
}

// CreateTemplateRequest provides options when creating a template.
type CreateTemplateRequest struct {
	Name string `json:"name" validate:"username,required"`

	// VersionID is an in-progress or completed job to use as
	// an initial version of the template.
	//
	// This is required on creation to enable a user-flow of validating a
	// template works. There is no reason the data-model cannot support
	// empty templates, but it doesn't make sense for users.
	VersionID       uuid.UUID                `json:"template_version_id" validate:"required"`
	ParameterValues []CreateParameterRequest `json:"parameter_values"`
}

func (c *Client) Organization(ctx context.Context, id uuid.UUID) (Organization, error) {
	res, err := c.request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/organizations/%s", id.String()), nil)
	if err != nil {
		return Organization{}, xerrors.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Organization{}, readBodyAsError(res)
	}

	var organization Organization
	return organization, json.NewDecoder(res.Body).Decode(&organization)
}

// OrganizationMemberByUsername fetches a member of an organization by name.
func (c *Client) OrganizationMemberByUsername(ctx context.Context, organizationID uuid.UUID, username string) (OrganizationMember, error) {
	res, err := c.request(ctx, http.MethodGet,
		fmt.Sprintf("/api/v2/organizations/%s/members/%s", organizationID.String(), username),
		nil,
	)
	if err != nil {
		return OrganizationMember{}, xerrors.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return OrganizationMember{}, readBodyAsError(res)
	}

	var member OrganizationMember
	return member, json.NewDecoder(res.Body).Decode(&member)
}

// ProvisionerDaemonsByOrganization returns provisioner daemons available for an organization.
func (c *Client) ProvisionerDaemonsByOrganization(ctx context.Context, organizationID uuid.UUID) ([]ProvisionerDaemon, error) {
	res, err := c.request(ctx, http.MethodGet,
		fmt.Sprintf("/api/v2/organizations/%s/provisionerdaemons", organizationID.String()),
		nil,
	)
	if err != nil {
		return nil, xerrors.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, readBodyAsError(res)
	}

	var daemons []ProvisionerDaemon
	return daemons, json.NewDecoder(res.Body).Decode(&daemons)
}

// CreateTemplateVersion processes source-code and optionally associates the version with a template.
// Executing without a template is useful for validating source-code.
func (c *Client) CreateTemplateVersion(ctx context.Context, organizationID uuid.UUID, req CreateTemplateVersionRequest) (TemplateVersion, error) {
	res, err := c.request(ctx, http.MethodPost,
		fmt.Sprintf("/api/v2/organizations/%s/templateversions", organizationID.String()),
		req,
	)
	if err != nil {
		return TemplateVersion{}, xerrors.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return TemplateVersion{}, readBodyAsError(res)
	}

	var templateVersion TemplateVersion
	return templateVersion, json.NewDecoder(res.Body).Decode(&templateVersion)
}

// CreateTemplate creates a new template inside an organization.
func (c *Client) CreateTemplate(ctx context.Context, organizationID uuid.UUID, request CreateTemplateRequest) (Template, error) {
	res, err := c.request(ctx, http.MethodPost,
		fmt.Sprintf("/api/v2/organizations/%s/templates", organizationID.String()),
		request,
	)
	if err != nil {
		return Template{}, xerrors.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return Template{}, readBodyAsError(res)
	}

	var template Template
	return template, json.NewDecoder(res.Body).Decode(&template)
}

// TemplatesByOrganization lists all templates inside of an organization.
func (c *Client) TemplatesByOrganization(ctx context.Context, organizationID uuid.UUID) ([]Template, error) {
	res, err := c.request(ctx, http.MethodGet,
		fmt.Sprintf("/api/v2/organizations/%s/templates", organizationID.String()),
		nil,
	)
	if err != nil {
		return nil, xerrors.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, readBodyAsError(res)
	}

	var templates []Template
	return templates, json.NewDecoder(res.Body).Decode(&templates)
}

// TemplateByName finds a template inside the organization provided with a case-insensitive name.
func (c *Client) TemplateByName(ctx context.Context, organizationID uuid.UUID, name string) (Template, error) {
	res, err := c.request(ctx, http.MethodGet,
		fmt.Sprintf("/api/v2/organizations/%s/templates/%s", organizationID.String(), name),
		nil,
	)
	if err != nil {
		return Template{}, xerrors.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Template{}, readBodyAsError(res)
	}

	var template Template
	return template, json.NewDecoder(res.Body).Decode(&template)
}
