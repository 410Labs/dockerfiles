// update-apex-ips keeps a zone apex A record in step with a subdomain.
//
// Given a source name such as www.example.com, it resolves the source's IPv4 addresses
// and compares them with the apex (example.com) in public DNS. If they match, it stops.
// Otherwise it compares with the apex A record in Route53 and, with -apply, updates it.
// (A zone apex cannot be a CNAME, so the apex gets the source's IPs as plain A records.)
//
// Default is plan-only: it prints the difference and exits 2 if the apex is out of date.
// Exit codes: 0 in sync (or applied), 1 error, 2 drift found in plan-only mode.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	r53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"golang.org/x/net/publicsuffix"
)

// newRecordTTL applies only when the apex has no A record yet; an existing record keeps its TTL.
const newRecordTTL = 300

type options struct {
	name    string
	zoneID  string
	apply   bool
	timeout time.Duration
}

func main() {
	var o options
	flag.StringVar(&o.name, "name", "", "source name whose IPs the apex should have, e.g. www.example.com (required)")
	flag.StringVar(&o.zoneID, "zone-id", "", "Route53 hosted zone ID (default: the public zone named for the apex)")
	flag.BoolVar(&o.apply, "apply", false, "update Route53 (default: plan only)")
	flag.DurationVar(&o.timeout, "timeout", 30*time.Second, "overall timeout")
	showVersion := flag.Bool("version", false, "print the module version this binary was built from, and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version())
		return
	}
	if o.name == "" {
		flag.Usage()
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	defer cancel()

	code, err := run(ctx, o)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	os.Exit(code)
}

// buildVersion is set at link time (-ldflags "-X main.buildVersion=vX.Y.Z") from the TAG file,
// so the published image reports the tag it is published under.
var buildVersion string

// version is buildVersion when it was set, else the module version in the build info
// ("(devel)" for a plain local build), else "unknown".
func version() string {
	if buildVersion != "" {
		return buildVersion
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "unknown"
}

func run(ctx context.Context, o options) (int, error) {
	apex, err := publicsuffix.EffectiveTLDPlusOne(strings.TrimSuffix(strings.ToLower(o.name), "."))
	if err != nil {
		return 0, fmt.Errorf("apex of %s: %w", o.name, err)
	}
	if fqdn(o.name) == fqdn(apex) {
		return 0, fmt.Errorf("%s is already the apex; pass the subdomain to copy from, e.g. www.%s", o.name, apex)
	}

	want, err := lookupIPv4(ctx, o.name)
	if err != nil {
		return 0, fmt.Errorf("resolve %s: %w", o.name, err)
	}
	if len(want) == 0 {
		// Never copy an empty answer onto the apex.
		return 0, fmt.Errorf("%s has no IPv4 addresses; refusing to continue", o.name)
	}
	fmt.Printf("source:   %s -> %s\n", o.name, strings.Join(want, ", "))

	// If public DNS already gives the apex the right IPs there is no reason to touch Route53.
	// A failed or empty lookup just means we go on to Route53.
	pub, perr := lookupIPv4(ctx, apex)
	if perr != nil {
		fmt.Printf("public:   %s lookup failed (%v); checking Route53\n", apex, perr)
	} else {
		fmt.Printf("public:   %s -> %s\n", apex, strings.Join(pub, ", "))
		if a, r := diff(pub, want); len(a) == 0 && len(r) == 0 {
			fmt.Println("public DNS already matches; nothing to do")
			return 0, nil
		}
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return 0, fmt.Errorf("aws config: %w", err)
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1" // Route53 is global but the SDK still wants a region
	}
	r53 := route53.NewFromConfig(cfg)

	if o.zoneID == "" {
		o.zoneID, err = findZone(ctx, r53, apex)
		if err != nil {
			return 0, fmt.Errorf("find hosted zone (or pass -zone-id): %w", err)
		}
	}
	have, ttl, err := currentRecord(ctx, r53, o.zoneID, apex)
	if err != nil {
		return 0, fmt.Errorf("route53 lookup: %w", err)
	}

	add, remove := diff(have, want)
	if have == nil {
		fmt.Printf("route53:  zone %s: (no A record for %s)\n", o.zoneID, apex)
		ttl = newRecordTTL
	} else {
		fmt.Printf("route53:  zone %s: %s (ttl %d)\n", o.zoneID, strings.Join(have, ", "), ttl)
	}
	if len(add) == 0 && len(remove) == 0 && have != nil {
		fmt.Println("route53 already matches; public DNS may be lagging or cached")
		return 0, nil
	}
	fmt.Printf("drift:    add [%s] remove [%s]\n", strings.Join(add, ", "), strings.Join(remove, ", "))
	if !o.apply {
		fmt.Println("plan only; rerun with -apply to update")
		return 2, nil
	}

	out, err := r53.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(o.zoneID),
		ChangeBatch: &r53types.ChangeBatch{
			Comment: aws.String("update-apex-ips: apex copied from " + o.name),
			Changes: []r53types.Change{{
				Action: r53types.ChangeActionUpsert,
				ResourceRecordSet: &r53types.ResourceRecordSet{
					Name:            aws.String(apex),
					Type:            r53types.RRTypeA,
					TTL:             aws.Int64(ttl),
					ResourceRecords: records(want),
				},
			}},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("route53 update: %w", err)
	}
	fmt.Printf("updated; change %s %s\n", aws.ToString(out.ChangeInfo.Id), out.ChangeInfo.Status)
	return 0, nil
}

// lookupIPv4 resolves host with the system resolver to sorted, de-duplicated IPv4 addresses.
// A name that does not exist returns an empty list rather than an error.
func lookupIPv4(ctx context.Context, host string) ([]string, error) {
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", strings.TrimSuffix(host, "."))
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, ip := range ips {
		out = append(out, ip.String())
	}
	return normalize(out), nil
}

// findZone returns the ID of the public hosted zone whose name is the longest suffix of name.
// It needs route53:ListHostedZones. Ties (e.g. two zones with the same name) are an error.
func findZone(ctx context.Context, c *route53.Client, name string) (string, error) {
	var zones []zoneRef
	p := route53.NewListHostedZonesPaginator(c, &route53.ListHostedZonesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return "", err
		}
		for _, z := range page.HostedZones {
			if z.Config == nil || !z.Config.PrivateZone {
				zones = append(zones, zoneRef{ID: strings.TrimPrefix(aws.ToString(z.Id), "/hostedzone/"), Name: aws.ToString(z.Name)})
			}
		}
	}
	return pickZone(zones, name)
}

type zoneRef struct{ ID, Name string }

func pickZone(zones []zoneRef, name string) (string, error) {
	n := fqdn(name)
	var best []zoneRef
	for _, z := range zones {
		zn := fqdn(z.Name)
		if n != zn && !strings.HasSuffix(n, "."+zn) {
			continue
		}
		if len(best) == 0 || len(zn) > len(fqdn(best[0].Name)) {
			best = []zoneRef{z}
		} else if len(zn) == len(fqdn(best[0].Name)) {
			best = append(best, z)
		}
	}
	switch len(best) {
	case 0:
		return "", fmt.Errorf("no public hosted zone contains %s", name)
	case 1:
		return best[0].ID, nil
	}
	return "", fmt.Errorf("%d public zones are named %s", len(best), best[0].Name)
}

// currentRecord returns the A record values and TTL, or nil if none exists.
// It refuses alias records, which this tool does not manage.
func currentRecord(ctx context.Context, c *route53.Client, zoneID, name string) ([]string, int64, error) {
	want := fqdn(name)
	out, err := c.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(zoneID),
		StartRecordName: aws.String(want),
		StartRecordType: r53types.RRTypeA,
		MaxItems:        aws.Int32(1),
	})
	if err != nil {
		return nil, 0, err
	}
	for _, rs := range out.ResourceRecordSets {
		if fqdn(aws.ToString(rs.Name)) != want || rs.Type != r53types.RRTypeA {
			continue
		}
		if rs.AliasTarget != nil {
			return nil, 0, errors.New("existing A record is an alias; not managed by this tool")
		}
		if rs.SetIdentifier != nil {
			return nil, 0, errors.New("existing A record is a weighted/latency/failover set; not managed by this tool")
		}
		var vals []string
		for _, r := range rs.ResourceRecords {
			vals = append(vals, aws.ToString(r.Value))
		}
		return normalize(vals), aws.ToInt64(rs.TTL), nil
	}
	return nil, 0, nil
}

func fqdn(s string) string { return strings.ToLower(strings.TrimSuffix(s, ".")) + "." }

func normalize(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if s = strings.TrimSpace(s); s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// diff returns the values to add to, and remove from, have so that it equals want.
func diff(have, want []string) (add, remove []string) {
	h, w := map[string]bool{}, map[string]bool{}
	for _, s := range have {
		h[s] = true
	}
	for _, s := range want {
		w[s] = true
		if !h[s] {
			add = append(add, s)
		}
	}
	for _, s := range have {
		if !w[s] {
			remove = append(remove, s)
		}
	}
	return
}

func records(ips []string) []r53types.ResourceRecord {
	rs := make([]r53types.ResourceRecord, 0, len(ips))
	for _, ip := range ips {
		rs = append(rs, r53types.ResourceRecord{Value: aws.String(ip)})
	}
	return rs
}
