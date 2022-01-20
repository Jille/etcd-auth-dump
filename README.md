# etcd-auth-dump

A libary and binary to dump authentication information from etcd. The commands are suitable for configuring an empty etcd cluster to get to the same authentication config.

Note that etcd doesn't return passwords, so those are not included in the dump.

## Parameters for the binary

All configuration is passed in through environment variables. It takes for example these settings:

- ETCD_ENDPOINTS is where to find your etcd cluster
- ETCD_USERNAME and ETCD_PASSWORD are used to connect to etcd. No authentication is used if you leave them unset/empty.

See https://github.com/Jille/etcd-client-from-env for the full list of parameters for connecting to etcd.

## Example output

```
etcdctl role add etcd-postgres-sync
etcdctl role grant-permission etcd-postgres-sync read '' --prefix
etcdctl user add postgres_syncer
etcdctl user grant-role postgres_syncer etcd-postgres-sync
etcdctl user add root
etcdctl user grant-role root root
etcdctl auth enable
```
