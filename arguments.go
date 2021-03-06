package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
)

var usage = fmt.Sprintf(`Muffet, the web repairgirl

Usage:
	muffet [-b <size>] [-c <concurrency>] [-e <pattern>...] [-f] [-j <header>...] [-l <times>] [-p] [-r] [-s] [-u] [-t <seconds>] [-v] [-x] <url>

Options:
	-b, --buffer-size <size>          Set HTTP response buffer size in bytes. [default: %v]
	-c, --concurrency <concurrency>   Roughly maximum number of concurrent HTTP connections. [default: %v]
	-e, --exclude <pattern>...        Exclude URLs matched with given regular expressions.
	-f, --ignore-fragments            Ignore URL fragments.
	-h, --help                        Show this help.
	-j, --header <header>...          Set custom headers.
	-l, --limit-redirections <times>  Limit a number of redirections. [default: %v]
	-p, --one-page-only               Only check links found in the given URL, do not follow links.
	-r, --follow-robots-txt           Follow robots.txt when scraping.
	-s, --follow-sitemap-xml          Scrape only pages listed in sitemap.xml.
	-u, --follow-url-params           Will not cut off the URL parameters.
	-t, --timeout <seconds>           Set timeout for HTTP requests in seconds. [default: %v]
	-v, --verbose                     Show successful results too.
	-x, --skip-tls-verification       Skip TLS certificates verification.`,
	defaultBufferSize, defaultConcurrency, defaultMaxRedirections, defaultTimeout.Seconds())

type arguments struct {
	BufferSize       int
	Concurrency      int
	ExcludedPatterns []*regexp.Regexp
	FollowRobotsTxt,
	FollowSitemapXML,
	FollowURLParams bool
	Headers         map[string]string
	IgnoreFragments bool
	MaxRedirections int
	Timeout         time.Duration
	URL             string
	Verbose,
	SkipTLSVerification bool
	OnePageOnly bool
}

func getArguments(ss []string) (arguments, error) {
	args := parseArguments(usage, ss)

	b, err := parseInt(args["--buffer-size"].(string))

	if err != nil {
		return arguments{}, err
	}

	c, err := parseInt(args["--concurrency"].(string))

	if err != nil {
		return arguments{}, err
	}

	ss, _ = args["--exclude"].([]string)
	rs, err := compileRegexps(ss)

	if err != nil {
		return arguments{}, err
	}

	hs := map[string]string(nil)

	if ss := args["--header"]; ss != nil {
		hs, err = parseHeaders(ss.([]string))

		if err != nil {
			return arguments{}, err
		}
	}

	r, err := parseInt(args["--limit-redirections"].(string))

	if err != nil {
		return arguments{}, err
	}

	t, err := parseInt(args["--timeout"].(string))

	if err != nil {
		return arguments{}, err
	}

	return arguments{
		b,
		c,
		rs,
		args["--follow-robots-txt"].(bool),
		args["--follow-sitemap-xml"].(bool),
		args["--follow-url-params"].(bool),
		hs,
		args["--ignore-fragments"].(bool),
		r,
		time.Duration(t) * time.Second,
		args["<url>"].(string),
		args["--verbose"].(bool),
		args["--skip-tls-verification"].(bool),
		args["--one-page-only"].(bool),
	}, nil
}

func parseArguments(u string, ss []string) map[string]interface{} {
	args, err := docopt.ParseArgs(u, ss, "1.3.3")

	if err != nil {
		panic(err)
	}

	return args
}

func parseInt(s string) (int, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	return int(i), err
}

func compileRegexps(ss []string) ([]*regexp.Regexp, error) {
	rs := make([]*regexp.Regexp, 0, len(ss))

	for _, s := range ss {
		r, err := regexp.Compile(s)

		if err != nil {
			return nil, err
		}

		rs = append(rs, r)
	}

	return rs, nil
}

func parseHeaders(ss []string) (map[string]string, error) {
	m := make(map[string]string, len(ss))

	for _, s := range ss {
		i := strings.IndexRune(s, ':')

		if i < 0 {
			return nil, errors.New("invalid header format")
		}

		m[s[:i]] = strings.TrimSpace(s[i+1:])
	}

	return m, nil
}
