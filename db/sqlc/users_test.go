package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/valkyraycho/bank_project/utils"
)

func TestCreateUser(t *testing.T) {
	randomUser(t)
}

func TestGetUser(t *testing.T) {
	user := randomUser(t)

	gotUser, err := testStore.GetUser(context.Background(), user.Username)
	require.NoError(t, err)
	require.NotEmpty(t, gotUser)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Equal(t, user.HashedPassword, gotUser.HashedPassword)
	require.Equal(t, user.Role, gotUser.Role)
	require.Equal(t, user.CreatedAt, gotUser.CreatedAt)
	require.Equal(t, user.PasswordChangedAt, gotUser.PasswordChangedAt)
}

func TestDeleteUser(t *testing.T) {
	user := randomUser(t)
	err := testStore.DeleteUser(context.Background(), user.ID)
	require.NoError(t, err)

	deletedUser, err := testStore.GetUser(context.Background(), user.Username)
	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, deletedUser)
}

func TestUpdateUserAll(t *testing.T) {
	user := randomUser(t)

	newUsername := utils.RandomName()
	newFullName := utils.RandomName()
	newEmail := utils.RandomEmail()
	newHashedPassword, err := utils.HashPassword(utils.RandomString(8))
	require.NoError(t, err)

	args := UpdateUserParams{
		ID: user.ID,
		Username: pgtype.Text{
			String: newUsername,
			Valid:  true,
		},
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
		HashedPassword: pgtype.Text{
			String: newHashedPassword,
			Valid:  true,
		},
		PasswordChangedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	updatedUser, err := testStore.UpdateUser(context.Background(), args)
	require.NoError(t, err)

	require.NotEmpty(t, updatedUser)
	require.Equal(t, newUsername, updatedUser.Username)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.WithinDuration(t, time.Now(), updatedUser.PasswordChangedAt, time.Duration(2*time.Second))
}

func TestUpdateOnlyUsername(t *testing.T) {
	user := randomUser(t)

	newUsername := utils.RandomName()

	args := UpdateUserParams{
		ID: user.ID,
		Username: pgtype.Text{
			String: newUsername,
			Valid:  true,
		},
	}

	updatedUser, err := testStore.UpdateUser(context.Background(), args)
	require.NoError(t, err)

	require.NotEmpty(t, updatedUser)
	require.Equal(t, newUsername, updatedUser.Username)
	require.Equal(t, user.FullName, updatedUser.FullName)
	require.Equal(t, user.Email, updatedUser.Email)
	require.Equal(t, user.HashedPassword, updatedUser.HashedPassword)
	require.WithinDuration(t, user.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Duration(2*time.Second))
}

func TestUpdateOnlyEmail(t *testing.T) {
	user := randomUser(t)

	newEmail := utils.RandomEmail()

	args := UpdateUserParams{
		ID: user.ID,
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
	}

	updatedUser, err := testStore.UpdateUser(context.Background(), args)
	require.NoError(t, err)

	require.NotEmpty(t, updatedUser)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, user.Username, updatedUser.Username)
	require.Equal(t, user.FullName, updatedUser.FullName)
	require.Equal(t, user.HashedPassword, updatedUser.HashedPassword)
	require.WithinDuration(t, user.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Duration(2*time.Second))
}

func TestUpdateOnlyFullName(t *testing.T) {
	user := randomUser(t)

	newFullName := utils.RandomName()

	args := UpdateUserParams{
		ID: user.ID,
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
	}

	updatedUser, err := testStore.UpdateUser(context.Background(), args)
	require.NoError(t, err)

	require.NotEmpty(t, updatedUser)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, user.Username, updatedUser.Username)
	require.Equal(t, user.Email, updatedUser.Email)
	require.Equal(t, user.HashedPassword, updatedUser.HashedPassword)
	require.WithinDuration(t, user.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Duration(2*time.Second))
}

func TestUpdateOnlyHashedPassword(t *testing.T) {
	user := randomUser(t)

	newHashedPassword, err := utils.HashPassword(utils.RandomString(8))
	require.NoError(t, err)

	args := UpdateUserParams{
		ID: user.ID,
		HashedPassword: pgtype.Text{
			String: newHashedPassword,
			Valid:  true,
		},
	}

	updatedUser, err := testStore.UpdateUser(context.Background(), args)
	require.NoError(t, err)

	require.NotEmpty(t, updatedUser)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, user.Username, updatedUser.Username)
	require.Equal(t, user.FullName, updatedUser.FullName)
	require.Equal(t, user.Email, updatedUser.Email)
	require.WithinDuration(t, user.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Duration(2*time.Second))
}

func randomUser(t *testing.T) User {
	hashedPassword, err := utils.HashPassword(utils.RandomString(8))
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	args := CreateUserParams{
		Username:       utils.RandomName(),
		HashedPassword: hashedPassword,
		Email:          utils.RandomEmail(),
		FullName:       utils.RandomName(),
	}
	user, err := testStore.CreateUser(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, args.HashedPassword, user.HashedPassword)
	require.Equal(t, args.Username, user.Username)
	require.Equal(t, args.FullName, user.FullName)
	require.Equal(t, args.Email, user.Email)

	return user
}
