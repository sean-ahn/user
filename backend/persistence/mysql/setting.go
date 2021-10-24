package mysql

type Setting struct {
	Host              string
	Port              int
	Name              string
	User              string
	Password          string
	MaxIdleConns      int
	MaxOpenConns      int
	ConnMaxLifetimeMs int
}
