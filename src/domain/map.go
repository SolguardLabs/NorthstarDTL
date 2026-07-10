package domain

type StringTable map[string]string

func (table StringTable) Clone() StringTable {
	if len(table) == 0 {
		return nil
	}
	clone := make(StringTable, len(table))
	for key, value := range table {
		clone[key] = value
	}
	return clone
}

func CloneRouteIDs(routes []RouteID) []RouteID {
	if len(routes) == 0 {
		return nil
	}
	clone := make([]RouteID, len(routes))
	copy(clone, routes)
	return clone
}
