package serverx

// TODO: Remove me

// import (
// 	"context"
// 	"errors"
// 	"net"
// 	"sort"

// 	"github.com/oligarch316/go-netx/addressx"
// 	"github.com/oligarch316/go-netx/listenerx/multi"
// )

// var errUnknownDialFailure = errors.New("serverx: unknown dial failure")

// // DialSet TODO.
// type DialSet struct{ mSet multi.Set }

// // Len TODO.
// func (ds DialSet) Len() int { return ds.mSet.Len() }

// // Dial TODO.
// func (ds DialSet) Dial(hs ...multi.Hash) (net.Conn, error) {
// 	var firstErr error

// 	for _, h := range hs {
// 		res, err := ds.mSet.Dial(h)
// 		switch {
// 		case err == nil:
// 			return res, nil
// 		case firstErr == nil:
// 			firstErr = err
// 		}
// 	}

// 	if firstErr == nil {
// 		return nil, errUnknownDialFailure
// 	}

// 	return nil, firstErr
// }

// // DialContext TODO.
// func (ds DialSet) DialContext(ctx context.Context, hs ...multi.Hash) (net.Conn, error) {
// 	var firstErr error

// 	for _, h := range hs {
// 		select {
// 		case <-ctx.Done():
// 			return nil, ctx.Err()
// 		default:
// 			res, err := ds.mSet.DialContext(ctx, h)
// 			switch {
// 			case err == nil:
// 				return res, nil
// 			case firstErr == nil:
// 				firstErr = err
// 			}
// 		}
// 	}

// 	if firstErr == nil {
// 		return nil, errUnknownDialFailure
// 	}

// 	return nil, firstErr
// }

// // Resolve TODO.
// // TODO: is returning []multi.Hash instead of []multi.Addr really the best idea
// // here. What about callers wanting to log/error with net.Addr info when issues occur?
// func (ds DialSet) Resolve(cmps ...addressx.Comparer) []multi.Hash {
// 	var (
// 		addrs    = ds.mSet.Addrs()
// 		ordering = addressx.Ordering(cmps)
// 		res      = make([]multi.Hash, len(addrs))
// 	)

// 	// TODO: Annoyingly repetitive but should be removed wholesale in the next commit
// 	sort.SliceStable(addrs, func(i, j int) bool { return ordering.Less(addrs[i], addrs[j]) })

// 	for i, addr := range addrs {
// 		res[i] = addr
// 	}

// 	return res
// }
