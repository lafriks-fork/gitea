// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"code.gitea.io/gitea/modules/auth"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/validation"
)

const (
	tplBranch base.TplName = "repo/branch"
)

// Branches render repository branch page
func Branches(ctx *context.Context) {
	ctx.Data["Title"] = "Branches"
	ctx.Data["IsRepoToolbarBranches"] = true

	brs, err := ctx.Repo.GitRepo.GetBranches()
	if err != nil {
		ctx.Handle(500, "repo.Branches(GetBranches)", err)
		return
	} else if len(brs) == 0 {
		ctx.Handle(404, "repo.Branches(GetBranches)", nil)
		return
	}

	ctx.Data["Branches"] = brs
	ctx.HTML(200, tplBranch)
}

// CreateBranch creates new branch in repository
func CreateBranch(ctx *context.Context, form auth.NewBranchForm) {
	if !ctx.Repo.CanCreateBranch() {
		ctx.Handle(404, "CreateBranch", nil)
	}

	newBranchName := form.Name
	if validation.GitRefNamePattern.MatchString(newBranchName) {
		ctx.Flash.Error(ctx.Tr("form.NewBranchName") + ctx.Tr("form.git_ref_name_error"))
		ctx.Redirect(ctx.Repo.RepoLink + "/src/" + ctx.Repo.BranchName)
		return
	}

	if _, err := ctx.Repo.Repository.GetBranch(newBranchName); err == nil {
		ctx.Flash.Error(ctx.Tr("repo.branch.branch_already_exists", newBranchName))
		ctx.Redirect(ctx.Repo.RepoLink + "/src/" + ctx.Repo.BranchName)
		return
	}

	if _, err := ctx.Repo.GitRepo.GetTag(newBranchName); err == nil {
		ctx.Flash.Error(ctx.Tr("repo.branch.tag_already_exists", newBranchName))
		ctx.Redirect(ctx.Repo.RepoLink + "/src/" + ctx.Repo.BranchName)
		return
	}

	var err error
	if ctx.Repo.IsViewBranch {
		err = ctx.Repo.Repository.CreateNewBranch(ctx.User, ctx.Repo.BranchName, newBranchName)
	} else {
		err = ctx.Repo.Repository.CreateNewBranchFromCommit(ctx.User, ctx.Repo.BranchName, newBranchName)
	}
	if err != nil {
		ctx.Handle(500, "CreateNewBranch", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("repo.branch.create_success", newBranchName))
	ctx.Redirect(ctx.Repo.RepoLink + "/src/" + newBranchName)
}
