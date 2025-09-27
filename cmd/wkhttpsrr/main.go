// Copyright 2025 David Adrian
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	wk "github.com/dadrian/wkhttpsrr"
)

func setFetchUsage(fs *flag.FlagSet) {
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  wkhttpsrr fetch [--ip a,b] [--timeout d] [--insecure] <host[:port]>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}
}

func setValidateUsage(fs *flag.FlagSet) {
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  wkhttpsrr validate [--file path] [--ip a,b] [--timeout d] [--insecure] <host[:port]>")
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

func commonFlags(fs *flag.FlagSet) (ips *string, timeout *time.Duration, insecure *bool) {
	ips = fs.String("ip", "", "comma-separated IP list to connect; skips DNS")
	timeout = fs.Duration("timeout", 10*time.Second, "request timeout")
	insecure = fs.Bool("insecure", false, "skip TLS verification (not recommended)")
	return
}

func fetchCmd(args []string) {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	setFetchUsage(fs)
	ips, timeout, insecure := commonFlags(fs)
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "fetch: missing <host[:port]>")
		fs.Usage()
		os.Exit(2)
	}

	addrs, err := wk.ParseIPList(*ips)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	origin := wk.Origin{
		Host:  fs.Arg(0),
		Addrs: addrs,
	}
	opts := []wk.Option{}
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
	ips, timeout, insecure := commonFlags(fs)
	_ = fs.Parse(args)

	var doc *wk.Document

	if *file != "" {
		b, err := os.ReadFile(*file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		doc, err = wk.Parse(b)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else {
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "validate: missing <host[:port]> or --file path")
			fs.Usage()
			os.Exit(2)
		}
		addrs, err := wk.ParseIPList(*ips)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		origin := wk.Origin{
			Host:  fs.Arg(0),
			Addrs: addrs,
		}
		opts := []wk.Option{}
		if *insecure {
			opts = append(opts, wk.WithInsecureSkipVerify(true))
		}
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		doc, err = wk.Fetch(ctx, origin, opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to fetch %q: %v\n", origin.Host, err)
		}
	}

	err := wk.Validate(doc)
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
