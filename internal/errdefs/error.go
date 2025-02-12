package errdefs

func InternalServerError() error {
	return &AppError{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "Internal Server Error",
	}
}

func CommentTooLongError(maxLength, currentLength int) *AppError {
	return &AppError{
		Code:    "COMMENT_TOO_LONG",
		Message: "Comment exceeds maximum allowed length",
		Extensions: map[string]interface{}{
			"maxLength":     maxLength,
			"currentLength": currentLength,
		},
	}
}

func UserDoesNotExistError(userID int) *AppError {
	return &AppError{
		Code:    "USER_DOES_NOT_EXIST",
		Message: "User does not exist",
		Extensions: map[string]interface{}{
			"userID": userID,
		},
	}
}

func PostDoesNotExistError(userID int) *AppError {
	return &AppError{
		Code:    "POST_DOES_NOT_EXIST",
		Message: "POST does not exist",
		Extensions: map[string]interface{}{
			"postID": userID,
		},
	}
}

func CommentsNotAllowed(postID int) *AppError {
	return &AppError{
		Code:    "COMMENTS_NOT_ALLOWED",
		Message: "comments not allowed",
		Extensions: map[string]interface{}{
			"postID": postID,
		},
	}
}

func NoChanelError(postID int) *AppError {
	return &AppError{
		Code:    "NO_CHANEL_TO_DELETE",
		Message: "no chanel to delere",
		Extensions: map[string]interface{}{
			"postID": postID,
		},
	}
}
