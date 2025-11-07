package initial

import (
	"github.com/7-solutions/backend-challenge/pkg/jwtx"
	"github.com/7-solutions/backend-challenge/pkg/validatorx"
	"go.uber.org/zap"
)

type sdk struct {
	validator validatorx.Validator
	jwt       jwtx.Jwt
	zapLogger *zap.Logger
}

func newSdk() *sdk {
	return &sdk{
		validator: validatorx.NewValidator(),
		jwt:       jwtx.NewJwt(),
		zapLogger: zap.Must(zap.NewDevelopment()),
	}
}
