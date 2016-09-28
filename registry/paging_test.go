package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestPages(t *testing.T) {
	http.HandleFunc("/v2/", pingHandler)
	http.HandleFunc("/v2/example.com/image/tags/list", tagsHandler)
	http.HandleFunc("/v2/_catalog", catalogHandler)
	go http.ListenAndServe(":9999", nil)

	hub, err := New("http://localhost:9999/", "", "")
	if err != nil {
		t.Fatal(err)
	}
	ts, err := hub.Tags("example.com/image")
	checkItems(t, tags, ts)

	rs, err := hub.Repositories()
	checkItems(t, repos, rs)
}

const perPage = 5

var tags = []string{
	"tag1",
	"tag2",
	"tag3",
	"tag4",
	"tag5",
	"tag6",
	"tag7",
	"tag8",
	"tag9",
	"tag10",
	"tag11",
	"tag12",
}

var repos = []string{
	"image1",
	"image2",
	"image3",
	"image4",
	"image5",
	"image6",
	"image7",
	"image8",
	"image9",
	"image10",
	"image11",
	"image12",
	"image13",
	"image14",
	"image15",
	"image16",
	"image17",
}

func checkItems(t *testing.T, expected []string, got []string) {
	same := true
	if len(expected) != len(got) {
		same = false
	} else {
		for i, t := range expected {
			if got[i] != t {
				same = false
				break
			}
		}
	}
	if !same {
		t.Fatalf("Received tags did not match: got %#v, expected %#v", got, expected)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func tagsHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Tags []string
	}
	pagedResult(w, r, tags, func(page []string) interface{} {
		return response{page}
	})
}

func catalogHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Repositories []string
	}
	pagedResult(w, r, repos, func(page []string) interface{} {
		return response{page}
	})
}

func pagedResult(w http.ResponseWriter, r *http.Request, items []string, mkResult func([]string) interface{}) {
	last := r.URL.Query().Get("last")
	nStr := r.URL.Query().Get("n")
	basePath := r.URL.Path

	var n int = perPage
	if nStr != "" {
		if _, err := fmt.Sscanf(nStr, "%d", &n); err != nil {
			return
		}
	}
	i := 0
	if last != "" {
		for ; i < len(items); i++ {
			if items[i] == last {
				break
			}
		}
		if i >= len(items) {
			http.Error(w, "last not found", http.StatusBadRequest)
			return
		}
	}
	if (i + n) < len(items) {
		qv := url.Values{}
		qv.Set("n", fmt.Sprintf("%d", n))
		qv.Set("last", items[i+n])
		nextURL := &url.URL{
			Scheme:   "http",
			Host:     "localhost:9999",
			Path:     basePath,
			RawQuery: qv.Encode(),
		}
		w.Header().Add("Link", fmt.Sprintf(`<%s>; title="next page"; rel="next"; type="application/json"`, nextURL))
	} else {
		n = len(items) - i
	}
	json.NewEncoder(w).Encode(mkResult(items[i : i+n]))
}
