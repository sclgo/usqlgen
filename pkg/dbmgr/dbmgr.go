package dbmgr

import (
	"database/sql"
	"github.com/samber/lo"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func FindNew(current []string, original []string) []string {
	diff, _ := lo.Difference(current, original)
	return diff
}

func RegisterNewDrivers(existing []string) {
	for _, driver := range FindNew(sql.Drivers(), existing) {
		dburl.Unregister(driver)
		dburl.Register(GetScheme(driver))
		d := drivers.Driver{}
		drivers.Register(driver, d)
	}
}

func GetScheme(driver string) dburl.Scheme {
	return dburl.Scheme{
		Driver:    driver,
		Generator: dburl.GenOpaque,
		Opaque:    true,
	}
}
