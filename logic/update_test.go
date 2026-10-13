package logic

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestForkRelease(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		body, want string
		wantError  bool
	}{
		{"ready", 200, `{"tag_name":"v1.2.3","assets":[{"name":"gf2gacha.exe","state":"uploaded"}]}`, "v1.2.3", false},
		{"no releases", 404, `{"message":"Not Found"}`, "", false},
		{"no executable", 200, `{"tag_name":"v1.2.3","assets":[{"name":"gf2gacha.zip","state":"uploaded"}]}`, "", false},
		{"upload pending", 200, `{"tag_name":"v1.2.3","assets":[{"name":"gf2gacha.exe","state":"starter"}]}`, "", false},
		{"rate limited", 403, `{"message":"rate limited"}`, "", true},
		{"invalid response", 200, `invalid`, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Accept") != "application/vnd.github+json" || r.Header.Get("User-Agent") == "" {
					t.Error("missing GitHub API request headers")
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			got, err := getLatestVersion(server.Client(), server.URL)
			if got != tc.want || (err != nil) != tc.wantError {
				t.Fatalf("got (%q, %v), want (%q, error=%v)", got, err, tc.want, tc.wantError)
			}
		})
	}
}

func TestReleaseDownloadIsScopedToFork(t *testing.T) {
	for _, version := range []string{"v1.2.3", "release/v1.2.3"} {
		got, err := releaseDownloadURL(version)
		want := "https://github.com/Dewcat/gf2gacha/releases/download/v1.2.3/gf2gacha.exe"
		if version == "release/v1.2.3" {
			want = "https://github.com/Dewcat/gf2gacha/releases/download/release%2Fv1.2.3/gf2gacha.exe"
		}
		if err != nil || got != want {
			t.Fatalf("got (%q, %v), want %q", got, err, want)
		}
	}
	if _, err := releaseDownloadURL(" "); err == nil {
		t.Fatal("empty version must not initiate a download")
	}
}
