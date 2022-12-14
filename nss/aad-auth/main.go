package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ubuntu/aad-auth/internal/cache"
	"github.com/ubuntu/aad-auth/internal/logger"
	"github.com/ubuntu/aad-auth/internal/nss"
)

var opts []cache.Option

func main() {
	flag.Usage = aadAuthUsage
	flag.Parse()

	switch flag.Arg(0) {
	case "getent":
		ctx := nss.CtxWithSyslogLogger(context.Background())
		defer logger.CloseLoggerFromContext(ctx)

		db := flag.Arg(1)

		var key *string
		if len(flag.Args()) > 2 {
			k := flag.Arg(2)
			key = &k
		}

		out, err := Getent(ctx, db, key, opts...)
		if err != nil {
			exit(1, fmt.Sprintf("Error when trying to list %v from %s: %v", key, db, err))
		}
		fmt.Print(out)
	case "":
		exit(1, "Missing required argument.")
	default:
		exit(1, fmt.Sprintf("Invalid argument %q", flag.Arg(0)))
	}
}

func aadAuthUsage() {
	fmt.Fprintf(os.Stderr, `
This executable should not be used directly, but should you wish too:

Usage: aad_auth getent {dbName} {key}
		
    - dbName: Name of the database to be queried.
        - Supported databases: %v
    - key (optional): name or uid/gid of the entry to be queried for.`, strings.Join(supportedDbs, ", "))
}

func exit(status int, message string) {
	if message != "" {
		fmt.Fprintln(os.Stderr, message)
	}
	flag.Usage()
	os.Exit(status)
}
