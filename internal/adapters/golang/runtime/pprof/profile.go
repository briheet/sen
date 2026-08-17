// Package pprof decodes pprof data produced by a Go target.
package pprof

import "github.com/briheet/senbon/internal/model"

type (
	LocationID = model.ProfileLocationID
	Profile    = model.Profile
	ValueType  = model.ValueType
	Sample     = model.ProfileSample
	Location   = model.ProfileLocation
	Frame      = model.ProfileFrame
)
