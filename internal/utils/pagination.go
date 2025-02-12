package utils

import (
	"github.com/Quizert/PostCommentService/internal/consts"
)

func ParseLimitOffset(limit *int, offset *int) (int, int) {
	var limitValue, offsetValue int

	if limit == nil || *limit <= 0 {
		limitValue = consts.DefaultLimit
	} else if *limit > consts.MaxLimit {
		limitValue = consts.MaxLimit
	} else {
		limitValue = *limit
	}

	if offset == nil {
		offsetValue = consts.DefaultOffset
	} else {
		offsetValue = *offset
	}
	return limitValue, offsetValue
}
