package rbac_echo

type Router interface {
	GetOwner(ctx AppContext) (string, error)
}
