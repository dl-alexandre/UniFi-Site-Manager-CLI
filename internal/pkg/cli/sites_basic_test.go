package cli

import (
	"testing"

	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/api"
	"github.com/dl-alexandre/UniFi-Site-Manager-CLI/internal/pkg/mocks"
	"github.com/stretchr/testify/assert"
)

// ========== ListSitesCmd Tests ==========

func TestListSitesCmd_Success(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office", Description: "Main office", HostID: "host-1", Status: "active"},
			{ID: "site-2", Name: "Home", Description: "Home network", HostID: "host-2", Status: "active"},
		},
	}

	mockClient.On("ListSites", 50, "").Return(&sites, nil)

	cmd := &ListSitesCmd{PageSize: 50}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Office")
	assert.Contains(t, output, "Home")
	mockClient.AssertExpectations(t)
}

func TestListSitesCmd_WithSearch(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office", Description: "Main office", HostID: "host-1", Status: "active"},
			{ID: "site-2", Name: "Home", Description: "Home network", HostID: "host-2", Status: "active"},
			{ID: "site-3", Name: "Remote Office", Description: "Remote location", HostID: "host-3", Status: "active"},
		},
	}

	mockClient.On("ListSites", 50, "").Return(&sites, nil)

	cmd := &ListSitesCmd{PageSize: 50, Search: "office"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Office")
	assert.Contains(t, output, "Remote Office")
	// Home should be filtered out (case-insensitive search for "office")
	assert.NotContains(t, output, "Home")
	mockClient.AssertExpectations(t)
}

func TestListSitesCmd_JSONFormat(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)
	ctx.Format = "json"

	sites := api.SitesResponse{
		Data: []api.Site{
			{ID: "site-1", Name: "Office", Status: "active"},
		},
	}

	mockClient.On("ListSites", 50, "").Return(&sites, nil)

	cmd := &ListSitesCmd{PageSize: 50}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "site-1")
	assert.Contains(t, output, "Office")
	assert.Contains(t, output, "[")
	assert.Contains(t, output, "]")
	mockClient.AssertExpectations(t)
}

// ========== GetSiteCmd Tests ==========

func TestGetSiteCmd_Success(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site := api.SiteResponse{
		Data: api.Site{
			ID:          "site-123",
			Name:        "Test Site",
			Description: "A test site description",
			HostID:      "host-1",
			Status:      "active",
		},
	}

	mockClient.On("GetSite", "site-123").Return(&site, nil)

	cmd := &GetSiteCmd{SiteID: "site-123"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "site-123")
	assert.Contains(t, output, "Test Site")
	assert.Contains(t, output, "A test site description")
	mockClient.AssertExpectations(t)
}

func TestGetSiteCmd_EmptySiteID(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	cmd := &GetSiteCmd{SiteID: ""}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.IsType(t, &api.ValidationError{}, err)
	assert.Contains(t, err.Error(), "site ID is required")
}

func TestGetSiteCmd_NotFound(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	mockClient.On("GetSite", "nonexistent").Return(nil, &api.NotFoundError{Resource: "site"})

	cmd := &GetSiteCmd{SiteID: "nonexistent"}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.IsType(t, &api.NotFoundError{}, err)
	mockClient.AssertExpectations(t)
}

func TestGetSiteCmd_JSONFormat(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)
	ctx.Format = "json"

	site := api.SiteResponse{
		Data: api.Site{
			ID:   "site-123",
			Name: "Test Site",
		},
	}

	mockClient.On("GetSite", "site-123").Return(&site, nil)

	cmd := &GetSiteCmd{SiteID: "site-123"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "site-123")
	assert.Contains(t, output, "Test Site")
	assert.Contains(t, output, "{")
	mockClient.AssertExpectations(t)
}

// ========== CreateSiteCmd Tests ==========

