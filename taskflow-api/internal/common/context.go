package common

import "context"

func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "userID", userID)
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("userID").(string)
	return userID, ok
}
