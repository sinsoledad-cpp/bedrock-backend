package startup

import "bedrock-backend/pkg/logger"

func InitLogger() logger.Logger {
	return logger.NewNopLogger()
}
