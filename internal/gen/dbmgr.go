package gen

import (
	"database/sql"
	"github.com/samber/lo"
	"github.com/xo/dburl"
)

// This file is copied as in the generated usql wrapper.
// It must be self-contained. TODO move to dedicated package so self-containment can be enforced.

// For now, we avoid depending on xo/usql, only on xo/dburl, to keep our dep tree small.
// This may change in the future.

func FindNew(current []string, original []string) []string {
	diff, _ := lo.Difference(current, original)
	return diff
}

// RegisterNewDrivers registers in xo/dburl all database/sql.Drivers()
// that are not present in provided existing list.
func RegisterNewDrivers(existing []string) []string {
	newDrivers := FindNew(sql.Drivers(), existing)
	for _, driver := range newDrivers {
		dburl.Unregister(driver)
		if len(driver) > 2 {
			dburl.Unregister(driver[:2])
		}
		dburl.Register(GetScheme(driver))
	}
	return newDrivers
}

func GetScheme(driver string) dburl.Scheme {
	return dburl.Scheme{
		Driver:    driver,
		Generator: dburl.GenOpaque,
		Opaque:    true,
	}
}
