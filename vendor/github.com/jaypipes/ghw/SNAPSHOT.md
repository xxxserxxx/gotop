# ghw snapshots

For ghw, snapshots are partial clones of the `/proc`, `/sys` (et. al.) subtrees copied from arbitrary
machines, which ghw can consume later. "partial" is because the snapshot doesn't need to contain a
complete copy of all the filesystem subtree (that is doable but inpractical). It only needs to contain
the paths ghw cares about. The snapshot concept was introduced [to make ghw easier to test](https://github.com/jaypipes/ghw/issues/66).

## Create and consume snapshot

The recommended way to create snapshots for ghw is to use the `ghw-snapshot` tool.
This tool is maintained by the ghw authors, and snapshots created with this tool are guaranteed to work.

To consume the ghw snapshots, please check the `README.md` document.

## Snapshot design and definitions

The remainder of this document will describe how a snapshot looks like and provides rationale for all the major design decisions.
Even though this document aims to provide all the necessary information to understand how ghw creates snapshots and what you should
expect, we recommend to check also the [project issues](https://github.com/jaypipes/ghw/issues) and the `git` history to have the full picture.

### Scope

ghw supports snapshots only on linux platforms. This restriction may be lifted in future releases.
Snapshots must be consumable in the following supported ways:

1. (way 1) from docker (or podman), mounting them as volumes. See `hack/run-against-snapshot.sh`
2. (way 2) using the environment variables `GHW_SNAPSHOT_*`. See `README.md` for the full documentation.

Other combinations are possible, but are unsupported and may stop working any time.
You should depend only on the supported ways to consume snapshots.

### Snapshot content constraints

Stemming from the use cases, the snapshot content must have the following properties:

0. (constraint 0) MUST contain the same information as live system (obviously). Whatever you learn from a live system, you MUST be able to learn from a snapshot.
1. (constraint 1) MUST NOT require any post processing before it is consumable besides, obviously, unpacking the `.tar.gz` on the right directory - and pointing ghw to that directory.
2. (constraint 2) MUST NOT require any special handling nor special code path in ghw. From ghw perspective running against a live system or against a snapshot should be completely transparent.
3. (constraint 3) MUST contain only data - no executable code is allowed ever. This makes snapshots trivially safe to share and consume.
4. (constraint 4) MUST NOT contain any personally-identifiable data. Data gathered into a snapshot is for testing and troubleshooting purposes and should be safe to send to troubleshooters to analyze.

It must be noted that trivially cloning subtrees from `/proc` and `/sys` and creating a tarball out of them doesn't work
because both pseudo filesystems make use of symlinks, and [docker doesn't really play nice with symlinks](https://github.com/jaypipes/ghw/commit/f8ffd4d24e62eb9017511f072ccf51f13d4a3399).
This conflcits with (way 1) above.

