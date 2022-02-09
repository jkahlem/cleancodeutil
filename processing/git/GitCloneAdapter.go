package git

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"returntypes-langserver/common/debug/log"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	httpproto "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

var ptransport *progressTransport

// Clones a repository defined using the options to path.
//
// The large part of the code here is copied from https://gist.github.com/tyru/82cae8bad2b116f442d08eeef456e23e
// The code was adapted to support go-git v5, be usable in a simple method call, stop reporting go routine
// on ending and use the log.Info logger.
func Clone(path string, options *git.CloneOptions) error {
	fs := osfs.New(path)
	// Git objects storer
	dot, err := fs.Chroot(".git")
	if err != nil {
		return err
	}
	s := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())
	if err != nil {
		return err
	}

	reporter := &progressReporter{total: 0, current: 0, finished: false}
	setProgressReporter(reporter)

	end := make(chan bool, 1)
	if isProtocolSupportedForProgressReport(options.URL) {
		// Report the progress every second until the clone process ends
		go func() {
			for {
				select {
				case <-end:
					return
				case <-time.Tick(time.Second):
					break
				}
				reporter.Report()
			}
		}()
	} else {
		log.Info("Protocol of %s is not supported for progress reporting. Therefore, there might be no information about the cloning/fetching progress.\n", options.URL)
	}

	_, err = git.Clone(&progressStorer{Storer: s, progress: reporter}, fs, options)
	end <- true
	return err
}

func isProtocolSupportedForProgressReport(url string) bool {
	if strings.HasPrefix(url, "http:") || strings.HasPrefix(url, "https:") {
		return true
	}
	return false
}

func setProgressReporter(reporter *progressReporter) {
	initializeHttpClient()
	ptransport.progress = reporter
}

func initializeHttpClient() {
	if ptransport == nil {
		ptransport = &progressTransport{
			RoundTripper: http.DefaultTransport,
		}
		httpClient := httpproto.NewClient(&http.Client{
			Transport: ptransport,
		})
		client.InstallProtocol("http", httpClient)
		client.InstallProtocol("https", httpClient)
	}
}

type progressReporter struct {
	total    uint32
	current  uint32
	finished bool
}

func (p *progressReporter) SetTotal(total uint32) {
	p.total = total
}

func (p *progressReporter) Inc() {
	p.current++
}

func (p *progressReporter) Report() {
	if p.total != 0 && !p.finished {
		percent := int(float64(p.current) / float64(p.total) * 100)
		log.Info("(%d%%) %d/%d\n", percent, p.current, p.total)
		if p.current == p.total {
			p.finished = true
		}
	}
}

type progressStorer struct {
	storage.Storer
	progress *progressReporter
}

func (s *progressStorer) SetEncodedObject(o plumbing.EncodedObject) (plumbing.Hash, error) {
	hash, err := s.Storer.SetEncodedObject(o)
	s.progress.Inc()
	return hash, err
}

type progressTransport struct {
	http.RoundTripper
	progress *progressReporter
}

func (t *progressTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.RoundTripper.RoundTrip(req)
	if req.Method == "POST" && strings.HasSuffix(req.URL.String(), "/git-upload-pack") {
		if err := t.extractTotalInHeader(res); err != nil {
			return nil, err
		}
	}
	return res, err
}

func (t *progressTransport) extractTotalInHeader(res *http.Response) error {
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()
	res.Body = ioutil.NopCloser(bytes.NewBuffer(content))

	sc := pktline.NewScanner(bytes.NewBuffer(content))
	hi := 0
	var header [12]byte
	for sc.Scan() {
		b := sc.Bytes()
		if len(b) > 0 && b[0] == '\x01' {
			for i := 1; i < len(b) && hi < 12; i++ {
				header[hi] = b[i]
				hi++
			}
			if hi >= 12 {
				total := binary.BigEndian.Uint32(header[8:12])
				t.progress.SetTotal(total)
			}
		}
	}
	if sc.Err() != nil {
		return sc.Err()
	}
	return nil
}
