package network

import (
	"sync"
	"time"

	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

type PeerCache interface {
	GetPeerForKey(string) (pstore.PeerInfo, bool)
	UpdateCache(string, pstore.PeerInfo)
}

type cpeer struct {
	key        string
	lastUpdate time.Time
	connCount  int
	info       pstore.PeerInfo
}

type peerCache struct {
	peers   map[peer.ID]*cpeer
	store   map[string]*cpeer
	muStore sync.RWMutex
}

type noopCache struct{}

func (nc *noopCache) GetPeerForKey(_ string) (_ pstore.PeerInfo, _ bool) { return }
func (nc *noopCache) UpdateCache(_ string, _ pstore.PeerInfo)            {}

func NewNoopCache() PeerCache {
	return &noopCache{}
}

func NewPeerCache(h host.Host) PeerCache {
	pc := &peerCache{
		peers: make(map[peer.ID]*cpeer),
		store: make(map[string]*cpeer),
	}

	h.Network().Notify(pc)
	return pc
}

func (pc *peerCache) GetPeerForKey(key string) (pinfo pstore.PeerInfo, ok bool) {
	pc.muStore.RLock()
	defer pc.muStore.RUnlock()

	var p *cpeer
	if p, ok = pc.store[key]; ok {
		pinfo = p.info
	}

	return
}

func (pc *peerCache) UpdateCache(key string, pi pstore.PeerInfo) {
	pc.muStore.Lock()
	defer pc.muStore.Unlock()

	p, ok := pc.peers[pi.ID]
	if !ok || p.connCount <= 0 {
		return
	}

	p.key = key
	p.info = pi
	pc.store[key] = p
}

func (pc *peerCache) Connected(net inet.Network, c inet.Conn) {
	pc.muStore.Lock()
	defer pc.muStore.Unlock()
	peerID := c.RemotePeer()
	p, ok := pc.peers[peerID]

	if !ok {
		p = &cpeer{
			connCount:  0,
			lastUpdate: time.Now(),
		}

		pc.peers[peerID] = p
	}

	p.connCount++
}

func (pc *peerCache) Disconnected(net inet.Network, c inet.Conn) {
	peerID := c.RemotePeer()
	pc.muStore.Lock()
	defer pc.muStore.Unlock()
	if p, ok := pc.peers[peerID]; ok {
		p.connCount--

		if p.connCount == 0 {
			if p.key != "" {
				delete(pc.store, p.key)
			}
			delete(pc.peers, peerID)
		}
	}
}

// Listen is no-op in this implementation.
func (pc *peerCache) Listen(n inet.Network, addr ma.Multiaddr) {
}

// ListenClose is no-op in this implementation.
func (pc *peerCache) ListenClose(n inet.Network, addr ma.Multiaddr) {

}

// OpenedStream is no-op in this implementation.
func (pc *peerCache) OpenedStream(n inet.Network, s inet.Stream) {

}

// ClosedStream is no-op in this implementation.
func (pc *peerCache) ClosedStream(n inet.Network, s inet.Stream) {
}
