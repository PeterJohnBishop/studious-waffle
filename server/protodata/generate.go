package protodata

import (
	"bufio"
	"cmp"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"

	"google.golang.org/protobuf/proto"
)

const outputUrl = "/Users/peterbishop/Development/studious-waffle/server/protodata/"
const inputUrl = "/Users/peterbishop/Development/studious-waffle/server/protodata/input/"

func OpenCSVReader(fileName string) (*csv.Reader, *os.File, error) {
	file, err := os.Open(inputUrl + fileName)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening file: %w", err)
	}

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	// Read the header line to skip it
	if _, err := reader.Read(); err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("error reading header: %w", err)
	}

	return reader, file, nil
}

func GenerateRouteData() bool {
	outPath := outputUrl + "routes.generated.go"
	reader, inFile, err := OpenCSVReader("routes.txt")
	if err != nil {
		fmt.Println("Error opening routes.txt:", err)
		return false
	}
	defer inFile.Close()

	var data []*RouteProto
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		var routeType int32
		fmt.Sscanf(row[5], "%d", &routeType)

		data = append(data, &RouteProto{
			RouteId:        proto.String(row[0]),
			AgencyId:       proto.String(row[1]),
			RouteShortName: proto.String(row[2]),
			RouteLongName:  proto.String(row[3]),
			RouteDesc:      proto.String(row[4]),
			RouteType:      proto.Int32(routeType),
			RouteUrl:       proto.String(row[6]),
			RouteColor:     proto.String(row[7]),
			RouteTextColor: proto.String(row[8]),
		})
	}

	fmt.Println("Sorting Routes...")
	slices.SortFunc(data, func(a, b *RouteProto) int {
		return cmp.Compare(a.GetRouteId(), b.GetRouteId())
	})

	return writeGeneratedFile(outPath, "Routes", data)
}

func GenerateTripData() bool {
	outPath := outputUrl + "trips.generated.go"
	reader, inFile, err := OpenCSVReader("trips.txt")
	if err != nil {
		return false
	}
	defer inFile.Close()

	var data []*TripProto
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		var dID int32
		fmt.Sscanf(row[4], "%d", &dID)

		data = append(data, &TripProto{
			RouteId:      proto.String(row[0]),
			ServiceId:    proto.String(row[1]),
			TripId:       proto.String(row[2]),
			TripHeadsign: proto.String(row[3]),
			DirectionId:  proto.Int32(dID),
			BlockId:      proto.String(row[5]),
			ShapeId:      proto.String(row[6]),
		})
	}

	slices.SortFunc(data, func(a, b *TripProto) int {
		return cmp.Compare(a.GetTripId(), b.GetTripId())
	})

	return writeGeneratedFile(outPath, "Trips", data)
}

func GenerateStopData() bool {
	outPath := outputUrl + "stops.generated.go"
	reader, inFile, err := OpenCSVReader("stops.txt")
	if err != nil {
		return false
	}
	defer inFile.Close()

	var data []*StopProto
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		lat, _ := strconv.ParseFloat(row[4], 64)
		lon, _ := strconv.ParseFloat(row[5], 64)
		var locType, wheel int32
		fmt.Sscanf(row[8], "%d", &locType)
		fmt.Sscanf(row[11], "%d", &wheel)

		data = append(data, &StopProto{
			StopId:             proto.String(row[0]),
			StopCode:           proto.String(row[1]),
			StopName:           proto.String(row[2]),
			StopDesc:           proto.String(row[3]),
			StopLat:            proto.Float64(lat),
			StopLon:            proto.Float64(lon),
			ZoneId:             proto.String(row[6]),
			StopUrl:            proto.String(row[7]),
			LocationType:       proto.Int32(locType),
			ParentStation:      proto.String(row[9]),
			StopTimezone:       proto.String(row[10]),
			WheelchairBoarding: proto.Int32(wheel),
		})
	}

	slices.SortFunc(data, func(a, b *StopProto) int {
		return cmp.Compare(a.GetStopId(), b.GetStopId())
	})

	return writeGeneratedFile(outPath, "Stops", data)
}

func GenerateStopTimeData() bool {
	reader, inFile, err := OpenCSVReader("stop_times.txt")
	if err != nil {
		return false
	}
	defer inFile.Close()

	var data []*StopTimeProto
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		var seq, pick, drop, timep int32
		fmt.Sscanf(row[4], "%d", &seq)
		fmt.Sscanf(row[6], "%d", &pick)
		fmt.Sscanf(row[7], "%d", &drop)
		fmt.Sscanf(row[9], "%d", &timep)
		dist, _ := strconv.ParseFloat(row[8], 64)

		data = append(data, &StopTimeProto{
			TripId:            proto.String(row[0]),
			ArrivalTime:       proto.String(row[1]),
			DepartureTime:     proto.String(row[2]),
			StopId:            proto.String(row[3]),
			StopSequence:      proto.Int32(seq),
			StopHeadsign:      proto.String(row[5]),
			PickupType:        proto.Int32(pick),
			DropOffType:       proto.Int32(drop),
			ShapeDistTraveled: proto.Float64(dist),
			Timepoint:         proto.Int32(timep),
		})
	}

	slices.SortFunc(data, func(a, b *StopTimeProto) int {
		if c := cmp.Compare(a.GetTripId(), b.GetTripId()); c != 0 {
			return c
		}
		return cmp.Compare(a.GetStopSequence(), b.GetStopSequence())
	})
	stopTimesByTripGenerated := writeGeneratedFile(outputUrl+"stop_times_by_trip.generated.go", "StopTimesByTrip", data)

	slices.SortFunc(data, func(a, b *StopTimeProto) int {
		if c := cmp.Compare(a.GetStopId(), b.GetStopId()); c != 0 {
			return c
		}
		return cmp.Compare(a.GetTripId(), b.GetTripId())
	})
	stopTimesByStopGenerated := writeGeneratedFile(outputUrl+"stop_times_by_stop.generated.go", "StopTimesByStop", data)

	return (stopTimesByTripGenerated && stopTimesByStopGenerated)
}

