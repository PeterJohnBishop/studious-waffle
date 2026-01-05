package server

import (
	"context"
	"io"
	"net/http"
	"sort"
	"studious-waffle/server/protodata"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

const rtdAlerts = "https://www.rtd-denver.com/files/gtfs-rt/Alerts.pb"
const rtdTripUpdates = "https://www.rtd-denver.com/files/gtfs-rt/TripUpdate.pb"
const rtdVehiclePosition = "https://www.rtd-denver.com/files/gtfs-rt/VehiclePosition.pb"

func fetchRawGTFS(url string) (*gtfs.FeedMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	feed := &gtfs.FeedMessage{}
	if err := proto.Unmarshal(data, feed); err != nil {
		return nil, err
	}
	return feed, nil
}

func FetchAlerts() ([]*protodata.AlertEntityProto, error) {
    rawFeed, err := fetchRawGTFS(rtdAlerts)
    if err != nil {
        return nil, err
    }

    var results []*protodata.AlertEntityProto

    for _, entity := range rawFeed.Entity {
        a := entity.GetAlert()
        if a == nil {
            continue
        }

        var activePeriods []*protodata.ActivePeriodProto
        for _, p := range a.GetActivePeriod() {
            activePeriods = append(activePeriods, &protodata.ActivePeriodProto{
                Start: proto.Int64(int64(p.GetStart())),
                End:   proto.Int64(int64(p.GetEnd())),
            })
        }

        var informedEntities []*protodata.InformedEntityProto
        for _, ie := range a.GetInformedEntity() {
            informedEntities = append(informedEntities, &protodata.InformedEntityProto{
                AgencyId:  proto.String(ie.GetAgencyId()),
                RouteId:   proto.String(ie.GetRouteId()),
                RouteType: proto.Int32(int32(ie.GetRouteType())),
                StopId:    proto.String(ie.GetStopId()),
            })
        }

        headerTranslations := mapTranslations(a.GetHeaderText())
        descTranslations := mapTranslations(a.GetDescriptionText())

        results = append(results, &protodata.AlertEntityProto{
            Id: proto.String(entity.GetId()),
            Alert: &protodata.AlertProto{
                ActivePeriod:    activePeriods,
                InformedEntity:  informedEntities,
                Cause:           proto.Int32(int32(a.GetCause())),
                Effect:          proto.Int32(int32(a.GetEffect())),
                HeaderText:      &protodata.TranslatedStringProto{Translation: headerTranslations},
                DescriptionText: &protodata.TranslatedStringProto{Translation: descTranslations},
            },
        })
    }
    return results, nil
}

func mapTranslations(rawText *gtfs.TranslatedString) []*protodata.TranslationProto {
    var translations []*protodata.TranslationProto
    if rawText == nil {
        return translations
    }
    for _, t := range rawText.GetTranslation() {
        translations = append(translations, &protodata.TranslationProto{
            Text:     proto.String(t.GetText()),
            Language: proto.String(t.GetLanguage()),
        })
    }
    return translations
}

func FetchTripUpdates() ([]*protodata.TripUpdateEntityProto, error) {
    rawFeed, err := fetchRawGTFS(rtdTripUpdates)
    if err != nil {
        return nil, err
    }

    var results []*protodata.TripUpdateEntityProto

    for _, entity := range rawFeed.Entity {
        tu := entity.GetTripUpdate()
        if tu == nil {
            continue
        }

        var stopUpdates []*protodata.StopTimeUpdateProto
        for _, stu := range tu.GetStopTimeUpdate() {
            stopUpdates = append(stopUpdates, &protodata.StopTimeUpdateProto{
                StopSequence: proto.Int32(int32(stu.GetStopSequence())),
                StopId:       proto.String(stu.GetStopId()),
                Arrival: &protodata.StopTimeEventProto{
                    Time: proto.Int64(int64(stu.GetArrival().GetTime())),
                },
                Departure: &protodata.StopTimeEventProto{
                    Time: proto.Int64(int64(stu.GetDeparture().GetTime())),
                },
                ScheduleRelationship: proto.Int32(int32(stu.GetScheduleRelationship())),
            })
        }

        results = append(results, &protodata.TripUpdateEntityProto{
            Id: proto.String(entity.GetId()),
            TripUpdate: &protodata.TripUpdateProto{
                Trip: &protodata.TripDescriptorProto{
                    TripId:               proto.String(tu.GetTrip().GetTripId()),
                    RouteId:              proto.String(tu.GetTrip().GetRouteId()),
                    DirectionId:          proto.Int32(int32(tu.GetTrip().GetDirectionId())),
                    ScheduleRelationship: proto.Int32(int32(tu.GetTrip().GetScheduleRelationship())),
                },
                Vehicle: &protodata.VehicleDescriptorProto{
                    Id:    proto.String(tu.GetVehicle().GetId()),
                    Label: proto.String(tu.GetVehicle().GetLabel()),
                },
                StopTimeUpdate: stopUpdates,
                Timestamp:      proto.Int64(int64(tu.GetTimestamp())),
            },
        })
    }
    return results, nil
}

func FetchVehiclePositions() ([]*protodata.VehiclePositionEntityProto, error) {
	rawFeed, err := fetchRawGTFS(rtdVehiclePosition)
	if err != nil {
		return nil, err
	}

	var results []*protodata.VehiclePositionEntityProto

	for _, entity := range rawFeed.Entity {
		v := entity.GetVehicle()
		if v == nil {
			continue
		}

		results = append(results, &protodata.VehiclePositionEntityProto{
			Id: proto.String(entity.GetId()),
			Vehicle: &protodata.VehiclePositionProto{
				Trip: &protodata.TripDescriptorProto{
					TripId:               proto.String(v.GetTrip().GetTripId()),
					RouteId:              proto.String(v.GetTrip().GetRouteId()),
					DirectionId:          proto.Int32(int32(v.GetTrip().GetDirectionId())),
					ScheduleRelationship: proto.Int32(int32(v.GetTrip().GetScheduleRelationship())),
				},
				Vehicle: &protodata.VehicleDescriptorProto{
					Id:    proto.String(v.GetVehicle().GetId()),
					Label: proto.String(v.GetVehicle().GetLabel()),
				},
				Position: &protodata.GeoPositionProto{
					Latitude:  proto.Float64(float64(v.GetPosition().GetLatitude())),
					Longitude: proto.Float64(float64(v.GetPosition().GetLongitude())),
					Bearing:   proto.Float64(float64(v.GetPosition().GetBearing())),
				},
				StopId:          proto.String(v.GetStopId()),
				CurrentStatus:   proto.Int32(int32(v.GetCurrentStatus())),
				Timestamp:       proto.Int64(int64(v.GetTimestamp())),
				OccupancyStatus: proto.Int32(int32(v.GetOccupancyStatus())),
			},
		})
	}
	return results, nil
}

// Routes
func findRouteByID(routeId string) (*protodata.RouteProto, bool) {
    data := Routes
    n := len(data)

    idx := sort.Search(n, func(i int) bool {
        // Protobuf generated field is usually RouteId
        return data[i].GetRouteId() >= routeId
    })

    if idx < n && data[i].GetRouteId() == routeId {
        return data[idx], true
    }
    return nil, false
}

// Trips
func findTripByID(tripId string) (*protodata.TripProto, bool) {
    data := Trips
    n := len(data)

    idx := sort.Search(n, func(i int) bool {
        return data[i].GetTripId() >= tripId
    })

    if idx < n && data[idx].GetTripId() == tripId {
        return data[idx], true
    }
    return nil, false
}

// Stops
func findStopById(stopId string) (*protodata.StopProto, bool) {
    data := Stops
    n := len(data)

    idx := sort.Search(n, func(i int) bool {
        return data[i].GetStopId() >= stopId
    })

    if idx < n && data[idx].GetStopId() == stopId {
        return data[idx], true
    }
    return nil, false
}

// Stop Times (Search by StopID)
func findStopTimesByStopID(stopId string) ([]*protodata.StopTimeProto, bool) {
    data := StopTimesByStop
    n := len(data)

    idx := sort.Search(n, func(i int) bool {
        return data[i].GetStopId() >= stopId
    })

    if idx < n && data[idx].GetStopId() == stopId {
        end := idx
        for end < n && data[end].GetStopId() == stopId {
            end++
        }
        return data[idx:end], true
    }

    return nil, false
}

// Stop Times (Search by TripID)
func findStopTimesByTripID(tripId string) ([]*protodata.StopTimeProto, bool) {
    data := StopTimesByTrip
    n := len(data)

    idx := sort.Search(n, func(i int) bool {
        return data[i].GetTripId() >= tripId
    })

    if idx < n && data[idx].GetTripId() == tripId {
        end := idx
        for end < n && data[end].GetTripId() == tripId {
            end++
        }
        return data[idx:end], true
    }
    return nil, false
}

func findStopTimeByTripAndStop(tripId, stopId string) (*protodata.StopTimeProto, bool) {
    stops, found := findStopTimesByTripID(tripId)
    if !found {
        return nil, false
    }

    for _, st := range stops {
        if st.GetStopId() == stopId {
            return st, true
        }
    }
    return nil, false
}

// Shapes
func findShapeById(shapeId string) ([]*protodata.ShapeProto, bool) {
    data := Shapes
    n := len(data)

    idx := sort.Search(n, func(i int) bool {
        return data[i].GetShapeId() >= shapeId
    })

    if idx < n && data[idx].GetShapeId() == shapeId {
        end := idx
        for end < n && data[end].GetShapeId() == shapeId {
            end++
        }
        return data[idx:end], true
    }
    return nil, false
}