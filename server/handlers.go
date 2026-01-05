package server

import (
	"net/http"
	"studious-waffle/server/protodata"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

// GET /vehiclepositions *real-time
func HandleVehiclePosition(c *gin.Context) {
	positions, err := FetchVehiclePositions()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	collection := &protodata.VehiclePositionCollection{
		Entities:  positions,
		Timestamp: proto.Int64(time.Now().Unix()),
	}

	data, err := proto.Marshal(collection)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, "application/x-protobuf", data)
}

// GET /alerts
func HandleAlert(c *gin.Context) {
    results, err := FetchAlerts()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch alerts"})
        return
    }
    c.JSON(http.StatusOK, results)
}

// GET /tripupdates
func HandleTripUpdate(c *gin.Context) {
    results, err := FetchTripUpdates()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trip updates"})
        return
    }
    c.JSON(http.StatusOK, results)
}

// GET /routes/:id
func HandleRoutesById(c *gin.Context) {
    id := c.Param("id")
    if route, found := findRouteByID(id); found {
        c.JSON(http.StatusOK, route)
    } else {
        c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
    }
}

// GET /stops/:id
func HandleStopsById(c *gin.Context) {
    id := c.Param("id")
    if stop, found := findStopById(id); found {
        c.JSON(http.StatusOK, stop)
    } else {
        c.JSON(http.StatusNotFound, gin.H{"error": "Stop not found"})
    }
}

// GET /shapes/:id
func HandleShapesById(c *gin.Context) {
    id := c.Param("id")
    // Note: findShapeById returns []*protodata.ShapeProto
    if shapes, found := findShapeById(id); found {
        c.JSON(http.StatusOK, shapes)
    } else {
        c.JSON(http.StatusNotFound, gin.H{"error": "Shape not found"})
    }
}

type GeoParams struct {
    Lat    float64 `form:"lat" binding:"required"`
    Lon    float64 `form:"lon" binding:"required"`
    Radius float64 `form:"radius,default=1.0"`
}

// GET /routes/near
func HandleNearRoutes(c *gin.Context) {
    var p GeoParams
    if err := c.ShouldBindQuery(&p); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lon query params required"})
        return
    }

    // Chain the lookups using protobuf-aware helper functions
    nearStops := _FindStopsWithinXMiles(p.Lat, p.Lon, p.Radius, Stops)
    stopTimes := _FindStopTimesForEachStop(nearStops, StopTimes)
    trips := _FindTripsForEachStopTime(stopTimes, Trips)
    routes := _FindRouteTripShape(trips, Routes)

    c.JSON(http.StatusOK, routes)
}

func _FindStopsWithinXMiles(userLat, userLon, miles float64, allStops []*protodata.StopProto) []*protodata.StopProto {
    nearbyStops := make([]*protodata.StopProto, 0)
    userCoord := haversine.Coord{Lat: userLat, Lon: userLon}

    for _, s := range allStops {
        // Use Get...() for protobuf fields
        stopCoord := haversine.Coord{Lat: s.GetStopLat(), Lon: s.GetStopLon()}
        dist, _ := haversine.Distance(userCoord, stopCoord)

        if dist <= miles {
            nearbyStops = append(nearbyStops, s)
        }
    }
    return nearbyStops
}

func _FindRouteTripShape(foundTrips []*protodata.TripProto, allRoutes []*protodata.RouteProto) []RouteShape {
    results := make([]RouteShape, 0)
    seenRoutes := make(map[string]struct{})

    for _, t := range foundTrips {
        routeID := t.GetRouteId()
        if _, exists := seenRoutes[routeID]; exists {
            continue
        }

        idx := sort.Search(len(allRoutes), func(j int) bool {
            return allRoutes[j].GetRouteId() >= routeID
        })

        if idx < len(allRoutes) && allRoutes[idx].GetRouteId() == routeID {
            results = append(results, RouteShape{
                ShapeID: t.GetShapeId(),
                TripID:  t.GetTripId(),
                RouteID: routeID,
            })
            seenRoutes[routeID] = struct{}{}
        }
    }
    return results
}