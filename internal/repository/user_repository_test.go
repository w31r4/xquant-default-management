package repository

import (
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"
	"xquant-default-management/internal/core"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// AnyTime is used to match any time.Time value in sqlmock
type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewUserRepository(gormDB)

		user := &core.User{
			BaseModel: core.BaseModel{ID: uuid.New()},
			Username:  "testuser",
			Password:  "password",
			Role:      "Applicant",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "users" ("id","created_at","updated_at","deleted_at","username","password","role") VALUES ($1,$2,$3,$4,$5,$6,$7)`)).
			WithArgs(sqlmock.AnyArg(), AnyTime{}, AnyTime{}, nil, user.Username, user.Password, user.Role).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewUserRepository(gormDB)

		user := &core.User{
			BaseModel: core.BaseModel{ID: uuid.New()},
			Username:  "testuser",
			Password:  "password",
			Role:      "Applicant",
		}
		dbErr := errors.New("db error")

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "users" ("id","created_at","updated_at","deleted_at","username","password","role") VALUES ($1,$2,$3,$4,$5,$6,$7)`)).
			WithArgs(sqlmock.AnyArg(), AnyTime{}, AnyTime{}, nil, user.Username, user.Password, user.Role).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.Create(user)
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewUserRepository(gormDB)

		expectedUser := &core.User{
			BaseModel: core.BaseModel{ID: uuid.New()},
			Username:  "founduser",
			Password:  "hashedpassword",
			Role:      "Approver",
		}

		rows := sqlmock.NewRows([]string{"id", "username", "password", "role"}).
			AddRow(expectedUser.BaseModel.ID, expectedUser.Username, expectedUser.Password, expectedUser.Role)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
			WithArgs(expectedUser.Username, 1).
			WillReturnRows(rows)

		user, err := repo.GetByUsername(expectedUser.Username)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.BaseModel.ID, user.BaseModel.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewUserRepository(gormDB)

		username := "nonexistent"
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
			WithArgs(username, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetByUsername(username)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		// Even on error, GORM might return a non-nil user struct with zero values.
		// Depending on desired behavior, you might assert user is nil or just check the error.
		assert.NotNil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
