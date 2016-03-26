package git

import (
	"errors"
	"fmt"
	"net/url"
)

type Repo struct {
	URL    string `json:"url"`
	Branch string `json:"branch"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
}

func (r *Repo) Validate() (errs []error) {
	if r.URL == "" {
		errs = append(errs, errors.New("URL required"))
	}
	if r.Branch == "" {
		errs = append(errs, errors.New("Branch required"))
	}
	if (r.User == "") != (r.Pass == "") {
		errs = append(errs, errors.New("User and pass are both required or neither required"))
	}
	return
}

// This is needed because git doesn't offer an option to provide credentials via stdin
func (r *Repo) URLWithCredentials() (string, error) {
	properUrl, err := url.Parse(r.URL)
	if err != nil {
		return "", fmt.Errorf("Invalid URL: %v", err)
	}
	if r.User != "" {
		// TODO: I'm doing this because sadly git doesn't allow stdin-based passwords
		properUrl.User = url.UserPassword(r.User, r.Pass)
	}
	return properUrl.String(), nil
}
