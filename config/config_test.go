// +build !integration

package config

import (
	"testing"
)

func TestEnterpriseConfig_GetBaseUrl(t *testing.T) {
	var baseUrlTests = []struct {
		Url         string
		ExpectUrl   bool
		ExpectError bool
	}{
		{"", false, false},
		{"http://[::1]a", false, true},
		{"https://api.github.com", true, false},
	}

	for _, tt := range baseUrlTests {
		subject := EnterpriseConfig{tt.Url, ""}

		actual, err := subject.GetBaseUrl()
		hasUrl := actual != nil
		hasErr := err != nil

		if hasUrl != tt.ExpectUrl {
			t.Errorf("GetBaseUrl(%q) expected to have url? %v, got it? %v", tt.Url, tt.ExpectUrl, hasUrl)
		}

		if hasErr != tt.ExpectError {
			t.Errorf("GetBaseUrl(%q) expected to have err? %v, got it? %v", tt.Url, tt.ExpectError, hasErr)
		}
	}
}

func TestEnterpriseConfig_GetUploadUrl(t *testing.T) {
	var uploadUrlTests = []struct {
		Url         string
		ExpectUrl   bool
		ExpectError bool
	}{
		{"", false, false},
		{"http://[::1]a", false, true},
		{"https://api.github.com", true, false},
	}

	for _, tt := range uploadUrlTests {
		subject := EnterpriseConfig{"", tt.Url}

		actual, err := subject.GetUploadUrl()
		hasUrl := actual != nil
		hasErr := err != nil

		if hasUrl != tt.ExpectUrl {
			t.Errorf("GetUploadUrl(%q) expected to have url? %v, got it? %v %v", tt.Url, tt.ExpectUrl, hasUrl, actual)
		}

		if hasErr != tt.ExpectError {
			t.Errorf("GetUploadUrl(%q) expected to have err? %v, got it? %v", tt.Url, tt.ExpectError, hasErr)
		}
	}
}
