package x

// Keys return all keys of map
func Keys[K comparable, V any](maps map[K]V) []K {
	items := []K{}
	for k := range maps {
		items = append(items, k)
	}
	return items
}

// UpdateMap update the map and return deleted
// , IN map[K]V1 | []V1
func UpdateMap[K comparable, V1 any, V2 any](
	original map[K]V2,
	inputs []V1,
	convertFn func(V1) V2,
	keyFn func(V1) K,
	updateForce bool,
) (changed map[K]V2, deleted map[K]V2) {
	changed = map[K]V2{}
	deleted = map[K]V2{}

	for _, i := range inputs {
		k := keyFn(i)
		// udpate if force update or not exits
		if o, ok := original[k]; ok && !updateForce {
			changed[k] = o
		} else {
			// panic if convert  is nil
			changed[k] = convertFn(i)
		}
	}

	// generete items we need to delete
	for k, v := range original {
		_, ok := changed[k]
		if !ok {
			deleted[k] = v
		}
	}

	return changed, deleted
}
