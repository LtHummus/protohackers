package pest

import "sync"

type Hub struct {
	siteLock sync.Mutex
	SiteMap  map[uint32]*Site
}

func NewHub() *Hub {
	return &Hub{
		SiteMap: map[uint32]*Site{},
	}
}

func (h *Hub) GetOrQuerySite(site uint32) (*Site, error) {
	h.siteLock.Lock()
	defer h.siteLock.Unlock()

	s := h.SiteMap[site]
	if s != nil {
		return s, nil
	}

	cs, err := NewSite(site)
	if err != nil {
		return nil, err
	}

	h.SiteMap[site] = cs
	return cs, nil
}
