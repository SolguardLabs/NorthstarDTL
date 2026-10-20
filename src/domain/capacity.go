package domain

const capacityStressHaircutBps int64 = 1_500

// CapacityProjection is a read-only operational view; source and destination
// amounts are intentionally kept in their respective asset units.
type CapacityProjection struct {
	RouteID                     RouteID     `json:"routeId"`
	SourceAsset                 Asset       `json:"sourceAsset"`
	DestinationAsset            Asset       `json:"destinationAsset"`
	Status                      RouteStatus `json:"status"`
	AdmissionOpen               bool        `json:"admissionOpen"`
	SourceExposure              Money       `json:"sourceExposure"`
	SourceLimit                 Money       `json:"sourceLimit"`
	SourceHeadroom              Money       `json:"sourceHeadroom"`
	DestinationLiquidity        Money       `json:"destinationLiquidity"`
	DestinationStressHaircutBps int64       `json:"destinationStressHaircutBps"`
	DestinationStressLiquidity  Money       `json:"destinationStressLiquidity"`
	DestinationStressLoss       Money       `json:"destinationStressLoss"`
}

func NewCapacityProjection(route Route) CapacityProjection {
	retainedBps := BpsScale - capacityStressHaircutBps
	liquidity := int64(route.OutputLiquidity)
	stressed := Money((liquidity/BpsScale)*retainedBps + (liquidity%BpsScale)*retainedBps/BpsScale)
	return CapacityProjection{
		RouteID:                     route.ID,
		SourceAsset:                 route.SourceAsset,
		DestinationAsset:            route.DestinationAsset,
		Status:                      route.StatusOrDefault(),
		AdmissionOpen:               route.ActiveForAdmission(),
		SourceExposure:              route.Exposure,
		SourceLimit:                 route.MaxExposure,
		SourceHeadroom:              route.RemainingExposure(),
		DestinationLiquidity:        route.OutputLiquidity,
		DestinationStressHaircutBps: capacityStressHaircutBps,
		DestinationStressLiquidity:  stressed,
		DestinationStressLoss:       route.OutputLiquidity - stressed,
	}
}
