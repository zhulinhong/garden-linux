package rootfs_provider

var DefaultUIDMap MappingList = MappingList{{
	FromID: 0,
	ToID:   10001,
	Size:   10000,
}}

var DefaultGIDMap = DefaultUIDMap

type Mapping struct {
	FromID int
	ToID   int
	Size   int
}

type MappingList []Mapping

func (m MappingList) Map(id int) int {
	for _, m := range m {
		if delta := id - m.FromID; delta < m.Size {
			return m.ToID + delta
		}
	}

	return id
}
