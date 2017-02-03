# Minsync

Minsync is a tool for synchronizing contents of large files to devices with differing read and write speeds.
The tool will read both copies of a file and copy only differentiating blocks.

The tool preserves existing and add new holes in sparse files.

## Running

minsync /disk/large.file /slowdisk/large.file

