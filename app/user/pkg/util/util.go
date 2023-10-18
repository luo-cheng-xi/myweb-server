package util

import (
	"github.com/google/wire"
	"myweb/app/user/pkg/mail"
)

var ProviderSet = wire.NewSet(NewJwtUtil, mail.NewEmailUtil)
