package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"strings"
	"time"

	wk "github.com/dadrian/wkhttpsrr"
)

func setFetchUsage(fs *flag.FlagSet) {
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  wkhttpsrr fetch [--ipv4 a,b] [--ipv6 x,y] [--timeout d] [--insecure] <host[:port]>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}
}

func setValidateUsage(fs *flag.FlagSet) {
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  wkhttpsrr validate [--file path] [--ipv4 a,b] [--ipv6 x,y] [--timeout d] [--insecure] <host[:port]>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}
}

func defaultUsage(fs *flag.FlagSet) {
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  wkhttpsrr {fetch|validate} [FLAGS] [ARGS]")
		fmt.Fprintln(os.Stderr)
		fs.PrintDefaults()
	}
}

func parseIPList(s string) ([]netip.Addr, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	addrs := make([]netip.Addr, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		a, err := netip.ParseAddr(p)
		if err != nil {
			return nil, fmt.Errorf("invalid IP %q: %w", p, err)
		}
		addrs = append(addrs, a)
	}
	return addrs, nil
}

func main() {
	// Only used for help
	dfs := flag.NewFlagSet("", flag.ContinueOnError)

	if len(os.Args) < 2 {
		// Show the subcommands
		defaultUsage(dfs)
		dfs.Usage()
		os.Exit(2)
	}

	sub := os.Args[1]
	switch sub {
	case "fetch":
		fetchCmd(os.Args[2:])
	case "validate":
		validateCmd(os.Args[2:])
	case "-h", "--help", "help":
		defaultUsage(dfs)
		dfs.Usage()
		os.Exit(2)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", sub)
		defaultUsage(dfs)
		dfs.Usage()
		os.Exit(2)
	}
}

func commonFlags(fs *flag.FlagSet) (ipv4 *string, ipv6 *string, timeout *time.Duration, insecure *bool) {
	ipv4 = fs.String("ipv4", "", "comma-separated IPv4 list to connect; skips DNS")
	ipv6 = fs.String("ipv6", "", "comma-separated IPv6 list to connect; skips DNS")
	timeout = fs.Duration("timeout", 10*time.Second, "request timeout")
	insecure = fs.Bool("insecure", false, "skip TLS verification (not recommended)")
	return
}

func buildOrigin(target string, v4s, v6s []netip.Addr) (wk.Origin, []wk.Option, error) {
	o, err := wk.ParseOrigin(target)
	if err != nil {
		return wk.Origin{}, nil, err
	}
	opts := []wk.Option{}
	if len(v4s) > 0 || len(v6s) > 0 {
		opts = append(opts, wk.WithIPs(v4s, v6s))
	}
	return o, opts, nil
}

func fetchCmd(args []string) {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	setFetchUsage(fs)
	ipv4, ipv6, timeout, insecure := commonFlags(fs)
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "fetch: missing <host[:port]>")
		fs.Usage()
		os.Exit(2)
	}
	v4s, err := parseIPList(*ipv4)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	v6s, err := parseIPList(*ipv6)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	origin, opts, err := buildOrigin(fs.Arg(0), v4s, v6s)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *insecure {
		opts = append(opts, wk.WithInsecureSkipVerify(true))
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	doc, err := wk.Fetch(ctx, origin, opts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	b, _ := wk.MarshalPretty(doc)
	fmt.Println(string(b))
}

func validateCmd(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	file := fs.String("file", "", "path to JSON file to validate instead of fetching")
	setValidateUsage(fs)
	ipv4, ipv6, timeout, insecure := commonFlags(fs)
	_ = fs.Parse(args)

	var (
		doc *wk.Document
		err error
	)
	if *file != "" {
		b, rerr := os.ReadFile(*file)
		if rerr != nil {
			fmt.Fprintln(os.Stderr, rerr)
			os.Exit(2)
		}
		doc, err = wk.Parse(b)
	} else {
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "validate: missing <host[:port]> or --file path")
			fs.Usage()
			os.Exit(2)
		}
		v4s, e1 := parseIPList(*ipv4)
		if e1 != nil {
			fmt.Fprintln(os.Stderr, e1)
			os.Exit(2)
		}
		v6s, e2 := parseIPList(*ipv6)
		if e2 != nil {
			fmt.Fprintln(os.Stderr, e2)
			os.Exit(2)
		}
		origin, opts, e3 := buildOrigin(fs.Arg(0), v4s, v6s)
		if e3 != nil {
			fmt.Fprintln(os.Stderr, e3)
			os.Exit(2)
		}
		if *insecure {
			opts = append(opts, wk.WithInsecureSkipVerify(true))
		}
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		doc, err = wk.Fetch(ctx, origin, opts...)
	}
	if err == nil {
		err = wk.Validate(doc)
	}
	if err != nil {
		// Unwrap for concise error output
		var uerr interface{ Unwrap() error }
		if errors.As(err, &uerr) && uerr.Unwrap() != nil {
			err = uerr.Unwrap()
		}
		fmt.Fprintln(os.Stderr, "invalid:", err)
		os.Exit(1)
	}
	fmt.Println("valid")
}
