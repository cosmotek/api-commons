package googlemaps

import (
	"context"
	"errors"
	"time"

	"github.com/google/go-cmp/cmp"
	"googlemaps.github.io/maps"
)

var (
	// ErrNoValidRoutes is used by the trip calculation methods
	// and should be returned when the API call returns no data.
	ErrNoValidRoutes = errors.New("no valid routes found")

	// ErrNoValidLocations is used by the geocoding methods
	// and should be returned when the API call returns no data.
	ErrNoValidLocations = errors.New("no valid locations found")
)

// Client is a Google Maps API client instance
type Client struct {
	APIKey string
}

// Trip is a trip overview that contains trip duration (time.Duration)
// and distance (in miles).
type Trip struct {
	Duration time.Duration

	// Distance in miles
	Distance int
}

// TimeZone ...
type TimeZone struct {
	DstOffset    int
	RawOffset    int
	TimeZoneID   string
	TimeZoneName string
}

// GetTripFromAddresses makes a Google Maps API call to calculate
// the distance and duration of a trip using a provided postal address.
func (c Client) GetTripFromAddresses(originAddress, destinationAddress string) (Trip, error) {
	cli, err := maps.NewClient(maps.WithAPIKey(c.APIKey))
	if err != nil {
		return Trip{}, err
	}

	routes, _, err := cli.Directions(context.Background(), &maps.DirectionsRequest{
		Origin:      originAddress,
		Destination: destinationAddress,
	})
	if err != nil {
		return Trip{}, err
	}

	if len(routes) == 0 {
		return Trip{}, ErrNoValidRoutes
	}

	return Trip{
		Duration: routes[0].Legs[0].Duration,
		Distance: routes[0].Legs[0].Distance.Meters / 1609, // Meters to Miles
	}, nil
}

// GetTripFromCoords makes a Google Maps API call to calculate
// the distance and duration of a trip using a provided lat and lng.
func (c Client) GetTripFromCoords(origin, destination LatLng) (Trip, error) {
	cli, err := maps.NewClient(maps.WithAPIKey(c.APIKey))
	if err != nil {
		return Trip{}, err
	}

	originCoords := maps.LatLng{
		Lat: origin.Latitude,
		Lng: origin.Longitude,
	}

	destinationCoords := maps.LatLng{
		Lat: destination.Latitude,
		Lng: destination.Longitude,
	}

	routes, _, err := cli.Directions(context.Background(), &maps.DirectionsRequest{
		Origin:      originCoords.String(),
		Destination: destinationCoords.String(),
	})
	if err != nil {
		return Trip{}, err
	}

	if len(routes) == 0 {
		return Trip{}, ErrNoValidRoutes
	}

	return Trip{
		Duration: routes[0].Legs[0].Duration,
		Distance: routes[0].Legs[0].Distance.Meters / 1609, // Meters to Miles
	}, nil
}

// LatLng is a object containing the data used to represent a latlng location
type LatLng struct {
	Latitude  float64
	Longitude float64
}

// GetGeocodedAddress converts a postal address to a LatLng using
// the Google Maps API
func (c Client) GetGeocodedAddress(address string) (LatLng, string, error) {
	cli, err := maps.NewClient(maps.WithAPIKey(c.APIKey))
	if err != nil {
		return LatLng{}, "", err
	}

	results, err := cli.Geocode(context.Background(), &maps.GeocodingRequest{
		Address: address,
	})
	if err != nil {
		return LatLng{}, "", err
	}

	if len(results) == 0 {
		return LatLng{}, "", ErrNoValidLocations
	}

	lat := results[0].Geometry.Location.Lat
	lng := results[0].Geometry.Location.Lng

	return LatLng{
		Latitude:  lat,
		Longitude: lng,
	}, results[0].FormattedAddress, nil
}

// Place is a object containing fields used to describe
// a location on Google Maps
type Place struct {
	PlaceID           string
	PhotoReferenceIDs []string
}

