package main

import (
	"fmt"
	"github.com/GeoNet/msg"
	"github.com/lib/pq"
	"log"
)

// These regions must exist in the DB.
var regionIDs = []msg.RegionID{
	msg.NewZealand,
	msg.AucklandNorthland,
	msg.TongagriroBayofPlenty,
	msg.Gisborne,
	msg.HawkesBay,
	msg.Taranaki,
	msg.Wellington,
	msg.NelsonWestCoast,
	msg.Canterbury,
	msg.Fiordland,
	msg.OtagoSouthland,
}

func saveQuake(q msg.Quake) error {

	//  could build the map from the struct using https://github.com/fatih/structs (reflection)
	//
	// Geom is added with a DB trigger for each new row
	var qv = map[string]interface{}{
		`PublicID`:              q.PublicID,
		`Type`:                  q.Type,
		`AgencyID`:              q.AgencyID,
		`ModificationTime`:      q.ModificationTime,
		`Time`:                  q.Time,
		`Longitude`:             q.Longitude,
		`Latitude`:              q.Latitude,
		`Depth`:                 q.Depth,
		`DepthType`:             q.DepthType,
		`MethodID`:              q.MethodID,
		`EarthModelID`:          q.EarthModelID,
		`EvaluationMode`:        q.EvaluationMode,
		`EvaluationStatus`:      q.EvaluationStatus,
		`UsedPhaseCount`:        q.UsedPhaseCount,
		`UsedStationCount`:      q.UsedStationCount,
		`StandardError`:         q.StandardError,
		`AzimuthalGap`:          q.AzimuthalGap,
		`MinimumDistance`:       q.MinimumDistance,
		`Magnitude`:             q.Magnitude,
		`MagnitudeUncertainty`:  q.MagnitudeUncertainty,
		`MagnitudeType`:         q.MagnitudeType,
		`MagnitudeStationCount`: q.MagnitudeStationCount,
		`Site`:                  q.Site,
		`Status`:                q.Status(),
		`Quality`:               q.Quality(),
		`Deleted`:               q.Status() == `deleted`,
		`BackupSite`:            q.Site == `backup`,
		`MMI`:                   q.MMI(),
	}

	mmi := q.MMI()
	qv[`MMI`] = mmi
	qv[`Intensity`] = msg.MMIIntensity(mmi)

	// don't use time.UnixNano() for modificationTimeMicro, the zero time overflows int64.
	mtUnixMicro := q.ModificationTime.Unix()*1000000 + int64(q.ModificationTime.Nanosecond()/1000)
	qv[`ModificationTimeUnixMicro`] = mtUnixMicro

	// Add the region MMID and intensity for all regions in the DB.
	for _, v := range regionIDs {
		l, err := q.ClosestInRegion(v)
		if err != nil {
			log.Println("error finding closest locality in " + string(v))
			log.Println("setting MMID and intensity unknown.")
			qv[`MMID_`+string(v)] = 0.0
			qv[`Intensity_`+string(v)] = msg.MMIIntensity(0.0)
			continue
		}
		qv[`MMID_`+string(v)] = l.MMIDistance
		qv[`Intensity_`+string(v)] = msg.MMIIntensity(l.MMIDistance)
	}

	var insert string
	var params string
	var values []interface{}
	var i int = 1

	for k, v := range qv {
		insert = insert + k + `, `
		params = params + fmt.Sprintf("$%d, ", i)
		values = append(values, v)
		i = i + 1
	}

	locality := "'unknown'"
	c, err := q.ClosestInRegion(msg.NewZealand)
	if err == nil {
		locality = fmt.Sprintf("$$%s %s of %s$$", msg.Distance(c.Distance), msg.Compass(c.Bearing), c.Locality.Name)
	}
	insert = insert + `Locality`
	params = params + locality

	// Quake History
	txn, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = txn.Exec(`DELETE FROM haz.quakehistory WHERE PublicID = $1 AND ModificationTimeUnixMicro = $2`, q.PublicID, mtUnixMicro)
	if err != nil {
		txn.Rollback()
		return err
	}

	_, err = txn.Exec(`INSERT INTO haz.quakehistory(`+insert+`) VALUES( `+params+` )`, values...)
	if err != nil {
		txn.Rollback()
		return err
	}

	err = txn.Commit()
	if err != nil {
		return err
	}

	// Clean out old quake history
	_, err = db.Exec(`DELETE FROM haz.quakehistory WHERE time < now() - interval '365 days'`)

	// Quake
	txn, err = db.Begin()
	if err != nil {
		return err
	}

	_, err = txn.Exec(`DELETE FROM haz.quake WHERE PublicID = $1 AND ModificationTime < $2`, q.PublicID, q.ModificationTime)
	if err != nil {
		txn.Rollback()
		return err
	}

	_, err = txn.Exec(`INSERT INTO haz.quake(`+insert+`) VALUES( `+params+` )`, values...)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// a unique_violation means the new quake info is older than in the table already.
			// this is not an error for this application - we want the latest information only in
			// the quake table.
			// http://www.postgresql.org/docs/9.3/static/errcodes-appendix.html
			if err.Code == `23505` {
				txn.Rollback()
				err = nil
			} else {
				txn.Rollback()
				return err
			}
		} else {
			txn.Rollback()
			return err
		}
	} else {
		err = txn.Commit()
		if err != nil {
			return err
		}
	}

	// Quake api
	txn, err = db.Begin()
	if err != nil {
		return err
	}

	_, err = txn.Exec(`DELETE FROM haz.quakeapi WHERE PublicID = $1 AND ModificationTime < $2`, q.PublicID, q.ModificationTime)
	if err != nil {
		txn.Rollback()
		return err
	}

	_, err = txn.Exec(`INSERT INTO haz.quakeapi(`+insert+`) VALUES( `+params+` )`, values...)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// a unique_violation means the new quake info is older than in the table already.
			// this is not an error for this application - we want the latest information only in
			// the quake table.
			// http://www.postgresql.org/docs/9.3/static/errcodes-appendix.html
			if err.Code == `23505` {
				txn.Rollback()
				err = nil
			} else {
				txn.Rollback()
				return err
			}
		} else {
			txn.Rollback()
			return err
		}
	} else {
		err = txn.Commit()
		if err != nil {
			return err
		}
	}

	// Clean out old quakes from quakeapi
	_, err = db.Exec(`DELETE FROM haz.quakeapi WHERE time < now() - interval '365 days' OR status = 'duplicate'`)

	return err
}
