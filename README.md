# zcashd-block-snapshot

A program to start `zcashd` and watch the output.

Takes a single argument of the block height to stop at.

Look for zcashd stdout lines that start with `UpdateTip`.

On matching line, check the height, if it matches the height given at startup:
- stop zcashd
- create gzipped tar archive of `./blocks/*` and `./chainstate/*`

## Requires

- `zcashd` binary installed in current $PATH
- `zcashd` configured through default config file
- executed from the zcashd data directory