func TestCreateSiteCmd_Success(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site := api.SiteResponse{
		Data: api.Site{
			ID:          "new-site-123",
			Name:        "New Site",
			Description: "New site description",
			Status:      "active",
		},
	}

	mockClient.On("CreateSite", api.CreateSiteRequest{Name: "New Site", Description: "New site description"}).Return(&site, nil)

	cmd := &CreateSiteCmd{Name: "New Site", Description: "New site description"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Site created successfully")
	assert.Contains(t, output, "new-site-123")
	assert.Contains(t, output, "New Site")
	mockClient.AssertExpectations(t)
}

func TestCreateSiteCmd_EmptyName(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	cmd := &CreateSiteCmd{Name: ""}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.IsType(t, &api.ValidationError{}, err)
	assert.Contains(t, err.Error(), "site name is required")
}

func TestCreateSiteCmd_APIError(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	mockClient.On("CreateSite", api.CreateSiteRequest{Name: "Test"}).Return(nil, &api.ValidationError{Message: "invalid request"})

	cmd := &CreateSiteCmd{Name: "Test"}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

// ========== UpdateSiteCmd Tests ==========

func TestUpdateSiteCmd_Success(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site := api.SiteResponse{
		Data: api.Site{
			ID:          "site-123",
			Name:        "Updated Name",
			Description: "Updated description",
			Status:      "active",
		},
	}

	mockClient.On("UpdateSite", "site-123", api.UpdateSiteRequest{Name: "Updated Name", Description: "Updated description"}).Return(&site, nil)

	cmd := &UpdateSiteCmd{SiteID: "site-123", Name: "Updated Name", Description: "Updated description"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Site updated successfully")
	assert.Contains(t, output, "Updated Name")
	mockClient.AssertExpectations(t)
}

func TestUpdateSiteCmd_EmptySiteID(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	cmd := &UpdateSiteCmd{SiteID: ""}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.IsType(t, &api.ValidationError{}, err)
}

func TestUpdateSiteCmd_PartialUpdate(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	site := api.SiteResponse{
		Data: api.Site{
			ID:   "site-123",
			Name: "Only Name Updated",
		},
	}

	mockClient.On("UpdateSite", "site-123", api.UpdateSiteRequest{Name: "Only Name Updated"}).Return(&site, nil)

	cmd := &UpdateSiteCmd{SiteID: "site-123", Name: "Only Name Updated", Description: ""}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Only Name Updated")
	mockClient.AssertExpectations(t)
}

// ========== DeleteSiteCmd Tests ==========

func TestDeleteSiteCmd_Success(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	mockClient.On("DeleteSite", "site-123").Return(nil)

	cmd := &DeleteSiteCmd{SiteID: "site-123", Force: true}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "deleted successfully")
	mockClient.AssertExpectations(t)
}

func TestDeleteSiteCmd_EmptySiteID(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	cmd := &DeleteSiteCmd{SiteID: ""}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.IsType(t, &api.ValidationError{}, err)
}

func TestDeleteSiteCmd_NotFound(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	mockClient.On("DeleteSite", "nonexistent").Return(&api.NotFoundError{Resource: "site"})

	cmd := &DeleteSiteCmd{SiteID: "nonexistent", Force: true}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.IsType(t, &api.NotFoundError{}, err)
	mockClient.AssertExpectations(t)
}

// ========== SiteHealthCmd Tests ==========

func TestSiteHealthCmd_Success(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	health := api.HealthResponse{
		Data: []api.HealthStatus{
			{Subsystem: "wifi", Status: "ok", NumAdopted: 5, NumClient: 25},
			{Subsystem: "wan", Status: "ok", Latency: 15},
		},
	}

	mockClient.On("GetSiteHealth", "site-123").Return(&health, nil)

	cmd := &SiteHealthCmd{SiteID: "site-123"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "wifi")
	assert.Contains(t, output, "wan")
	assert.Contains(t, output, "ok")
	mockClient.AssertExpectations(t)
}

func TestSiteHealthCmd_EmptySiteID(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	cmd := &SiteHealthCmd{SiteID: ""}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.IsType(t, &api.ValidationError{}, err)
}

func TestSiteHealthCmd_JSONFormat(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)
	ctx.Format = "json"

	health := api.HealthResponse{
		Data: []api.HealthStatus{
			{Subsystem: "wifi", Status: "ok", NumAdopted: 5},
		},
	}

	mockClient.On("GetSiteHealth", "site-123").Return(&health, nil)

	cmd := &SiteHealthCmd{SiteID: "site-123"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "wifi")
	assert.Contains(t, output, "[")
	mockClient.AssertExpectations(t)
}

// ========== SiteStatsCmd Tests ==========

func TestSiteStatsCmd_Success(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	stats := api.PerformanceResponse{
		Data: api.PerformanceStats{
			SiteID:     "site-123",
			Period:     "day",
			RxBytes:    1073741824,
			TxBytes:    2147483648,
			NumClients: 50,
			NumDevices: 10,
		},
	}

	mockClient.On("GetSiteStats", "site-123", "day").Return(&stats, nil)

	cmd := &SiteStatsCmd{SiteID: "site-123", Period: "day"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "site-123")
	assert.Contains(t, output, "day")
	mockClient.AssertExpectations(t)
}

func TestSiteStatsCmd_EmptySiteID(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	cmd := &SiteStatsCmd{SiteID: ""}

	err := cmd.Run(ctx)
	assert.Error(t, err)
	assert.IsType(t, &api.ValidationError{}, err)
}

func TestSiteStatsCmd_DefaultPeriod(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)

	stats := api.PerformanceResponse{
		Data: api.PerformanceStats{
			SiteID: "site-123",
			Period: "day",
		},
	}

	mockClient.On("GetSiteStats", "site-123", "day").Return(&stats, nil)

	cmd := &SiteStatsCmd{SiteID: "site-123", Period: "day"}

	_ = captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	mockClient.AssertExpectations(t)
}

func TestSiteStatsCmd_JSONFormat(t *testing.T) {
	mockClient := mocks.NewSiteManager(t)
	ctx := createTestContext(mockClient)
	ctx.Format = "json"

	stats := api.PerformanceResponse{
		Data: api.PerformanceStats{
			SiteID: "site-123",
			Period: "week",
		},
	}

	mockClient.On("GetSiteStats", "site-123", "week").Return(&stats, nil)

	cmd := &SiteStatsCmd{SiteID: "site-123", Period: "week"}

	output := captureStdout(func() {
		err := cmd.Run(ctx)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "{")
	assert.Contains(t, output, "site-123")
	mockClient.AssertExpectations(t)
}
