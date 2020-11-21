package googlemaps

import (
	"context"

	"googlemaps.github.io/maps"
)

type PlaceSearchResult struct {
	FormattedAddress string
	Location         LatLng
}

func (c Client) SearchPlacesWithText(query string) ([]PlaceSearchResult, error) {
	cli, err := maps.NewClient(maps.WithAPIKey(c.APIKey))
	if err != nil {
		return nil, err
	}

	results, err := cli.FindPlaceFromText(context.Background(), &maps.FindPlaceFromTextRequest{
		Input:     query,
		InputType: maps.FindPlaceFromTextInputTypeTextQuery,
		Fields: []maps.PlaceSearchFieldMask{
			maps.PlaceSearchFieldMaskFormattedAddress,
			maps.PlaceSearchFieldMaskGeometryLocationLat,
			maps.PlaceSearchFieldMaskGeometryLocationLng,
		},
	})
	if err != nil {
		return nil, err
	}

	PlaceSearchResults := make([]PlaceSearchResult, 0)
	for _, res := range results.Candidates {
		PlaceSearchResults = append(PlaceSearchResults, PlaceSearchResult{
			FormattedAddress: res.FormattedAddress,
			Location: LatLng{
				Latitude:  res.Geometry.Location.Lat,
				Longitude: res.Geometry.Location.Lng,
			},
		})
	}

	return PlaceSearchResults, nil
}
