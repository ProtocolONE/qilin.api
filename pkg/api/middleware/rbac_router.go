package middleware

type RbacRouter interface {
	GetOwner(ctx QilinContext) (string, error)
}
