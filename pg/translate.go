package pg

import (
	"fmt"
	"strconv"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type WALTranslations struct {
	Major        string
	Directory    string
	Lsn          string
	Wal          string
	WalDump      string
	Queries      WALQueries
}

type WALQueries struct {
	Lag    string
}

func Translate(pgMajor string) (WALTranslations, error) {
	log.Debug().Str("pg-major", pgMajor).Msg("translating wal interactions based on supplied postgres major")

	var translations WALTranslations

	var parsedMajor, err = strconv.ParseFloat(pgMajor, 64)
	if err != nil {
		return translations, errors.Wrap(err, "failed to parse major")
	}

	if parsedMajor < 9.6 {
		return translations, fmt.Errorf("pg majors < 9.6 are unsupported (supplied was %s)", pgMajor)
	}

	translations = WALTranslations{}
	{
		translations.Major = pgMajor
		queries := WALQueries{}
		if parsedMajor < 10 {
			translations.Directory = "pg_xlog"
			translations.Lsn = "location"
			translations.Wal = "xlog"
			translations.WalDump = "pg_xlogdump"
			queries.Lag = "SELECT timeline_id, redo_location, pg_last_xlog_replay_location() FROM pg_control_checkpoint()"
		} else {
			translations.Directory = "pg_wal"
			translations.Lsn = "lsn"
			translations.Wal = "wal"
			translations.WalDump = "pg_waldump"
			queries.Lag = "SELECT timeline_id, redo_lsn, pg_last_wal_receive_lsn() FROM pg_control_checkpoint()"
		}

		translations.Queries = queries
	}

	// log.Error().Err(err).Msg("failed the thing")
	log.Debug().
		Str("major", translations.Major).
		Str("directory", translations.Directory).
		Str("lsn", translations.Lsn).
		Str("wal", translations.Wal).
		Msg("wal translations ready")

	return translations, nil
}
