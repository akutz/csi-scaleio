package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thecodeteam/gocsi/csi"

	"github.com/thecodeteam/csi-scaleio/service"
)

func TestPluginInfo(t *testing.T) {

	ctx := context.Background()

	gclient, stop := startServer(ctx, t)
	defer stop()

	client := csi.NewIdentityClient(gclient)

	info, err := client.GetPluginInfo(ctx,
		&csi.GetPluginInfoRequest{
			Version: service.SupportedVersions[0],
		})

	assert.NoError(t, err)
	assert.Equal(t, info.GetName(), service.Name)
	assert.Equal(t, info.GetVendorVersion(), service.VendorVersion)
}

func TestGetSupportedVersions(t *testing.T) {

	ctx := context.Background()

	gclient, stop := startServer(ctx, t)
	defer stop()

	client := csi.NewIdentityClient(gclient)

	vers, err := client.GetSupportedVersions(ctx,
		&csi.GetSupportedVersionsRequest{})

	assert.NoError(t, err)
	assert.NotEmpty(t, vers.GetSupportedVersions())
}
