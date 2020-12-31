package structures

import (
	"math/rand"

	"github.com/USACE/go-consequences/consequences"
	"github.com/USACE/go-consequences/hazards"
)

type BaseStructure struct {
	Name   string
	DamCat string
	X, Y   float64
}
type StructureStochastic struct {
	BaseStructure
	OccType                     OccupancyTypeStochastic
	StructVal, ContVal, FoundHt consequences.ParameterValue
}
type StructureDeterministic struct {
	BaseStructure
	OccType                     OccupancyTypeDeterministic
	StructVal, ContVal, FoundHt float64
}

func (s BaseStructure) GetX() float64 {
	return s.X
}
func (s BaseStructure) GetY() float64 {
	return s.Y
}
func (s StructureStochastic) SampleStructure(seed int64) StructureDeterministic {
	ot := s.OccType.SampleOccupancyType(seed)
	sv := s.StructVal.SampleValue(rand.Float64())
	cv := s.ContVal.SampleValue(rand.Float64())
	fh := s.FoundHt.SampleValue(rand.Float64())
	return StructureDeterministic{OccType: ot, StructVal: sv, ContVal: cv, FoundHt: fh, BaseStructure: BaseStructure{DamCat: s.DamCat}}
}

func (s StructureStochastic) ComputeConsequences(d interface{}) consequences.ConsequenceDamageResult {
	return s.SampleStructure(rand.Int63()).ComputeConsequences(d) //this needs work so seeds can be controlled.
}
func (s StructureDeterministic) ComputeConsequences(d interface{}) consequences.ConsequenceDamageResult { //what if we invert this general model to hazard.damage(consequence receptor)
	header := []string{"structure damage", "content damage"}
	results := []interface{}{0.0, 0.0}
	var ret = consequences.ConsequenceDamageResult{Headers: header, Results: results}
	de, ok := d.(hazards.DepthEvent)
	if ok {
		depth := de.Depth
		return computeFloodConsequences(depth, s)
	}
	def, okd := d.(float64)
	if okd {
		return computeFloodConsequences(def, s)
	}
	fire, okf := d.(hazards.FireEvent)
	if okf {
		damagePercent := s.OccType.Structuredamfun.SampleValue(fire.Intensity) / 100 //assumes what type the damage array is in
		cdamagePercent := s.OccType.Contentdamfun.SampleValue(fire.Intensity) / 100
		ret.Results[0] = damagePercent * s.StructVal
		ret.Results[1] = cdamagePercent * s.ContVal
		return ret
	}
	return ret
}
func computeFloodConsequences(d float64, s StructureDeterministic) consequences.ConsequenceDamageResult {
	header := []string{"structure damage", "content damage"}
	results := []interface{}{0.0, 0.0}
	var ret = consequences.ConsequenceDamageResult{Headers: header, Results: results}
	depthAboveFFE := d - s.FoundHt
	damagePercent := s.OccType.Structuredamfun.SampleValue(depthAboveFFE) / 100 //assumes what type the damage array is in
	cdamagePercent := s.OccType.Contentdamfun.SampleValue(depthAboveFFE) / 100
	ret.Results[0] = damagePercent * s.StructVal
	ret.Results[1] = cdamagePercent * s.ContVal
	return ret
}

/*
func DefaultStructure() StructureDeterministic {
	//get the occupancy type map
	m := OccupancyTypeMap()
	// select a base structure type for testing
	var o = m["RES1-1SNB"]
	var s = StructureDeterministic{OccType: o.SampleOccupancyType(1), StructVal: 100.0, ContVal: 10.0, FoundHt: 0.0, BaseStructure: BaseStructure{DamCat: "category"}}
	return s
}

func DefaultStructureU() StructureStochastic {
	//get the occupancy type map
	m := OccupancyTypeMap()
	// select a base structure type for testing
	var o = m["RES1-1SNB"]
	sv := statistics.NormalDistribution{Mean: 0, StandardDeviation: 1}
	cv := statistics.NormalDistribution{Mean: 0, StandardDeviation: 1}
	spv := consequences.ParameterValue{Value: sv}
	cpv := consequences.ParameterValue{Value: cv}
	fhpv := consequences.ParameterValue{Value: 0}
	var s = StructureStochastic{OccType: o, StructVal: spv, ContVal: cpv, FoundHt: fhpv, BaseStructure: BaseStructure{DamCat: "category"}}
	return s
}
func ConvertBaseStructureToFire(s StructureDeterministic) StructureDeterministic {
	var fire = hazards.FireDamageFunction{}
	s.OccType.Structuredamfun = fire
	s.OccType.Contentdamfun = fire
	return s
}*/