package grpcx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/multi"
	"google.golang.org/grpc"
	grpcr "google.golang.org/grpc/resolver"
)

const (
	rsvAddrPrefix    = namespace
	rsvAddrPrefixLen = len(namespace)

	rsvSchemeDNS = "dns"
)

// ResolverParams TODO.
type ResolverParams struct {
	SchemeName, DNSHostName *string
	NetworkPriority         []string
}

func buildResolversOption(md multi.Dialer, params ResolverParams) grpc.DialOption {
	var (
		addrSet  = rsvAddrSet{addrs: md.Addrs()}
		builders []grpcr.Builder
	)

	if len(params.NetworkPriority) > 0 {
		addrSet.defaultComparer = netx.ByPriorityNetwork(params.NetworkPriority...)
	}

	if params.SchemeName != nil {
		builders = append(builders, &rsvBuilderScheme{
			schemeName: *params.SchemeName,
			addrSet:    addrSet,
		})
	}

	if params.DNSHostName != nil {
		builders = append(builders, &rsvBuilderDNS{
			hostName: *params.DNSHostName,
			addrSet:  addrSet,
		})
	}

	return grpc.WithResolvers(builders...)
}

func rsvFormatHash(h multi.Hash) string { return rsvAddrPrefix + h.Format() }

func rsvParseHash(s string) (local bool, h multi.Hash, err error) {
	if local = strings.HasPrefix(rsvAddrPrefix, s); !local {
		return
	}

	h, err = multi.ParseHash(s[rsvAddrPrefixLen:])
	return
}

type rsvAddrSet struct {
	addrs           []multi.Addr
	defaultComparer netx.AddrComparer
}

func (ras rsvAddrSet) process(endpoint string) (*grpcr.State, error) {
	var (
		nAddrs = len(ras.addrs)
		sorter multi.AddrSorter
	)

	if nAddrs < 1 {
		return nil, errors.New("no local addresses")
	}

	if cleaned := strings.Trim(endpoint, "/"); cleaned != "" {
		networks := strings.Split(cleaned, "/")
		sorter = append(sorter, netx.ByPriorityNetwork(networks...))
	}

	if ras.defaultComparer != nil {
		sorter = append(sorter, ras.defaultComparer)
	}

	var (
		mAddrs = make([]multi.Addr, nAddrs)
		rAddrs = make([]grpcr.Address, nAddrs)
	)

	copy(mAddrs, ras.addrs)

	if sorter != nil {
		sorter.Stable(mAddrs)
	}

	for i, mAddr := range mAddrs {
		rAddrs[i] = grpcr.Address{Addr: rsvFormatHash(mAddr)}
	}

	return &grpcr.State{Addresses: rAddrs}, nil
}

type rsvBuilderScheme struct {
	schemeName string
	addrSet    rsvAddrSet
}

func (rbs *rsvBuilderScheme) Scheme() string { return rbs.schemeName }

func (rbs *rsvBuilderScheme) Build(target grpcr.Target, cc grpcr.ClientConn, _ grpcr.BuildOptions) (grpcr.Resolver, error) {
	state, err := rbs.addrSet.process(target.Endpoint)
	if err != nil {
		return nil, Error{ Component: rbs.schemeName+" resolver", err: err }
	}

	cc.UpdateState(*state)
	return &rsvNoop{}, nil
}

type rsvBuilderDNS struct {
	hostName string
	addrSet  rsvAddrSet
}

func (*rsvBuilderDNS) Scheme() string { return rsvSchemeDNS }

func (rbd *rsvBuilderDNS) Build(target grpcr.Target, cc grpcr.ClientConn, opts grpcr.BuildOptions) (grpcr.Resolver, error) {
	split := strings.SplitN(target.Endpoint, "/", 2)

	if split[0] != rbd.hostName {
		orig := grpcr.Get(rsvSchemeDNS)
		if _, ok := orig.(*rsvBuilderDNS); ok {
			return nil, Error{
				Component: rsvSchemeDNS+" resolver",
				err: fmt.Errorf("unable to recover original %s resolver", rsvSchemeDNS),
			}
		}

		return orig.Build(target, cc, opts)
	}

	var endpoint string
	if len(split) > 1 {
		endpoint = split[1]
	}

	state, err := rbd.addrSet.process(endpoint)
	if err != nil {
		return nil, Error{ Component: rsvSchemeDNS+" resolver", err: err }
	}

	cc.UpdateState(*state)
	return &rsvNoop{}, nil
}

type rsvNoop struct{}

func (*rsvNoop) ResolveNow(grpcr.ResolveNowOptions) {}

func (*rsvNoop) Close() {}
