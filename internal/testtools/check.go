package testtools

func HaveSameElements[T comparable](set1 []T, set2 []T) bool {
	c1 := valueCounts(set1)
	c2 := valueCounts(set2)
	return areCountsEqual(c1, c2)
}

func valueCounts[T comparable](set []T) map[T]int {
	result := make(map[T]int)
	for _, x := range set {
		result[x]++
	}
	return result
}

func areCountsEqual[T comparable](map1, map2 map[T]int) bool {
	if map1 == nil && map2 == nil {
		return true
	}

	if map1 == nil || map2 == nil {
		return false
	}

	if len(map1) != len(map2) {
		return false
	}

	for key, value1 := range map1 {
		value2, ok := map2[key]
		if !ok || value1 != value2 {
			return false
		}
	}

	return true
}
