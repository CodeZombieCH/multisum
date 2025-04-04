# multisum

`multisum` is a small utility that creates multiple checksum files in a single read operation.

The calculated checksum files have the same format as the output generated by the GNU Coreutils `md5sum`, `sha256sum`, `sha512sum` and alike (to be precise, the [Untagged output format](https://www.gnu.org/software/coreutils/manual/html_node/cksum-output-modes.html)). To validate files with the created checksum files, these utils should be used (e.g. `sha256sum --check SHA256SUMS`).


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

### Example

    multisum --sha256 --sha512 --target ./checksums/photos ~/photos

This will create a `SHA256SUMS` and `SHA512SUMS` file in the `./checksums/photos` directory for all the files in the `~/photos` directory.
