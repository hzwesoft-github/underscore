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

type SliceMap[K comparable, V any] map[K][]V

func NewSliceMap[K comparable, V any]() SliceMap[K, V] {
	return make(SliceMap[K, V])
}

func AddSliceMapValue[K comparable, V any](m SliceMap[K, V], key K, value V) {
	if _, ok := m[key]; !ok {
		m[key] = make([]V, 0)
	}

	m[key] = append(m[key], value)
}

type Set[T comparable] map[T]bool

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func AddToSet[T comparable](set Set[T], value T) Set[T] {
	set[value] = true
	return set
}

func SetValues[T comparable](set Set[T]) (list []T) {
	for k := range set {
		list = append(list, k)
	}

	return list
}
