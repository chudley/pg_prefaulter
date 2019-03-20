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
	Lag             string
	ByteLagPrimary  string
	ByteLagFollower string
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


	var byteLagPrimaryFmt = `SELECT
	    state,
	    sync_state,
	    (pg_%[2]s_%[1]s_diff(sent_%[1]s, write_%[1]s))::FLOAT8 AS durability_lag_bytes,
	    (pg_%[2]s_%[1]s_diff(sent_%[1]s, flush_%[1]s))::FLOAT8 AS flush_lag_bytes,
	    (pg_%[2]s_%[1]s_diff(sent_%[1]s, replay_%[1]s))::FLOAT8 AS visibility_lag_bytes,
	    COALESCE(EXTRACT(EPOCH FROM '0'::INTERVAL), 0.0)::FLOAT8 AS visibility_lag_ms
	    FROM
	    pg_catalog.pg_stat_replication
	    ORDER BY visibility_lag_bytes
	    LIMIT 1`

	var byteLagFollowerFmt = `SELECT
	    'receiving' AS state,
	    'applying' AS sync_state,
	    0.0::FLOAT8 AS durability_lag_bytes,
	    0.0::FLOAT8 AS flush_lag_bytes,
	    COALESCE((pg_%[2]s_%[1]s_diff(pg_last_%[2]s_receive_%[1]s(), pg_last_%[2]s_replay_%[1]s()))::FLOAT8, 0.0)::FLOAT8 AS visibility_lag_bytes,
	    COALESCE(EXTRACT(EPOCH FROM (NOW() - pg_last_xact_replay_timestamp())::INTERVAL), 0.0)::FLOAT8 AS visibility_lag_ms
	    LIMIT 1`

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

		queries.ByteLagPrimary = fmt.Sprintf(byteLagPrimaryFmt, translations.Lsn, translations.Wal)
		queries.ByteLagFollower = fmt.Sprintf(byteLagFollowerFmt, translations.Lsn, translations.Wal)

		translations.Queries = queries
	}

	log.Debug().
		Str("major", translations.Major).
		Str("directory", translations.Directory).
		Str("lsn", translations.Lsn).
		Str("wal", translations.Wal).
		Msg("wal translations ready")

	return translations, nil
}
