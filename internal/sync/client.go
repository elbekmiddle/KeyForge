// Package sync is the client side of doc section 10:
//
//	Local config -> Sync Engine -> API -> Database
//
// It is deliberately simple: last-write-wins by file mtime, no merge, no
// version vectors (doc sections 28-30 stay unbuilt, same as the doc
// itself defers them). What it does do: register/login, register this
// device, push local profiles, pull the authoritative set back, and stay
// connected over a websocket to know when to re-pull without polling.
package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/elbekmiddle/KeyForge/internal/mapping"
	"github.com/elbekmiddle/KeyForge/internal/profile"
)

const defaultBaseURL = "http://localhost:3000"

// overrideBaseURL lets a GUI settings screen change the backend URL for
// the running process without needing to touch the environment (which a
// desktop app can't easily do for itself anyway). CLI usage is
// unaffected — it only ever sets KEYFORGE_BACKEND_URL.
var overrideBaseURL string

// BaseURL returns the backend base URL: an in-process override set via
// SetBaseURLOverride, then KEYFORGE_BACKEND_URL if set, otherwise
// http://localhost:3000.
func BaseURL() string {
	if overrideBaseURL != "" {
		return overrideBaseURL
	}

	if url := os.Getenv("KEYFORGE_BACKEND_URL"); url != "" {
		return url
	}

	return defaultBaseURL
}

// SetBaseURLOverride sets the in-process backend URL override (see
// BaseURL). Passing "" clears it, falling back to the environment
// variable / default again.
func SetBaseURLOverride(url string) {
	overrideBaseURL = url
}

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: BaseURL(),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type tokenPair struct {
	UserID       string `json:"userId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Register calls POST /auth/register.
func (c *Client) Register(email, password string) (Session, error) {
	return c.authRequest("/auth/register", map[string]string{
		"email":    email,
		"password": password,
	})
}

// Login calls POST /auth/login.
func (c *Client) Login(email, password string) (Session, error) {
	return c.authRequest("/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
}

// Refresh calls POST /auth/refresh.
func (c *Client) Refresh(refreshToken string) (Session, error) {
	return c.authRequest("/auth/refresh", map[string]string{
		"refreshToken": refreshToken,
	})
}

func (c *Client) authRequest(path string, body map[string]string) (Session, error) {
	var pair tokenPair

	if err := c.doJSON(http.MethodPost, path, "", body, &pair); err != nil {
		return Session{}, err
	}

	return Session{
		UserID:       pair.UserID,
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}

// RegisterDevice calls POST /devices (doc section 8).
func (c *Client) RegisterDevice(accessToken, guid, machineFingerprint, name string) error {
	body := map[string]string{
		"guid":               guid,
		"machineFingerprint": machineFingerprint,
	}
	if name != "" {
		body["name"] = name
	}

	return c.doJSON(http.MethodPost, "/devices", accessToken, body, nil)
}

type syncMapping struct {
	From string   `json:"from"`
	To   []string `json:"to"`
}

type syncMode struct {
	ID       uint8         `json:"id"`
	Name     string        `json:"name"`
	Mappings []syncMapping `json:"mappings"`
}

type syncProfile struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	App       string     `json:"app"`
	Modes     []syncMode `json:"modes"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type syncRequest struct {
	DeviceGUID string        `json:"deviceGuid"`
	Profiles   []syncProfile `json:"profiles"`
}

type syncResponse struct {
	Accepted []string        `json:"accepted"`
	Skipped  []string        `json:"skipped"`
	Profiles []remoteProfile `json:"profiles"`
}

type remoteProfile struct {
	ClientID string     `json:"clientId"`
	Name     string     `json:"name"`
	App      string     `json:"app"`
	Modes    []syncMode `json:"modes"`
}

// PushProfiles uploads every local profile for guid, then returns the
// authoritative set the server now has for this device (which may
// include profiles from other clients/dashboards it rejected or that
// were already newer).
func (c *Client) PushProfiles(
	accessToken, guid string,
	profiles []profile.Profile,
) (accepted, skipped []string, remote []profile.Profile, err error) {
	req := syncRequest{DeviceGUID: guid}

	for _, p := range profiles {
		modTime, statErr := profile.ModTime(guid, p.ID)
		if statErr != nil {
			modTime = time.Now()
		}

		req.Profiles = append(req.Profiles, toSyncProfile(p, modTime))
	}

	var resp syncResponse
	if err := c.doJSON(http.MethodPost, "/profiles/sync", accessToken, req, &resp); err != nil {
		return nil, nil, nil, err
	}

	for _, rp := range resp.Profiles {
		remote = append(remote, fromRemoteProfile(rp))
	}

	return resp.Accepted, resp.Skipped, remote, nil
}

// PullProfiles fetches the authoritative profile set for guid.
func (c *Client) PullProfiles(accessToken, guid string) ([]profile.Profile, error) {
	var remote []remoteProfile

	path := fmt.Sprintf("/profiles?deviceGuid=%s", guid)
	if err := c.doJSON(http.MethodGet, path, accessToken, nil, &remote); err != nil {
		return nil, err
	}

	profiles := make([]profile.Profile, 0, len(remote))
	for _, rp := range remote {
		profiles = append(profiles, fromRemoteProfile(rp))
	}

	return profiles, nil
}

func toSyncProfile(p profile.Profile, updatedAt time.Time) syncProfile {
	modes := make([]syncMode, 0, len(p.Modes))

	for _, m := range p.Modes {
		mappings := make([]syncMapping, 0, len(m.Mappings))
		for _, mp := range m.Mappings {
			mappings = append(mappings, syncMapping{From: mp.From, To: mp.To})
		}

		modes = append(modes, syncMode{ID: m.ID, Name: m.Name, Mappings: mappings})
	}

	return syncProfile{
		ID:        p.ID,
		Name:      p.Name,
		App:       p.App,
		Modes:     modes,
		UpdatedAt: updatedAt,
	}
}

func fromRemoteProfile(rp remoteProfile) profile.Profile {
	modes := make([]profile.Mode, 0, len(rp.Modes))

	for _, m := range rp.Modes {
		mappings := make([]mapping.Mapping, 0, len(m.Mappings))
		for _, mp := range m.Mappings {
			mappings = append(mappings, mapping.Mapping{From: mp.From, To: mp.To})
		}

		modes = append(modes, profile.Mode{ID: m.ID, Name: m.Name, Mappings: mappings})
	}

	return profile.Profile{
		ID:    rp.ClientID,
		Name:  rp.Name,
		App:   rp.App,
		Modes: modes,
	}
}

func (c *Client) doJSON(method, path, accessToken string, body any, out any) error {
	var reader io.Reader

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("sync: failed to encode request: %w", err)
		}

		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("sync: failed to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("sync: request to %s failed: %w", path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("sync: failed to read response from %s: %w", path, err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("sync: %s returned %d: %s", path, resp.StatusCode, string(data))
	}

	if out == nil || len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("sync: failed to parse response from %s: %w", path, err)
	}

	return nil
}
