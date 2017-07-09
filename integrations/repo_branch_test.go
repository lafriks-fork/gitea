// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"net/http"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/Unknwon/i18n"
	"github.com/stretchr/testify/assert"
)

func testCreateBranch(t *testing.T, session *TestSession, user, repo, oldRefName, newBranchName string, expectedStatus int) string {
	req := NewRequest(t, "GET", path.Join(user, repo, "branches/_new", oldRefName)+"?name="+url.QueryEscape(newBranchName))
	resp := session.MakeRequest(t, req, expectedStatus)
	if expectedStatus != http.StatusFound {
		return ""
	}
	return RedirectURL(t, resp)
}

func TestCreateBranch(t *testing.T) {
	tests := []struct {
		OldBranchOrCommit string
		NewBranch         string
		CreateRelease     string
		FlashMessage      string
		ExpectedStatus    int
	}{
		{
			OldBranchOrCommit: "master",
			NewBranch:         "feature/test1",
			ExpectedStatus:    http.StatusFound,
			FlashMessage:      i18n.Tr("en", "repo.branch.create_success", "feature/test1"),
		},
		{
			OldBranchOrCommit: "master",
			NewBranch:         "feature=test1",
			ExpectedStatus:    http.StatusFound,
			FlashMessage:      i18n.Tr("en", "form.NewBranchName") + i18n.Tr("en", "form.git_ref_name_error"),
		},
		{
			OldBranchOrCommit: "master",
			NewBranch:         "master",
			ExpectedStatus:    http.StatusFound,
			FlashMessage:      i18n.Tr("en", "repo.branch.branch_already_exists", "master"),
		},
		{
			OldBranchOrCommit: "acd1d892867872cb47f3993468605b8aa59aa2e0",
			NewBranch:         "feature/test2",
			ExpectedStatus:    http.StatusNotFound,
		},
		{
			OldBranchOrCommit: "65f1bf27bc3bf70f64657658635e66094edbcb4d",
			NewBranch:         "feature/test3",
			ExpectedStatus:    http.StatusFound,
			FlashMessage:      i18n.Tr("en", "repo.branch.create_success", "feature/test3"),
		},
		{
			OldBranchOrCommit: "master",
			NewBranch:         "v1.0.0",
			CreateRelease:     "v1.0.0",
			ExpectedStatus:    http.StatusFound,
			FlashMessage:      i18n.Tr("en", "repo.branch.tag_already_exists", "v1.0.0"),
		},
		{
			OldBranchOrCommit: "v1.0.0",
			NewBranch:         "feature/test4",
			CreateRelease:     "v1.0.0",
			ExpectedStatus:    http.StatusFound,
			FlashMessage:      i18n.Tr("en", "repo.branch.create_success", "feature/test4"),
		},
	}
	for _, test := range tests {
		prepareTestEnv(t)
		session := loginUser(t, "user2")
		if test.CreateRelease != "" {
			createNewRelease(t, session, "/user2/repo1", test.CreateRelease, test.CreateRelease, false, false)
		}
		redirectURL := testCreateBranch(t, session, "user2", "repo1", test.OldBranchOrCommit, test.NewBranch, test.ExpectedStatus)
		if test.ExpectedStatus == http.StatusFound {
			req := NewRequest(t, "GET", redirectURL)
			resp := session.MakeRequest(t, req, http.StatusOK)
			htmlDoc := NewHTMLParser(t, resp.Body)
			assert.Equal(t,
				test.FlashMessage,
				strings.TrimSpace(htmlDoc.doc.Find(".ui.message").Text()),
			)
		}
	}
}
