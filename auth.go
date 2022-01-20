package authdump

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alessio/shellescape"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	// ErrUnchanged is return when authentication configuration hasn't changed since prevRevision.
	ErrUnchanged = errors.New("unchanged: auth revision is the same as before")
)

// Dump returns shell commands that'd set up authentication on an empty cluster the same as the given cluster.
// Note that passwords can't be recovered.
// You can optionally pass in a prevision revision, to get ErrUnchanged if etcd is unchanged since that auth revision.
// The return arguments are 1) a list of shell commands, 2) the auth revision of this dump and 3) an optional error.
func Dump(ctx context.Context, c *clientv3.Client, prevRevision uint64) (commands []string, dumpedAuthRevision uint64, err error) {
	as, err := c.AuthStatus(ctx)
	if err != nil {
		return nil, 0, err
	}
	if as.AuthRevision == prevRevision {
		return nil, 0, ErrUnchanged
	}

	rl, err := c.RoleList(ctx)
	if err != nil {
		return nil, 0, err
	}

	for _, n := range rl.Roles {
		commands = append(commands, "etcdctl role add "+shellescape.Quote(n))
		u, err := c.RoleGet(ctx, n)
		if err != nil {
			return nil, 0, err
		}
		for _, p := range u.Perm {
			prefix := "etcdctl role grant-permission " + shellescape.Quote(n) + " " + shellescape.Quote(strings.ToLower(p.PermType.String())) + " "
			if bytes.Compare(p.Key, []byte{0}) == 0 && bytes.Compare(p.RangeEnd, []byte{0}) == 0 {
				commands = append(commands, prefix+"'' --prefix")
			} else if bytes.Compare(p.Key, p.RangeEnd) == 0 {
				commands = append(commands, prefix+shellescape.Quote(string(p.Key)))
			} else if compareOffByOne(p.Key, p.RangeEnd) {
				commands = append(commands, prefix+shellescape.Quote(string(p.Key))+" --prefix")
			} else {
				commands = append(commands, prefix+shellescape.Quote(string(p.Key))+" "+shellescape.Quote(string(p.RangeEnd)))
			}
		}
	}

	ul, err := c.UserList(ctx)
	if err != nil {
		return nil, 0, err
	}

	for _, n := range ul.Users {
		commands = append(commands, "etcdctl user add "+shellescape.Quote(n))
		u, err := c.UserGet(ctx, n)
		if err != nil {
			return nil, 0, err
		}
		for _, r := range u.Roles {
			commands = append(commands, "etcdctl user grant-role "+shellescape.Quote(n)+" "+shellescape.Quote(r))
		}
	}

	if as.Enabled {
		commands = append(commands, "etcdctl auth enable")
	} else {
		commands = append(commands, "etcdctl auth disable")
	}

	as2, err := c.AuthStatus(ctx)
	if err != nil {
		return nil, 0, err
	}

	if as.AuthRevision != as2.AuthRevision {
		return nil, 0, fmt.Errorf("authentication configuration was changed during the dump (auth revision %d at start, %d at end)", as.AuthRevision, as2.AuthRevision)
	}

	return commands, as2.AuthRevision, nil
}

func compareOffByOne(a, b []byte) bool {
	l := len(a)
	if l != len(b) || l == 0 {
		return false
	}
	var off byte = 1
	for l--; l >= 0; l-- {
		if a[l]+off != b[l] {
			return false
		}
		if a[l] != 255 {
			off = 0
		}
	}
	return true
}
