package lang

type DMap[K1 comparable, K2 comparable, V any] map[K1]map[K2]V

func NewDMap[K1 comparable, K2 comparable, V any]() DMap[K1, K2, V] {
	return make(DMap[K1, K2, V])
}

func AddDMapValue[K1 comparable, K2 comparable, V any](m DMap[K1, K2, V], k1 K1, k2 K2, v V) {
	if _, ok := m[k1]; !ok {
		m[k1] = make(map[K2]V)
	}

	m[k1][k2] = v
}

func HasDMapKey[K1 comparable, K2 comparable, V any](m DMap[K1, K2, V], k1 K1, k2 K2) bool {
	if _, ok := m[k1]; !ok {
		return false
	}

	_, ok := m[k1][k2]
	return ok
}
