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

## Testing

Create an export directory to save the archive file
```
mkdir export-dir
```

Ensure local uid 2001 has write access to the export directory.  
This is the uid of the unprivledged zcashd docker user.
```
sudo chown 2001 export-dir/
```

Run an ephemeral docker container.  
Map in the export directory.  
Override the entrypoint (so we can pull in a new binary).  

```
docker run --rm -ti \
  -v $(pwd)/export-dir/:/export-dir/ \
  --entrypoint bash \
  electriccoinco/zcashd
```

### Now **from inside the container**  
Update zcash params (or map them in ^v^)

```
zcashd@3ac0a0efacd2:~$ zcash-fetch-params
```

Change to the zcashd data directory in the container.
```
zcashd@3ac0a0efacd2:~$ cd .zcash
```

Fetch the prebuilt binary
```
zcashd@3ac0a0efacd2:~/.zcash$ curl -LO https://github.com/doubtingben/zcashd-block-snapshot/releases/download/v0.0.3/zcashd-block-snapshot-v0.0.3
```

Make it executable
```
zcashd@3ac0a0efacd2:~/.zcash$ chmod +x zcashd-block-snapshot-v0.0.3
```

Run it!  
The onlyargument is the height to stop at.
```
zcashd@3ac0a0efacd2:~/.zcash$ ./zcashd-block-snapshot-v0.0.3 -stop-height 1001 -export-dir /export-dir/
<--------- SNIP ----------->
Updated height: 999, stopping at: 1000
New tip hash: 0000000b70480327694608408728c65c1f1a300bfe705b01baca0f5504092e1b Height: 1000
================== REACHED END HEIGHT ==================
================== ZIPPING BLOCKS ==================
Adding file: ./blocks
Adding file: blocks/blk00000.dat
Adding file: blocks/index
Adding file: blocks/index/000003.log
Adding file: blocks/index/CURRENT
Adding file: blocks/index/LOCK
Adding file: blocks/index/LOG
Adding file: blocks/index/MANIFEST-000002
Adding file: blocks/rev00000.dat
Adding file: ./chainstate
Adding file: chainstate/000003.log
Adding file: chainstate/CURRENT
Adding file: chainstate/LOCK
Adding file: chainstate/LOG
Adding file: chainstate/MANIFEST-000002
```

Check the create archive's contents
```
zcashd@3ac0a0efacd2:~/.zcash$ tar zvtf /export-dir/zcashd-1000.tar.gz 
drwx------ zcashd/zcashd     0 2020-03-26 03:08 ./blocks
-rw------- zcashd/zcashd 16777216 2020-03-26 03:08 blocks/blk00000.dat
drwx------ zcashd/zcashd        0 2020-03-26 03:08 blocks/index
-rw------- zcashd/zcashd     1750 2020-03-26 03:08 blocks/index/000003.log
-rw------- zcashd/zcashd       16 2020-03-26 03:08 blocks/index/CURRENT
-rw------- zcashd/zcashd        0 2020-03-26 03:08 blocks/index/LOCK
-rw------- zcashd/zcashd       57 2020-03-26 03:08 blocks/index/LOG
-rw------- zcashd/zcashd       50 2020-03-26 03:08 blocks/index/MANIFEST-000002
-rw------- zcashd/zcashd  1048576 2020-03-26 03:08 blocks/rev00000.dat
drwx------ zcashd/zcashd        0 2020-03-26 03:08 ./chainstate
-rw------- zcashd/zcashd      110 2020-03-26 03:08 chainstate/000003.log
-rw------- zcashd/zcashd       16 2020-03-26 03:08 chainstate/CURRENT
-rw------- zcashd/zcashd        0 2020-03-26 03:08 chainstate/LOCK
-rw------- zcashd/zcashd       57 2020-03-26 03:08 chainstate/LOG
-rw------- zcashd/zcashd       50 2020-03-26 03:08 chainstate/MANIFEST-000002
```

Looks good?  
Copy to the export directory (mapped outside the contianer)   

And exit the container.
```
zcashd@3ac0a0efacd2:~/.zcash$ cp ./zcashd-1000.tar.gz /export-dir/export-dir/
zcashd@3ac0a0efacd2:~/.zcash$ exit
exit
```

You should now have a block backup!
```
$ ls -l export-dir/
total 4400
-rw-r--r--. 1 2001 2001 4505091 Mar 25 23:11 zcashd-1000.tar.gz
```

## Verifying a backup

Creat a new zcashd container
```
docker run --rm -ti \
  -v $(pwd)/export-dir/:/export-dir/ \
  --entrypoint bash \
  electriccoinco/zcashd
```

List your exports
```
zcashd@db65b3de01bc:~$ ls -l /export-dir/
total 1156996
-rw-r--r-- 1 zcashd zcashd  12059826 Mar 26 13:32 zcashd-1330.tar.gz
-rw-r--r-- 1 zcashd zcashd 683893836 Mar 26 13:58 zcashd-20000.tar.gz
-rw-r--r-- 1 zcashd zcashd 475499425 Mar 26 14:07 zcashd-20050.tar.gz
```

Verify zcash data directory is empty
```
zcashd@db65b3de01bc:~$ ls -la .zcash
total 8
drwxr-xr-x 2 zcashd root   4096 Feb 20 15:48 .
drwxr-xr-x 1 zcashd zcashd 4096 Mar 26 14:12 ..
-rw-r--r-- 1 zcashd root      0 Feb 20 15:48 zcash.conf
```

Untar an export to the data directory
```
zcashd@db65b3de01bc:~$ tar zxvf /export-dir/zcashd-20050.tar.gz -C .zcash/
./blocks
blocks/blk00000.dat
blocks/blk00001.dat
blocks/blk00002.dat
...
```

Start zcashd in the new container, with the exporter blockchain data

```
zcashd@db65b3de01bc:~$ zcashd -printtoconsole
...
Opening LevelDB in /srv/zcashd/.zcash/blocks/index
Opened LevelDB successfully
Opening LevelDB in /srv/zcashd/.zcash/chainstate
Opened LevelDB successfully
LoadBlockIndexDB: last block file = 3
LoadBlockIndexDB: last block file info: CBlockFileInfo(blocks=2535, size=54987494, heights=17508...20065, time=2016-11-27...2016-12-01)
Checking all blk files are present...
LoadBlockIndexDB: transaction index disabled
LoadBlockIndexDB: insight explorer disabled
LoadBlockIndexDB: hashBestChain=000000005a5beab5e09bf8f15cfd5cfe92dc1bf6fb3e11f53a777f1af96465a9 height=20057 date=2016-12-01 16:31:20 progress=0.010321
...
```

Look for the lines like those above that the BlockFile's are found and loaded successfully!
