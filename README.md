# multisum

`multisum` is a small utility that creates multiple checksum file in a single read operation.


## Usage

    multisum <hash-options> <checksum-mode> --target <target-dir> <source-dir>

where:

- `<hash-options>` is at least one of the supported hashes to calculate checksums for
    - `--md5`
    - `--sha1`
    - `--sha256`
    - `--sha512`
- `<checksum-mode>` is one of the checksum output modes
    - `--binary` print checksums in binary mode (*), default
    - `--text` print checksums in text mode ( )
- `<target-dir>` is the target directory to write the checksum files (e.g. SHA256SUMS file)
- `<source-dir>` is the source directory to create the checksum files for
