package workers

import (
	"github.com/sentinel-official/sentinel-go-sdk/libs/geoip"

	"github.com/sentinel-official/dvpn-node/context"
)

// handleGeoIPLocation  fetches the geo IP location using the provided client and sets it in the given context.
func handleGeoIPLocation(ctx *context.Context, client geoip.Client) func() error {
	return func() error {
		// Fetch the geo IP location.
		location, err := client.Get("")
		if err != nil {
			return err
		}

		// Set the fetched location in the context.
		ctx.SetLocation(location)

		return nil
	}
}

// onErrorBestRPCEndpoint returns a function to handle errors.
func onErrorGeoIPLocation(_ *context.Context) func(error) bool {
	return func(err error) bool {
		// Do nothing on error
		return false
	}
}