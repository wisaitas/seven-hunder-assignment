package app

import "time"

var Config struct {
	Service struct {
		BodyLimit    int               `env:"BODY_LIMIT" envDefault:"10"`
		Port         int               `env:"PORT" envDefault:"8080"`
		Name         string            `env:"NAME" envDefault:"backend-challenge"`
		MaskMap      map[string]string `env:"MASK_MAP" envDefault:""`
		ReadTimeout  time.Duration     `env:"READ_TIMEOUT" envDefault:"10s"`
		WriteTimeout time.Duration     `env:"WRITE_TIMEOUT" envDefault:"10s"`
	} `envPrefix:"SERVICE_"`
	MongoDB struct {
		Username   string  `env:"USERNAME" envDefault:"root"`
		Password   string  `env:"PASSWORD" envDefault:"password"`
		Host       string  `env:"HOST" envDefault:"localhost"`
		Port       int     `env:"PORT" envDefault:"27017"`
		Database   string  `env:"DATABASE" envDefault:"backend-challenge"`
		AuthSource *string `env:"AUTH_SOURCE" envDefault:"admin"`
	} `envPrefix:"MONGODB_"`
	Redis struct {
		Host     string `env:"HOST" envDefault:"localhost"`
		Port     int    `env:"PORT" envDefault:"6379"`
		Password string `env:"PASSWORD" envDefault:"password"`
		DB       int    `env:"DB" envDefault:"0"`
	} `envPrefix:"REDIS_"`
	JWT struct {
		Secret     string `env:"SECRET" envDefault:"secret-key"`
		AccessTTL  int64  `env:"ACCESS_TTL" envDefault:"15"`
		RefreshTTL int64  `env:"REFRESH_TTL" envDefault:"24"`
	} `envPrefix:"JWT_"`
}
