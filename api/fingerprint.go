package api

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (a *API) GenerateFingerprintString(location string) (string, error) {
	fingerprint := a.GenerateFingerprint(location)

	fp, err := json.Marshal(fingerprint)
	if err != nil {
		return "", err
	}

	fpString := url.QueryEscape(base64.StdEncoding.EncodeToString(fp))
	return fpString, nil
}

func (a *API) GenerateFingerprint(location string) DeviceFingerprint {
	ua := a.GetUserAgent()

	return DeviceFingerprint{
		UserAgent: strings.ToLower(ua.String),
		Browser: Browser{
			MajorVersion: strconv.Itoa(ua.VersionNo.Major),
			Name:         ua.Name,
		},
		Canvas:        "6bdc41824a1a2337d441c497c083b42be8e88c4d36a6b11811da2322d4f1242b",
		Checksum:      "b76ebc6b4e103c279d889d48befac3b505a185d6e4b57c90fde1af2c50656527",
		CookieEnabled: "true",
		Fonts: Fonts{
			InstalledFonts: []string{
				"Arial",
				"Arial Black",
				"Arial Narrow",
				"Arial Rounded MT Bold",
				"Comic Sans MS",
				"Courier",
				"Courier New",
				"Georgia",
				"Impact",
				"Papyrus",
				"Tahoma",
				"Times",
				"Times New Roman",
				"Trebuchet MS",
				"Verdana",
			},
		},
		FormFields: FormFields{
			URL:        location,
			FormInputs: []any{},
		},
		JavaEnabled: "false",
		Language: Language{
			Language: "en-US",
		},
		Latency: Latency{
			RequestTime:    "0",
			NetworkLatency: "4",
		},
		Plugins: Plugins{
			InstalledPlugins: []string{"internal-pdf-viewer", "mhjfbmdgcfjbbpaeojofohoefgiehjai"},
		},
		Rt: []string{
			"0",
			"0",
			"3",
			"0",
			"2",
			"1",
			"0",
			"0",
			"0",
			"0",
			"0",
			"6",
			"0",
			"1",
			"0",
			"0",
			"0",
			"0",
			"0",
			"0",
			"0",
			"0",
			"13",
		},
		Screen: Screen{
			ColorDepth:           "30",
			FontSmoothingEnabled: "true",
			Height:               "1440",
			Width:                "2560",
		},
		System: System{
			OperatingSystem: "Mac OS X",
			OSVersion:       ua.OSVersion,
			Platform:        "MacIntel",
		},
		Tcn:       "13239",
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
		Timezone: Timezone{
			Timezone: "-08:00",
		},
		TrueBrowser: "Safari",
		Version:     "2.0.0",
	}
}

type DeviceFingerprint struct {
	UserAgent     string     `json:"userAgent"`
	Timezone      Timezone   `json:"timezone"`
	Screen        Screen     `json:"screen"`
	Language      Language   `json:"language"`
	Fonts         Fonts      `json:"fonts"`
	Plugins       Plugins    `json:"plugins"`
	CookieEnabled string     `json:"cookieEnabled"`
	JavaEnabled   string     `json:"javaEnabled"`
	Canvas        string     `json:"canvas"`
	TrueBrowser   string     `json:"trueBrowser"`
	Tcn           string     `json:"tcn"`
	Browser       Browser    `json:"browser"`
	System        System     `json:"system"`
	Latency       Latency    `json:"latency"`
	FormFields    FormFields `json:"formFields"`
	Version       string     `json:"version"`
	Timestamp     string     `json:"timestamp"`
	Checksum      string     `json:"checksum"`
	Rt            []string   `json:"rt"`
}

type Browser struct {
	Name         string `json:"name"`
	MajorVersion string `json:"majorVersion"`
}

type Fonts struct {
	InstalledFonts []string `json:"installedFonts"`
}

type FormFields struct {
	URL        string `json:"url"`
	FormInputs []any  `json:"formInputs"`
}

type Language struct {
	Language string `json:"language"`
}

type Latency struct {
	RequestTime    string `json:"requestTime"`
	NetworkLatency string `json:"networkLatency"`
}

type Plugins struct {
	InstalledPlugins []string `json:"installedPlugins"`
}

type Screen struct {
	Width                string `json:"width"`
	Height               string `json:"height"`
	ColorDepth           string `json:"colorDepth"`
	FontSmoothingEnabled string `json:"fontSmoothingEnabled"`
}

type System struct {
	OperatingSystem string `json:"operatingSystem"`
	OSVersion       string `json:"osVersion"`
	Platform        string `json:"platform"`
}

type Timezone struct {
	Timezone string `json:"timezone"`
}
