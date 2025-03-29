package main

import (
	"crypto"
	"log"
	"os"

	flag "github.com/spf13/pflag"
)

type options struct {
	// Checksum print mode
	printTextMode   bool
	printBinaryMode bool

	// Hashes to compute
	useMD5    bool
	useSha1   bool
	useSha256 bool
	useSha512 bool

	// Path to directory of files to be checksum'ed
	sourceDirectoryPath string
	// Path to directory where to create the checksum files
	targetDirectoryPath string
}

func readOptions() options {
	o := options{}

	// Modes
	flag.BoolVar(&o.printTextMode, "text", false, "print checksums in binary mode (*)")
	flag.BoolVar(&o.printBinaryMode, "binary", false, "print checksums in text mode ( )")

	flag.BoolVar(&o.useMD5, "md5", false, "calculate MD5 sums")
	flag.BoolVar(&o.useSha1, "sha1", false, "calculate SHA1 sums")
	flag.BoolVar(&o.useSha256, "sha256", false, "calculate SHA256 sums")
	flag.BoolVar(&o.useSha512, "sha512", false, "calculate SHA512 sums")

	flag.StringVar(&o.targetDirectoryPath, "target", "", "target directory to store checksum files")

	flag.Parse()

	if flag.NArg() != 1 {
		log.Printf("missing positional arguments:")
		log.Printf("multisum [OPTIONS] source-dir")
		os.Exit(1)
	}

	o.sourceDirectoryPath = os.Args[len(os.Args)-1]

	return o
}

func setDefaultOptions(o *options) {
	// Set default
	if !o.printBinaryMode && !o.printTextMode {
		// default to binary
		o.printBinaryMode = true
	}
}

func validateOptions(o options) {
	if o.printBinaryMode && o.printTextMode {
		log.Fatalf("invalid arguments: --binary and --text are mutually exclusive")
	}

	hasHash := o.useMD5 || o.useSha256 || o.useSha512
	if !hasHash {
		log.Fatalf("invalid arguments: must use at least one hash flag")
	}
}

func createConfig(o options) ChecksumCalculatorConfig {
	cfg := ChecksumCalculatorConfig{}

	if o.printTextMode {
		cfg.printMode = ModeText
	} else {
		cfg.printMode = ModeBinary
	}

	hashes := make([]crypto.Hash, 0)
	if o.useMD5 {
		hashes = append(hashes, crypto.MD5)
	}
	if o.useSha1 {
		hashes = append(hashes, crypto.SHA1)
	}
	if o.useSha256 {
		hashes = append(hashes, crypto.SHA256)
	}
	if o.useSha512 {
		hashes = append(hashes, crypto.SHA512)
	}
	cfg.hashes = hashes

	cfg.targetDirectoryPath = o.targetDirectoryPath
	cfg.sourceDirectoryPath = o.sourceDirectoryPath

	return cfg
}

func main() {
	log.SetFlags(0)

	options := readOptions()

	setDefaultOptions(&options)

	validateOptions(options)

	// Check args
	checkDirectoryPath("store", options.sourceDirectoryPath)
	checkDirectoryPath("repo", options.targetDirectoryPath)

	cfg := createConfig(options)
	cfg.Print()

	c := NewChecksumCalculator(cfg)
	if err := c.Calculate(cfg, options.sourceDirectoryPath); err != nil {
		log.Fatalf("failed: %v", err)
	}
}

func checkDirectoryPath(argName string, path string) {
	if len(path) == 0 {
		log.Fatalf("%s: empty path", argName)
	}

	f, err := os.Stat(path)
	if err != nil {
		log.Fatalf("%s: invalid path or not accessible", argName)
	}

	if !f.IsDir() {
		log.Fatalf("%s: not a directory", argName)
	}
}
