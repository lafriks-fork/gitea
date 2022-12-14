// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package secrets

import (
	"context"
	"fmt"

	auth_model "code.gitea.io/gitea/models/auth"
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/setting"

	"xorm.io/builder"
)

// MasterKeyProviderType is the type of master key provider
type MasterKeyProviderType string

// Types of master key providers
const (
	MasterKeyProviderTypeNone  MasterKeyProviderType = "none"
	MasterKeyProviderTypePlain MasterKeyProviderType = "plain"
)

var (
	masterKey   MasterKeyProvider
	encProvider EncryptionProvider
)

// Init initializes master key provider based on settings
func Init() error {
	switch MasterKeyProviderType(setting.MasterKeyProvider) {
	case MasterKeyProviderTypeNone:
		masterKey = NewNopMasterKeyProvider()
	case MasterKeyProviderTypePlain:
		masterKey = NewPlainMasterKeyProvider()
	default:
		return fmt.Errorf("invalid master key provider %v", setting.MasterKeyProvider)
	}

	if err := masterKey.Init(); err != nil {
		return err
	}

	encProvider = NewAesEncryptionProvider()

	return nil
}

// GenerateMasterKey generates a new master key and returns secret or secrets for unsealing
func GenerateMasterKey() ([][]byte, error) {
	return masterKey.GenerateMasterKey()
}

func Encrypt(secret []byte) ([]byte, error) {
	key, err := masterKey.GetMasterKey()
	if err != nil {
		return nil, err
	}

	if len(key) == 0 {
		return secret, nil
	}

	return encProvider.Encrypt(secret, key)
}

func EncryptString(secret string) (string, error) {
	key, err := masterKey.GetMasterKey()
	if err != nil {
		return "", err
	}

	if len(key) == 0 {
		return secret, nil
	}

	return encProvider.EncryptString(secret, key)
}

func Decrypt(enc []byte) ([]byte, error) {
	key, err := masterKey.GetMasterKey()
	if err != nil {
		return nil, err
	}

	if len(key) == 0 {
		return enc, nil
	}

	return encProvider.Decrypt(enc, key)
}

func DecryptString(enc string) (string, error) {
	key, err := masterKey.GetMasterKey()
	if err != nil {
		return "", err
	}

	if len(key) == 0 {
		return enc, nil
	}

	return encProvider.DecryptString(enc, key)
}

func InsertRepoSecret(ctx context.Context, repoID int64, key, data string) error {
	v, err := EncryptString(data)
	if err != nil {
		return err
	}
	return db.Insert(ctx, &auth_model.Secret{
		RepoID: repoID,
		Name:   key,
		Data:   v,
	})
}

func InsertOwnerSecret(ctx context.Context, ownerID int64, key, data string) error {
	v, err := EncryptString(data)
	if err != nil {
		return err
	}
	return db.Insert(ctx, &auth_model.Secret{
		OwnerID: ownerID,
		Name:    key,
		Data:    v,
	})
}

func DeleteSecretByID(ctx context.Context, id int64) error {
	_, err := db.DeleteByBean(ctx, &auth_model.Secret{ID: id})
	return err
}

func DeleteSecretsByRepoID(ctx context.Context, repoID int64) error {
	_, err := db.DeleteByBean(ctx, &auth_model.Secret{RepoID: repoID})
	return err
}

func DeleteSecretsByOwnerID(ctx context.Context, ownerID int64) error {
	_, err := db.DeleteByBean(ctx, &auth_model.Secret{ID: ownerID})
	return err
}

func FindRepoSecrets(ctx context.Context, repoID int64) ([]*auth_model.Secret, error) {
	var res []*auth_model.Secret
	return res, db.GetEngine(ctx).Where(builder.Eq{"repo_id": repoID}).Find(&res)
}

func FindOwnerSecrets(ctx context.Context, ownerID int64) ([]*auth_model.Secret, error) {
	var res []*auth_model.Secret
	return res, db.GetEngine(ctx).Where(builder.Eq{"owner_id": ownerID}).Find(&res)
}
