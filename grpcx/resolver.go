package grpcx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/oligarch316/go-netx/multi"
	"github.com/oligarch316/go-netx/multi/addrsort"
	"github.com/oligarch316/go-netx/serverx"
	"google.golang.org/grpc/resolver"
)

const rsvSchemeDNS = "dns"

var (
	rsvAddrPrefix    = fmt.Sprintf("_%s_:", id.Namespace)
	rsvAddrPrefixLen = len(rsvAddrPrefix)
)

func rsvFormatHash(h multi.Hash) string { return rsvAddrPrefix + h.Format() }

func rsvParseHash(s string) (local bool, h multi.Hash, err error) {
	if local = strings.HasPrefix(rsvAddrPrefix, s); !local {
		return
	}

	h, err = multi.ParseHash(s[rsvAddrPrefixLen:])
	return
}

// ResolverParams TODO.
type ResolverParams struct {
	SchemeName, DNSHostName *string
	AddressOrder            addrsort.CompareList
}

func (rp ResolverParams) build(dSet *serverx.DialSet) []resolver.Builder {
	var (
		res  []resolver.Builder
		rSet = rsvSet{baseSorter: rp.AddressOrder, set: dSet}
	)

	if rp.SchemeName != nil {
		res = append(res, &rsvBuilderScheme{
			schemeName: *rp.SchemeName,
			rsvSet:     rSet,
		})
	}

	if rp.DNSHostName != nil {
		res = append(res, &rsvBuilderDNS{
			hostName: *rp.DNSHostName,
			rsvSet:   rSet,
		})
	}

	return res
}

type rsvSet struct {
	baseSorter addrsort.CompareList
	set        *serverx.DialSet
}

func (rs rsvSet) process(endpoint string) (*resolver.State, error) {
	var (
		nAddrs = rs.set.Len()
		sorter = rs.baseSorter
	)

	if nAddrs < 1 {
		return nil, errors.New("no local addresses")
	}

	if cleaned := strings.Trim(endpoint, "/"); cleaned != "" {
		cmp := addrsort.ByPriorityNetwork(strings.Split(cleaned, "/")...)
		sorter = append(addrsort.CompareList{cmp}, sorter...)
	}

	var (
		hashes = rs.set.Resolve(sorter...)
		res    = &resolver.State{
			Addresses: make([]resolver.Address, nAddrs),
		}
	)

	for i, h := range hashes {
		res.Addresses[i] = resolver.Address{Addr: rsvFormatHash(h)}
	}

	return res, nil
}

type rsvBuilderScheme struct {
	schemeName string
	rsvSet
}

func (rbs *rsvBuilderScheme) Scheme() string { return rbs.schemeName }

func (rbs *rsvBuilderScheme) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	state, err := rbs.process(target.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("grpcx: %s resolver: %w", rbs.schemeName, err)
	}

	cc.UpdateState(*state)
	return &rsvNoop{}, nil
}

type rsvBuilderDNS struct {
	hostName string
	rsvSet
}

func (*rsvBuilderDNS) Scheme() string { return rsvSchemeDNS }

func (rbd *rsvBuilderDNS) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	split := strings.SplitN(target.Endpoint, "/", 2)

	if split[0] != rbd.hostName {
		orig := resolver.Get(rsvSchemeDNS)
		if _, ok := orig.(*rsvBuilderDNS); ok {
			return nil, fmt.Errorf("grpcx: %s resolver: unable to recover original %s resolver", rsvSchemeDNS, rsvSchemeDNS)
		}

		return orig.Build(target, cc, opts)
	}

	var endpoint string
	if len(split) > 1 {
		endpoint = split[1]
	}

	state, err := rbd.process(endpoint)
	if err != nil {
		return nil, fmt.Errorf("grpcx: %s resolver: %w", rsvSchemeDNS, err)
	}

	cc.UpdateState(*state)
	return &rsvNoop{}, nil
}

type rsvNoop struct{}

func (*rsvNoop) ResolveNow(resolver.ResolveNowOptions) {}

func (*rsvNoop) Close() {}
