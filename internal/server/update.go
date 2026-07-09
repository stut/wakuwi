package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const releasesAPI = "https://api.github.com/repos/stut/wakuwi/releases/latest"

// Release identifies a newer published version for the UI to advertise.
type Release struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

// updateChecker polls GitHub once a day for the latest release and remembers
// it when it is newer than the running version.
type updateChecker struct {
	current semver

	mu     sync.Mutex
	latest *Release
}

func newUpdateChecker(version string) *updateChecker {
	c := &updateChecker{}
	// Dev builds ("dev", or anything unparseable) never advertise updates,
	// so don't spend GitHub rate limit on them either.
	current, ok := parseSemver(version)
	if !ok {
		return c
	}
	c.current = current
	go c.run()
	return c
}

func (c *updateChecker) run() {
	for {
		c.check()
		time.Sleep(24 * time.Hour)
	}
}

// Latest returns the newest known release, or nil when up to date.
func (c *updateChecker) Latest() *Release {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.latest
}

func (c *updateChecker) check() {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(releasesAPI)
	if err != nil {
		log.Printf("update check: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("update check: %s returned %s", releasesAPI, resp.Status)
		return
	}

	var body struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		log.Printf("update check: decode response: %v", err)
		return
	}

	latest, ok := parseSemver(body.TagName)
	if !ok || !c.current.less(latest) {
		return
	}

	c.mu.Lock()
	c.latest = &Release{Version: strings.TrimPrefix(body.TagName, "v"), URL: body.HTMLURL}
	c.mu.Unlock()
	log.Printf("update check: new version %s available at %s", body.TagName, body.HTMLURL)
}

type semver struct {
	major, minor, patch int
	prerelease          string
}

func parseSemver(v string) (semver, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	var s semver
	if i := strings.IndexByte(v, '-'); i >= 0 {
		s.prerelease = v[i+1:]
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return semver{}, false
	}
	nums := [3]int{}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return semver{}, false
		}
		nums[i] = n
	}
	s.major, s.minor, s.patch = nums[0], nums[1], nums[2]
	return s, true
}

func (s semver) less(o semver) bool {
	if s.major != o.major {
		return s.major < o.major
	}
	if s.minor != o.minor {
		return s.minor < o.minor
	}
	if s.patch != o.patch {
		return s.patch < o.patch
	}
	// Same triple: a prerelease (e.g. 0.5.0-dev) precedes the release (0.5.0).
	return s.prerelease != "" && o.prerelease == ""
}