func GenerateShapeData() bool {
	outPath := outputUrl + "shapes.generated.go"
	reader, inFile, err := OpenCSVReader("shapes.txt")
	if err != nil {
		return false
	}
	defer inFile.Close()

	var data []*ShapeProto
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		lat, _ := strconv.ParseFloat(row[1], 64)
		lon, _ := strconv.ParseFloat(row[2], 64)
		var seq int32
		fmt.Sscanf(row[3], "%d", &seq)
		dist, _ := strconv.ParseFloat(row[4], 64)

		data = append(data, &ShapeProto{
			ShapeId:           proto.String(row[0]),
			ShapePtLat:        proto.Float64(lat),
			ShapePtLon:        proto.Float64(lon),
			ShapePtSequence:   proto.Int32(seq),
			ShapeDistTraveled: proto.Float64(dist),
		})
	}

	// sort primarily by ShapeId, then by sequence
	slices.SortFunc(data, func(a, b *ShapeProto) int {
		if c := cmp.Compare(a.GetShapeId(), b.GetShapeId()); c != 0 {
			return c
		}
		return cmp.Compare(a.GetShapePtSequence(), b.GetShapePtSequence())
	})

	return writeGeneratedFile(outPath, "Shapes", data)
}

func writeGeneratedFile(outPath string, varName string, data interface{}) bool {
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Println("Error creating Go file:", err)
		return false
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	fmt.Fprintln(writer, "// Code generated by transit-generator; DO NOT EDIT.")
	fmt.Fprintln(writer, "package protodata")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "import \"google.golang.org/protobuf/proto\"")
	fmt.Fprintln(writer, "")
	fmt.Fprintf(writer, "var %s = []", varName)

	switch v := data.(type) {
	case []*RouteProto:
		fmt.Fprintln(writer, "*RouteProto{")
		for _, r := range v {
			fmt.Fprintf(writer, "\t{RouteId: proto.String(%q), AgencyId: proto.String(%q), RouteShortName: proto.String(%q), RouteLongName: proto.String(%q), RouteDesc: proto.String(%q), RouteType: proto.Int32(%d), RouteUrl: proto.String(%q), RouteColor: proto.String(%q), RouteTextColor: proto.String(%q)},\n",
				r.GetRouteId(), r.GetAgencyId(), r.GetRouteShortName(), r.GetRouteLongName(), r.GetRouteDesc(), r.GetRouteType(), r.GetRouteUrl(), r.GetRouteColor(), r.GetRouteTextColor())
		}
	case []*TripProto:
		fmt.Fprintln(writer, "*TripProto{")
		for _, t := range v {
			fmt.Fprintf(writer, "\t{RouteId: proto.String(%q), ServiceId: proto.String(%q), TripId: proto.String(%q), TripHeadsign: proto.String(%q), DirectionId: proto.Int32(%d), BlockId: proto.String(%q), ShapeId: proto.String(%q)},\n",
				t.GetRouteId(), t.GetServiceId(), t.GetTripId(), t.GetTripHeadsign(), t.GetDirectionId(), t.GetBlockId(), t.GetShapeId())
		}
	case []*StopProto:
		fmt.Fprintln(writer, "*StopProto{")
		for _, s := range v {
			fmt.Fprintf(writer, "\t{StopId: proto.String(%q), StopName: proto.String(%q), StopLat: proto.Float64(%f), StopLon: proto.Float64(%f), LocationType: proto.Int32(%d)},\n",
				s.GetStopId(), s.GetStopName(), s.GetStopLat(), s.GetStopLon(), s.GetLocationType())
		}
	case []*ShapeProto:
		fmt.Fprintln(writer, "*ShapeProto{")
		for _, s := range v {
			fmt.Fprintf(writer, "\t{ShapeId: proto.String(%q), ShapePtLat: proto.Float64(%f), ShapePtLon: proto.Float64(%f), ShapePtSequence: proto.Int32(%d), ShapeDistTraveled: proto.Float64(%f)},\n",
				s.GetShapeId(), s.GetShapePtLat(), s.GetShapePtLon(), s.GetShapePtSequence(), s.GetShapeDistTraveled())
		}
	case []*StopTimeProto:
		fmt.Fprintln(writer, "*StopTimeProto{")
		for _, st := range v {
			fmt.Fprintf(writer, "\t{TripId: proto.String(%q), ArrivalTime: proto.String(%q), DepartureTime: proto.String(%q), StopId: proto.String(%q), StopSequence: proto.Int32(%d)},\n",
				st.GetTripId(), st.GetArrivalTime(), st.GetDepartureTime(), st.GetStopId(), st.GetStopSequence())
		}
	}

	fmt.Fprintln(writer, "}")
	return true
}
