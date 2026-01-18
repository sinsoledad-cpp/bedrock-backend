package bootstrap

import "bedrock-backend/pkg/validate"

func InitValidate() {
	if err := validate.InitTrans("zh"); err != nil {
		panic(err)
	}
}
