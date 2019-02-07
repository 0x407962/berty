package migrations

import (
	"berty.tech/core/sql/migrations/v0001init"
	"berty.tech/core/sql/migrations/v0002pushtokens"
	"berty.tech/core/sql/migrations/v0003notificationssettings"
	"berty.tech/core/sql/migrations/v0004conversationinfos"
	"github.com/go-gormigrate/gormigrate"
)

func GetMigrations() []*gormigrate.Migration {
	migrations := []*gormigrate.Migration{}

	migrations = append(migrations, v0001init.GetMigration())
	migrations = append(migrations, v0002pushtokens.GetMigration())
	migrations = append(migrations, v0003notificationssettings.GetMigration())
	migrations = append(migrations, v0004conversationinfos.GetMigration())

	return migrations
}
