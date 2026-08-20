package runtime

import (
	"time"

	"github.com/briheet/sen/internal/adapters/tigerbeetle/analysis"
	"github.com/briheet/sen/internal/model"
)

func requestProfile(startedAt time.Time, duration time.Duration, requests map[requestKey]*requestWindow) map[string]*model.Profile {
	profile := &model.Profile{
		StartedAt: startedAt, Duration: duration,
		SampleTypes:       []model.ValueType{{Type: "requests", Unit: "count"}, {Type: "latency", Unit: "nanoseconds"}},
		DefaultSampleType: "requests", Locations: make(map[model.ProfileLocationID]model.ProfileLocation),
	}
	var locationID model.ProfileLocationID
	for key, window := range requests {
		if window.count == 0 {
			continue
		}
		operationLocation := locationID + 1
		replicaLocation := locationID + 2
		locationID += 2
		profile.Locations[operationLocation] = location(operationLocation, key.operation, analysis.OperationPath(key.operation))
		profile.Locations[replicaLocation] = location(replicaLocation, "replica", analysis.ReplicaPath(key.replica))
		profile.Samples = append(profile.Samples, model.ProfileSample{
			Values: []int64{int64(window.count), int64(microseconds(window.sumUS))},
			Stack:  []model.ProfileLocationID{operationLocation, replicaLocation},
		})
	}
	if len(profile.Samples) == 0 {
		return nil
	}
	return map[string]*model.Profile{profileName: profile}
}

func location(id model.ProfileLocationID, function, file string) model.ProfileLocation {
	return model.ProfileLocation{ID: id, Frames: []model.ProfileFrame{{Function: function, File: file, Line: 1}}}
}
