package store

import "github.com/blang/semver"

type migration struct {
	fromVersion   semver.Version
	toVersion     semver.Version
	migrationFunc func(execer) error
}

// migrations defines the set of migrations necessary to advance the database to the latest
// expected version.
//
// Note that the canonical schema is currently obtained by applying all migrations to an empty
// database.
var migrations = []migration{
	{semver.MustParse("0.0.0"), semver.MustParse("0.1.0"), func(e execer) error {
		_, err := e.Exec(`
			CREATE TABLE System (
				Key VARCHAR(64) PRIMARY KEY,
				Value VARCHAR(1024) NULL
			);
		`)
		if err != nil {
			return err
		}

		_, err = e.Exec(`
			CREATE TABLE gameday (
				id CHAR(26) PRIMARY KEY,
				title VARCHAR(32) NOT NULL,
				team_id CHAR(26) NOT NULL,
				scheduled_at BIGINT NOT NULL,
				created_at BIGINT NOT NULL,
				updated_at BIGINT NOT NULL
			);
		`)
		if err != nil {
			return err
		}

		// add index
		_, err = e.Exec(`
			CREATE UNIQUE INDEX gameday_team_scheduled_at ON gameday (team_id, scheduled_at);
		`)
		if err != nil {
			return err
		}

		_, err = e.Exec(`
			CREATE TABLE team (
				id CHAR(26) PRIMARY KEY,
				name VARCHAR(32) NOT NULL,
				created_at BIGINT NOT NULL,
				updated_at BIGINT NOT NULL
			);
		`)
		if err != nil {
			return err
		}
		_, err = e.Exec(`
			CREATE UNIQUE INDEX gameday_team ON team (name);
		`)
		if err != nil {
			return err
		}

		_, err = e.Exec(`
			CREATE TABLE team_member (
				id CHAR(26) PRIMARY KEY,
				team_id CHAR(26) NOT NULL,
				user_id VARCHAR(32) NOT NULL,
				label VARCHAR(32) NOT NULL,
				created_at BIGINT NOT NULL,
				updated_at BIGINT NOT NULL
			);
		`)
		if err != nil {
			return err
		}
		_, err = e.Exec(`
			CREATE UNIQUE INDEX gameday_team_user ON team_member (team_id, user_id);
		`)
		if err != nil {
			return err
		}
		return nil
	}},
}
