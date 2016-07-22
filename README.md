# Minsync

Minsync is a tool for synchronizing contents of large files to devices with differing read and write speeds.
The tool will read both copies of a file and copy only differentiating blocks.

The handles sparse files well.

## Running

minsync /disk/large.file /slowdisk/large.file

