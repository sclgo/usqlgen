package gen

import (
	"database/sql"
	"github.com/samber/lo"
	"github.com/xo/dburl"
)

// This file is copied as in the generated usql wrapper.
// It must be self-contained. TODO move to dedicated package so self-containment can be enforced.

func FindNew(current []string, original []string) []string {
	diff, _ := lo.Difference(current, original)
	return diff
}

func RegisterNewDrivers(existing []string) []string {
	newDrivers := FindNew(sql.Drivers(), existing)
	for _, driver := range newDrivers {
		dburl.Unregister(driver)
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