// GetPlaceDetails fetches and returns place details from
// the Google Maps API using the provided postal address.
func (c Client) GetPlaceDetails(address string) (Place, error) {
	cli, err := maps.NewClient(maps.WithAPIKey(c.APIKey))
	if err != nil {
		return Place{}, err
	}

	results, err := cli.Geocode(context.Background(), &maps.GeocodingRequest{
		Address: address,
	})
	if err != nil {
		return Place{}, err
	}

	if len(results) == 0 {
		return Place{}, ErrNoValidLocations
	}

	placeID := results[0].PlaceID
	placeResult, err := cli.PlaceDetails(context.Background(), &maps.PlaceDetailsRequest{
		PlaceID: placeID,
	})
	if err != nil {
		return Place{}, err
	}

	// TODO use placeResult.Geometry to show the address boundries
	// TODO use hours from placeResult?

	photosReferences := make([]string, 0)
	for _, photo := range placeResult.Photos {
		photosReferences = append(photosReferences, photo.PhotoReference)
	}

	return Place{
		PlaceID:           placeID,
		PhotoReferenceIDs: photosReferences,
	}, nil
}

// AddressComponents is an object returned from geocoding requests used
// to map from coords to a postal address with components.
type AddressComponents struct {
	FullAddress string

	StreetNumber string

	// Route is the body of the address e.g "E College Ave."
	Route      string
	RouteShort string

	Neighborhood      string
	NeighborhoodShort string

	City      string
	CityShort string

	State      string
	StateShort string

	Country      string
	CountryShort string
}

// GetPostalAddressFromCoords converts a LatLng to a PostalAddress using
// the Google Maps API
func (c Client) GetPostalAddressFromCoords(input LatLng) (AddressComponents, error) {
	cli, err := maps.NewClient(maps.WithAPIKey(c.APIKey))
	if err != nil {
		return AddressComponents{}, err
	}

	results, err := cli.Geocode(context.Background(), &maps.GeocodingRequest{
		LatLng: &maps.LatLng{
			Lat: input.Latitude,
			Lng: input.Longitude,
		},
	})
	if err != nil {
		return AddressComponents{}, err
	}

	if len(results) == 0 {
		return AddressComponents{}, ErrNoValidLocations
	}

	addrComp := AddressComponents{
		FullAddress: results[0].FormattedAddress,
	}

	for _, comp := range results[0].AddressComponents {
		if cmp.Equal(comp.Types, []string{"street_number"}) {
			addrComp.StreetNumber = comp.LongName
		} else if cmp.Equal(comp.Types, []string{"route"}) {
			addrComp.Route = comp.LongName
			addrComp.RouteShort = comp.ShortName
		} else if cmp.Equal(comp.Types, []string{"neighborhood", "political"}) {
			addrComp.Neighborhood = comp.LongName
			addrComp.NeighborhoodShort = comp.ShortName
		} else if cmp.Equal(comp.Types, []string{"locality", "political"}) {
			addrComp.City = comp.LongName
			addrComp.CityShort = comp.ShortName
		} else if cmp.Equal(comp.Types, []string{"administrative_area_level_1", "political"}) {
			addrComp.State = comp.LongName
			addrComp.StateShort = comp.ShortName
		} else if cmp.Equal(comp.Types, []string{"country", "political"}) {
			addrComp.Country = comp.LongName
			addrComp.CountryShort = comp.ShortName
		}
	}

	return addrComp, nil
}

// GetTimezoneFromLocation makes a Google Maps API call get the
// timezone of the location
func (c Client) GetTimezoneFromLocation(location LatLng) (TimeZone, error) {
	cli, err := maps.NewClient(maps.WithAPIKey(c.APIKey))
	if err != nil {
		return TimeZone{}, err
	}

	timezone, err := cli.Timezone(context.Background(), &maps.TimezoneRequest{
		Location: &maps.LatLng{
			Lat: location.Latitude,
			Lng: location.Longitude,
		},
		Timestamp: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC), // date during DST
	})
	if err != nil {
		return TimeZone{}, err
	}

	return TimeZone{
		DstOffset:    timezone.DstOffset,
		RawOffset:    timezone.RawOffset,
		TimeZoneID:   timezone.TimeZoneID,
		TimeZoneName: timezone.TimeZoneName,
	}, nil
}
