package grpcx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/oligarch316/go-netx/listenerx/multi"
	"google.golang.org/grpc/resolver"
)

const rsvSchemeDNS = "dns"

var (
	rsvAddrPrefix    = fmt.Sprintf("_%s_:", ID)
	rsvAddrPrefixLen = len(rsvAddrPrefix)
)

func rsvFormatHash(h multi.SetHash) string { return rsvAddrPrefix + h.HashString() }

func rsvParseHash(s string) (local bool, h multi.SetHash, err error) {
	if local = strings.HasPrefix(s, rsvAddrPrefix); !local {
		return
	}

	h, err = multi.ParseSetHash(s[rsvAddrPrefixLen:])
	return
}

type hashResolver interface{ Resolve() []multi.SetAddr }

// ResolverParams TODO.
type ResolverParams struct{ SchemeName, DNSHostName *string }

func (rp ResolverParams) build(hResolver hashResolver) []resolver.Builder {
	var (
		builder = rsvBuilder{hResolver: hResolver}
		res     []resolver.Builder
	)

	if rp.SchemeName != nil {
		res = append(res, &rsvBuilderScheme{
			schemeName: *rp.SchemeName,
			rsvBuilder: builder,
		})
	}

	if rp.DNSHostName != nil {
		res = append(res, &rsvBuilderDNS{
			hostName:   *rp.DNSHostName,
			rsvBuilder: builder,
		})
	}

	return res
}

type rsvBuilderScheme struct {
	schemeName string
	rsvBuilder
}

func (rbs *rsvBuilderScheme) Scheme() string { return rbs.schemeName }

func (rbs *rsvBuilderScheme) Build(_ resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	return rbs.build(cc)
}

type rsvBuilderDNS struct {
	hostName string
	rsvBuilder
}

func (*rsvBuilderDNS) Scheme() string { return rsvSchemeDNS }

func (rbd *rsvBuilderDNS) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if target.URL.Host == rbd.hostName {
		return rbd.build(cc)
	}

	orig := resolver.Get(rsvSchemeDNS)
	if _, ok := orig.(*rsvBuilderDNS); ok {
		return nil, fmt.Errorf("grpcx: %s resolver: unable to recover original %s resolver", rsvSchemeDNS, rsvSchemeDNS)
	}

	return orig.Build(target, cc, opts)
}

type rsvBuilder struct{ hResolver hashResolver }

func (rb *rsvBuilder) build(cc resolver.ClientConn) (resolver.Resolver, error) {
	hashAddrs := rb.hResolver.Resolve()
	if len(hashAddrs) < 1 {
		return nil, errors.New("no local addresses")
	}

	rsvAddrs := make([]resolver.Address, len(hashAddrs))
	for i, hashAddr := range hashAddrs {
		rsvAddrs[i] = resolver.Address{Addr: rsvFormatHash(hashAddr)}
	}

	cc.UpdateState(resolver.State{Addresses: rsvAddrs})
	return &rsvNoop{}, nil
}

type rsvNoop struct{}

func (*rsvNoop) ResolveNow(resolver.ResolveNowOptions) {}
func (*rsvNoop) Close()                                {}
